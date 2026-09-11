package service

import (
	"backend/model"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/robfig/cron/v3"
)

// ============ 任务注册表 ============
// 后台可调度的任务必须先注册到 registry, 方法名是注册 key 而非反射任意 Go 方法

type taskEntry struct {
	meta    model.TaskMeta
	handler func(params json.RawMessage) error
}

// RegisterTask 注册可调度任务
func (js *JobScheduler) RegisterTask(name, description string, paramsExample json.RawMessage, handler func(params json.RawMessage) error) {
	js.mu.Lock()
	defer js.mu.Unlock()
	js.tasks[name] = taskEntry{
		meta:    model.TaskMeta{Name: name, Description: description, ParamsExample: paramsExample},
		handler: handler,
	}
	log.Printf("[JobScheduler] 注册任务: %s (%s)", name, description)
}

// ListRegisteredTasks 列出所有已注册任务(供后台下拉选择)
func (js *JobScheduler) ListRegisteredTasks() []model.TaskMeta {
	js.mu.Lock()
	defer js.mu.Unlock()
	list := make([]model.TaskMeta, 0, len(js.tasks))
	for _, entry := range js.tasks {
		list = append(list, entry.meta)
	}
	return list
}

func (js *JobScheduler) getTask(taskName string) (func(json.RawMessage) error, bool) {
	js.mu.Lock()
	defer js.mu.Unlock()
	entry, ok := js.tasks[taskName]
	return entry.handler, ok
}

// GetTask 查询任务是否已注册
func (js *JobScheduler) GetTask(taskName string) (model.TaskMeta, bool) {
	js.mu.Lock()
	defer js.mu.Unlock()
	entry, ok := js.tasks[taskName]
	return entry.meta, ok
}

// ============ cron 定义调度 ============

// LoadCronDefinitions 启动时加载定时任务定义并注册到调度器
func (js *JobScheduler) LoadCronDefinitions() error {
	// 清理上次进程中断遗留的 running 记录
	if err := DB.Model(&model.JobRun{}).Where("status = ?", model.JobRunStatusRunning).Updates(map[string]interface{}{
		"status":      model.JobRunStatusFailed,
		"error":       "服务重启导致任务中断",
		"finished_at": time.Now(),
	}).Error; err != nil {
		log.Printf("[JobScheduler] 清理中断的定时任务记录失败: %v", err)
	}

	var defs []model.JobDefinition
	if err := DB.Where("enabled = ?", true).Find(&defs).Error; err != nil {
		return err
	}
	for i := range defs {
		js.ScheduleDefinition(&defs[i])
	}
	log.Printf("[JobScheduler] 已加载 %d 个启用的定时任务", len(defs))
	return nil
}

// ValidateCronExpr 校验 cron 表达式(5段标准格式), 并返回下一次执行时间
func ValidateCronExpr(expr string) (time.Time, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := parser.Parse(expr)
	if err != nil {
		return time.Time{}, fmt.Errorf("cron 表达式无效: %v", err)
	}
	return schedule.Next(time.Now()), nil
}

// ScheduleDefinition 注册/更新任务定义的调度 (新增或编辑后调用)
func (js *JobScheduler) ScheduleDefinition(def *model.JobDefinition) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	key := fmt.Sprintf("cron-def:%d", def.ID)
	if existing, ok := js.cronJobs[key]; ok {
		js.scheduler.RemoveJob(existing.ID())
		delete(js.cronJobs, key)
	}

	if !def.Enabled {
		return nil
	}

	job, err := js.scheduler.NewJob(
		gocron.CronJob(def.CronExpr, false),
		gocron.NewTask(js.cronTick, def.ID),
		gocron.WithName(key),
		gocron.WithTags(key),
	)
	if err != nil {
		return fmt.Errorf("注册 cron 任务失败: %v", err)
	}

	js.cronJobs[key] = job
	log.Printf("[JobScheduler] 已注册定时任务: %s (%s) [%s]", def.Name, def.TaskName, def.CronExpr)
	return nil
}

// UnscheduleDefinition 移除任务定义的调度 (停用或删除后调用)
func (js *JobScheduler) UnscheduleDefinition(defID uint) {
	js.mu.Lock()
	defer js.mu.Unlock()

	key := fmt.Sprintf("cron-def:%d", defID)
	if existing, ok := js.cronJobs[key]; ok {
		js.scheduler.RemoveJob(existing.ID())
		delete(js.cronJobs, key)
	}
}

// NextRunAt 查询定义的下一次执行时间(未启用或未注册返回零值)
func (js *JobScheduler) NextRunAt(defID uint) *time.Time {
	js.mu.Lock()
	defer js.mu.Unlock()

	job, ok := js.cronJobs[fmt.Sprintf("cron-def:%d", defID)]
	if !ok {
		return nil
	}
	next, err := job.NextRun()
	if err != nil || next.IsZero() {
		return nil
	}
	return &next
}

// cronTick 调度器到点触发
func (js *JobScheduler) cronTick(defID uint) {
	js.runDefinition(defID, model.JobTriggerScheduler, 1)
}

// TriggerDefinition 手动立即执行 (异步, 返回是否成功启动)
func (js *JobScheduler) TriggerDefinition(defID uint) bool {
	var def model.JobDefinition
	if err := DB.First(&def, defID).Error; err != nil {
		log.Printf("[JobScheduler] 手动触发失败: 定义 %d 不存在", defID)
		return false
	}
	go js.runDefinition(defID, model.JobTriggerManual, 1)
	return true
}

// ============ 执行引擎 ============
// 手动触发与调度触发走同一入口, 区分 trigger_type

// cronRunning 正在执行的定义(单实例内存态, 防止 cron/手动重入)
var cronRunning = struct {
	sync.Mutex
	ids map[uint]bool
}{ids: make(map[uint]bool)}

func acquireDefinition(defID uint) bool {
	cronRunning.Lock()
	defer cronRunning.Unlock()
	if cronRunning.ids[defID] {
		return false
	}
	cronRunning.ids[defID] = true
	return true
}

func releaseDefinition(defID uint) {
	cronRunning.Lock()
	defer cronRunning.Unlock()
	delete(cronRunning.ids, defID)
}

func (js *JobScheduler) runDefinition(defID uint, triggerType string, attempt int) {
	if !acquireDefinition(defID) {
		if triggerType == model.JobTriggerScheduler {
			// 上一轮还在执行, 记录一次跳过
			now := time.Now()
			DB.Create(&model.JobRun{
				Definition:  defID,
				Attempt:     attempt,
				Status:      model.JobRunStatusSkipped,
				TriggerType: triggerType,
				StartedAt:   &now,
				FinishedAt:  &now,
				Error:       "上一轮执行尚未结束, 本次跳过",
			})
			log.Printf("[JobScheduler] 定时任务 %d 上一轮未结束, 跳过本轮", defID)
		}
		return
	}
	defer releaseDefinition(defID)

	var def model.JobDefinition
	if err := DB.First(&def, defID).Error; err != nil {
		log.Printf("[JobScheduler] 定时任务 %d 定义不存在", defID)
		return
	}

	handler, ok := js.getTask(def.TaskName)
	now := time.Now()
	run := model.JobRun{
		Definition:  def.ID,
		Attempt:     attempt,
		Status:      model.JobRunStatusRunning,
		TriggerType: triggerType,
		StartedAt:   &now,
	}
	if !ok {
		finishAt := time.Now()
		run.Status = model.JobRunStatusFailed
		run.FinishedAt = &finishAt
		run.Error = fmt.Sprintf("任务方法未注册: %s", def.TaskName)
		DB.Create(&run)
		log.Printf("[JobScheduler] 定时任务 %s 执行失败: %s", def.Name, run.Error)
		return
	}
	DB.Create(&run)

	log.Printf("[JobScheduler] 执行定时任务: %s (%s) 第%d次尝试", def.Name, def.TaskName, attempt)

	// panic 保护, 避免协程崩溃拖垮进程
	err := func() (cbErr error) {
		defer func() {
			if r := recover(); r != nil {
				cbErr = fmt.Errorf("任务执行 panic: %v", r)
			}
		}()
		return handler(def.Params)
	}()

	finishAt := time.Now()
	run.FinishedAt = &finishAt
	if err != nil {
		run.Status = model.JobRunStatusFailed
		run.Error = err.Error()
		DB.Model(&model.JobRun{}).Where("id = ?", run.ID).Updates(map[string]interface{}{
			"status":      run.Status,
			"finished_at": finishAt,
			"error":       run.Error,
		})
		log.Printf("[JobScheduler] 定时任务 %s 执行失败: %v", def.Name, err)

		// 失败重试: 延迟 attempt 分钟后单次调度
		if attempt <= def.MaxRetries {
			retryAt := time.Now().Add(time.Duration(attempt) * time.Minute)
			if _, err := js.scheduler.NewJob(
				gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(retryAt)),
				gocron.NewTask(js.runDefinitionRetry, def.ID, triggerType, attempt+1),
				gocron.WithName(fmt.Sprintf("cron-def-retry:%d:%d", def.ID, attempt)),
				gocron.WithTags(fmt.Sprintf("cron-def-retry:%d", def.ID)),
			); err != nil {
				log.Printf("[JobScheduler] 定时任务 %s 重试调度失败: %v", def.Name, err)
			} else {
				log.Printf("[JobScheduler] 定时任务 %s 将于 %v 重试(第%d次)", def.Name, retryAt, attempt+1)
			}
		}
		return
	}

	run.Status = model.JobRunStatusSuccess
	DB.Model(&model.JobRun{}).Where("id = ?", run.ID).Updates(map[string]interface{}{
		"status":      run.Status,
		"finished_at": finishAt,
	})
	log.Printf("[JobScheduler] 定时任务 %s 执行成功", def.Name)
}

// runDefinitionRetry 重试入口
func (js *JobScheduler) runDefinitionRetry(defID uint, triggerType string, attempt int) {
	js.runDefinition(defID, triggerType, attempt)
}
