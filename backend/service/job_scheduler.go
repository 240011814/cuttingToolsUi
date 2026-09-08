package service

import (
	"backend/model"
	"fmt"
	"log"
	"sync"
	"time"

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
	var jobs []model.Job
	if err := DB.Where("status = ?", model.JobStatusPending).Find(&jobs).Error; err != nil {
		return err
	}

	for _, job := range jobs {
		if job.ScheduledAt.Before(time.Now()) {
			// 过去的任务：立即执行
			go js.executeJob(job)
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
	if !job.ScheduledAt.Before(time.Now()) {
		js.scheduleJob(*job)
	} else {
		// 过去的任务立即执行
		go js.executeJob(*job)
	}

	return nil
}

// UnscheduleJob 从调度器移除任务（不改变数据库状态）
func (js *JobScheduler) UnscheduleJob(jobID uint, jobType string) {
	js.mu.Lock()
	defer js.mu.Unlock()

	jobKey := js.getJobKey(jobID, jobType)
	if job, ok := js.jobMap[jobKey]; ok {
		js.scheduler.RemoveJob(job.ID())
		delete(js.jobMap, jobKey)
	}
}

func (js *JobScheduler) scheduleJob(job model.Job) {
	js.mu.Lock()
	defer js.mu.Unlock()

	jobKey := js.getJobKey(job.JobID, job.JobType)

	// 如果已有任务，先移除
	if existingJob, ok := js.jobMap[jobKey]; ok {
		js.scheduler.RemoveJob(existingJob.ID())
		delete(js.jobMap, jobKey)
	}

	// 计算实际调度时间（提醒时间 - 提前分钟数）
	scheduledAt := job.ScheduledAt
	if job.AdvanceMinutes > 0 {
		scheduledAt = job.ScheduledAt.Add(-time.Duration(job.AdvanceMinutes) * time.Minute)
	}

	// 如果已过期，不调度
	if scheduledAt.Before(time.Now()) {
		log.Printf("[JobScheduler] 任务 %s:%d 时间已过期，跳过调度", job.JobType, job.JobID)
		return
	}

	gocronJob, err := js.scheduler.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(scheduledAt),
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
	delay := time.Until(scheduledAt)
	log.Printf("[JobScheduler] 已注册任务: %s:%d, %v 后执行 (提前通知: %d分钟)", job.JobType, job.JobID, delay, job.AdvanceMinutes)
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

	// 执行回调
	if err := callback(job); err != nil {
		js.handleJobError(job, err)
		return
	}

	// 执行成功
	DB.Model(&model.Job{}).Where("id = ?", job.ID).Update("status", model.JobStatusCompleted)
	log.Printf("[JobScheduler] 任务执行成功: %s:%d", job.JobType, job.JobID)

	// 从内存中移除
	js.mu.Lock()
	jobKey := js.getJobKey(job.JobID, job.JobType)
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
		// 更新重试次数和错误信息
		DB.Model(&model.Job{}).Where("id = ?", job.ID).Updates(map[string]interface{}{
			"retry_count": job.RetryCount + 1,
			"last_error":  err.Error(),
			"status":      model.JobStatusPending,
		})

		// 重新调度（延迟重试）
		job.RetryCount++
		job.Status = model.JobStatusPending
		job.ScheduledAt = time.Now().Add(time.Duration(job.RetryCount) * time.Minute)
		js.scheduleJob(job)
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

	newJob := model.Job{
		JobType:        job.JobType,
		JobID:          job.JobID,
		UserID:         job.UserID,
		ScheduledAt:    nextAt,
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
		return at.AddDate(0, interval, 0)
	case "yearly":
		return at.AddDate(interval, 0, 0)
	default:
		return time.Time{}
	}
}