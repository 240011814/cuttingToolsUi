package service

import (
	"backend/model"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/go-co-op/gocron/v2"
)

type JobScheduler struct {
	scheduler gocron.Scheduler
	mu        sync.Mutex
	jobMap    map[string]gocron.Job // jobKey -> job
	callbacks map[string]model.JobCallback // jobType -> callback
}

func NewJobScheduler() (*JobScheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	js := &JobScheduler{
		scheduler: s,
		jobMap:    make(map[string]gocron.Job),
		callbacks: make(map[string]model.JobCallback),
	}

	return js, nil
}

func (js *JobScheduler) Start() {
	js.scheduler.Start()
	log.Println("[JobScheduler] 已启动")
}

func (js *JobScheduler) GetScheduler() gocron.Scheduler {
	return js.scheduler
}

func (js *JobScheduler) Shutdown() error {
	return js.scheduler.Shutdown()
}

// RegisterCallback 注册任务回调函数
func (js *JobScheduler) RegisterCallback(jobType string, callback model.JobCallback) {
	js.mu.Lock()
	defer js.mu.Unlock()
	js.callbacks[jobType] = callback
	log.Printf("[JobScheduler] 注册回调: %s", jobType)
}

// LoadAll 从数据库加载所有待执行的任务
func (js *JobScheduler) LoadAll() error {
	// 清理上次进程中断遗留的 running 任务，避免永久卡死
	if err := DB.Model(&model.Job{}).Where("status = ?", model.JobStatusRunning).Updates(map[string]interface{}{
		"status":     model.JobStatusFailed,
		"last_error": "服务重启导致任务中断",
	}).Error; err != nil {
		log.Printf("[JobScheduler] 清理中断任务失败: %v", err)
	}

	var jobs []model.Job
	if err := DB.Where("status = ?", model.JobStatusPending).Find(&jobs).Error; err != nil {
		return err
	}

	for _, job := range jobs {
		if job.ScheduledAt.Before(time.Now()) {
			// 过期任务不补发
			log.Printf("[JobScheduler] 任务 %s:%d 已过期，跳过执行", job.JobType, job.JobID)
			js.skipExpiredJob(job)
			continue
		}
		js.scheduleJob(job)
	}

	log.Printf("[JobScheduler] 已加载 %d 个待执行任务", len(jobs))
	return nil
}

// ScheduleJob 创建新任务
func (js *JobScheduler) ScheduleJob(job *model.Job) error {
	// 保存到数据库
	if err := DB.Create(job).Error; err != nil {
		return err
	}

	// 注册到调度器
	if job.ScheduledAt.Before(time.Now()) {
		js.skipExpiredJob(*job)
	} else {
		js.scheduleJob(*job)
	}

	return nil
}

// skipExpiredJob 过期任务不补发：重复任务基于原计划时间创建下一次执行，非重复任务标记为失败
func (js *JobScheduler) skipExpiredJob(job model.Job) {
	status := model.JobStatusFailed
	lastError := "任务已过期，跳过执行"
	if job.RepeatType != "" && job.RepeatType != "none" {
		js.createNextRepeat(job)
		status = model.JobStatusCompleted
		lastError = "任务已过期，跳过本次执行"
	}

	DB.Model(&model.Job{}).Where("id = ?", job.ID).Updates(map[string]interface{}{
		"status":     status,
		"last_error": lastError,
	})
}

// UnscheduleJob 从调度器移除任务（不改变数据库状态），jobID 为 jobs 表行 ID
func (js *JobScheduler) UnscheduleJob(rowID uint, jobType string) {
	js.mu.Lock()
	defer js.mu.Unlock()

	jobKey := js.getJobKey(rowID, jobType)
	if job, ok := js.jobMap[jobKey]; ok {
		js.scheduler.RemoveJob(job.ID())
		delete(js.jobMap, jobKey)
	}
}

func (js *JobScheduler) scheduleJob(job model.Job) {
	// 计算实际调度时间（提醒时间 - 提前分钟数）
	scheduledAt := job.ScheduledAt
	if job.AdvanceMinutes > 0 {
		scheduledAt = job.ScheduledAt.Add(-time.Duration(job.AdvanceMinutes) * time.Minute)
	}
	js.scheduleJobAt(job, scheduledAt)
}

// scheduleJobAt 按指定时间调度任务（重试场景使用，不改变任务的 ScheduledAt，避免影响重复周期计算）
func (js *JobScheduler) scheduleJobAt(job model.Job, at time.Time) {
	js.mu.Lock()
	defer js.mu.Unlock()

	// 按行 ID 键控：重复任务的每一行是独立的执行实例，删除/更新该行才能正确取消调度
	jobKey := js.getJobKey(job.ID, job.JobType)

	// 如果已有任务，先移除
	if existingJob, ok := js.jobMap[jobKey]; ok {
		js.scheduler.RemoveJob(existingJob.ID())
		delete(js.jobMap, jobKey)
	}

	// 如果已过期，不调度
	if at.Before(time.Now()) {
		log.Printf("[JobScheduler] 任务 %s:%d 时间已过期，跳过调度", job.JobType, job.JobID)
		return
	}

	gocronJob, err := js.scheduler.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(at),
		),
		gocron.NewTask(js.executeJob, job),
		gocron.WithName(fmt.Sprintf("%s:%d", job.JobType, job.JobID)),
		gocron.WithTags(job.JobType, fmt.Sprintf("%s:%d", job.JobType, job.JobID)),
	)
	if err != nil {
		log.Printf("[JobScheduler] 注册任务失败 (%s:%d): %v", job.JobType, job.JobID, err)
		return
	}

	js.jobMap[jobKey] = gocronJob
	log.Printf("[JobScheduler] 已注册任务: %s:%d, %v 后执行 (提前通知: %d分钟)", job.JobType, job.JobID, time.Until(at), job.AdvanceMinutes)
}

func (js *JobScheduler) executeJob(job model.Job) {
	log.Printf("[JobScheduler] 执行任务: %s:%d", job.JobType, job.JobID)

	// 更新状态为运行中
	DB.Model(&model.Job{}).Where("id = ?", job.ID).Update("status", model.JobStatusRunning)

	// 查找回调函数
	js.mu.Lock()
	callback, exists := js.callbacks[job.JobType]
	js.mu.Unlock()

	if !exists {
		log.Printf("[JobScheduler] 未找到回调函数: %s", job.JobType)
		js.handleJobError(job, fmt.Errorf("未找到回调函数: %s", job.JobType))
		return
	}

	// 执行回调（panic 保护，避免协程崩溃拖垮整个进程）
	err := func() (cbErr error) {
		defer func() {
			if r := recover(); r != nil {
				cbErr = fmt.Errorf("任务执行 panic: %v", r)
			}
		}()
		return callback(job)
	}()
	if err != nil {
		js.handleJobError(job, err)
		return
	}

	// 执行成功
	DB.Model(&model.Job{}).Where("id = ?", job.ID).Update("status", model.JobStatusCompleted)
	log.Printf("[JobScheduler] 任务执行成功: %s:%d", job.JobType, job.JobID)

	// 从内存中移除
	js.mu.Lock()
	jobKey := js.getJobKey(job.ID, job.JobType)
	delete(js.jobMap, jobKey)
	js.mu.Unlock()

	// 如果有重复类型，创建下一次任务
	if job.RepeatType != "none" {
		js.createNextRepeat(job)
	}
}

func (js *JobScheduler) handleJobError(job model.Job, err error) {
	log.Printf("[JobScheduler] 任务执行失败: %s:%d, 错误: %v", job.JobType, job.JobID, err)

	// 检查是否需要重试
	if job.RetryCount < job.MaxRetries {
		// 更新重试次数和错误信息（retry_count 原子递增，避免并发回退）
		DB.Model(&model.Job{}).Where("id = ?", job.ID).Updates(map[string]interface{}{
			"retry_count": gorm.Expr("retry_count + 1"),
			"last_error":  err.Error(),
			"status":      model.JobStatusPending,
		})

		// 重新调度（延迟重试），不修改 ScheduledAt，保证重复周期基于原计划时间计算
		job.RetryCount++
		job.Status = model.JobStatusPending
		retryAt := time.Now().Add(time.Duration(job.RetryCount) * time.Minute)
		js.scheduleJobAt(job, retryAt)
	} else {
		// 超过最大重试次数，标记为失败
		DB.Model(&model.Job{}).Where("id = ?", job.ID).Updates(map[string]interface{}{
			"status":     model.JobStatusFailed,
			"last_error": err.Error(),
		})
	}
}

func (js *JobScheduler) createNextRepeat(job model.Job) {
	if job.RepeatType == "none" {
		return
	}

	nextAt := calcNextTime(job.ScheduledAt, job.RepeatType, job.RepeatInterval)
	maxIter := 100
	for i := 0; nextAt.Before(time.Now()) && i < maxIter; i++ {
		nextAt = calcNextTime(nextAt, job.RepeatType, job.RepeatInterval)
	}

	if nextAt.IsZero() || nextAt.Before(time.Now()) {
		return
	}

	if job.RepeatEndAt != nil && nextAt.After(*job.RepeatEndAt) {
		return
	}

	// 同链已有未来的待执行任务时不再创建，防止重复生成
	var pendingCount int64
	DB.Model(&model.Job{}).Where("job_type = ? AND job_id = ? AND status = ? AND scheduled_at > ? AND id <> ?",
		job.JobType, job.JobID, model.JobStatusPending, time.Now(), job.ID).Count(&pendingCount)
	if pendingCount > 0 {
		return
	}

	newJob := model.Job{
		JobType:        job.JobType,
		JobID:          job.JobID,
		UserID:         job.UserID,
		ScheduledAt:    nextAt,
		AdvanceMinutes: job.AdvanceMinutes,
		Status:         model.JobStatusPending,
		MaxRetries:     3,
		Params:         job.Params,
		RepeatType:     job.RepeatType,
		RepeatInterval: job.RepeatInterval,
		RepeatEndAt:    job.RepeatEndAt,
	}

	if err := DB.Create(&newJob).Error; err != nil {
		log.Printf("[JobScheduler] 创建下次任务失败: %v", err)
		return
	}

	js.scheduleJob(newJob)
	log.Printf("[JobScheduler] 已创建下次重复任务: %s:%d, 执行时间: %v", job.JobType, job.JobID, nextAt)
}

func (js *JobScheduler) getJobKey(jobID uint, jobType string) string {
	return fmt.Sprintf("%s:%d", jobType, jobID)
}

func calcNextTime(at time.Time, repeatType string, interval int) time.Time {
	switch repeatType {
	case "daily":
		return at.AddDate(0, 0, interval)
	case "weekly":
		return at.AddDate(0, 0, 7*interval)
	case "monthly":
		return addMonthsSameDay(at, interval)
	case "yearly":
		return addMonthsSameDay(at, 12*interval)
	default:
		return time.Time{}
	}
}

// addMonthsSameDay 按月推进并保持"日"不变；目标月份没有该日（如 1月31日、2月29日）时跳到下一个月的同一天
func addMonthsSameDay(at time.Time, months int) time.Time {
	next := at.AddDate(0, months, 0)
	// AddDate 会把溢出的日归一化到下个月（如 1月31日+1月=3月2日），此处检测并跳到下一个存在该日的月份
	for i := 0; next.Day() != at.Day() && i < 12; i++ {
		months++
		next = at.AddDate(0, months, 0)
	}
	return next
}