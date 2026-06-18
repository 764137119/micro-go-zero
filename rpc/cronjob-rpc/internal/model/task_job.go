package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// TaskJob 任务注册表
type TaskJob struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name          string    `gorm:"column:name;type:varchar(128);not null;uniqueIndex"`
	CronExpr      string    `gorm:"column:cron_expr;type:varchar(64);not null"`
	TargetType    int32     `gorm:"column:target_type;type:tinyint;not null;default:0"` // 0: gRPC, 1: HTTP
	Target        string    `gorm:"column:target;type:varchar(256);not null"`
	RequestBody   string    `gorm:"column:request_body;type:text"`
	MaxRetries    int32     `gorm:"column:max_retries;type:int;not null;default:3"`
	RetryInterval int32     `gorm:"column:retry_interval;type:int;not null;default:30"` // 重试间隔(秒)
	Description   string    `gorm:"column:description;type:varchar(512)"`
	Enabled       bool      `gorm:"column:enabled;type:tinyint(1);not null;default:1"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (TaskJob) TableName() string {
	return "task_jobs"
}

// TaskExecution 执行历史
type TaskExecution struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TaskName    string    `gorm:"column:task_name;type:varchar(128);not null;index"`
	ScheduledAt time.Time `gorm:"column:scheduled_at;type:datetime;not null;index"`
	StartedAt   time.Time `gorm:"column:started_at;type:datetime"`
	FinishedAt  time.Time `gorm:"column:finished_at;type:datetime"`
	Status      string    `gorm:"column:status;type:varchar(16);not null;default:'pending';index"` // pending/running/success/failed/retrying
	RetryCount  int32     `gorm:"column:retry_count;type:int;not null;default:0"`
	MaxRetries  int32     `gorm:"column:max_retries;type:int;not null;default:3"`
	Result      string    `gorm:"column:result;type:text"`
	TraceID     string    `gorm:"column:trace_id;type:varchar(64)"`
	ExecNode    string    `gorm:"column:exec_node;type:varchar(128)"`
}

func (TaskExecution) TableName() string {
	return "task_executions"
}

// TaskJobRepo 任务注册表仓储
type TaskJobRepo struct {
	db *gorm.DB
}

func NewTaskJobRepo(db *gorm.DB) *TaskJobRepo {
	return &TaskJobRepo{db: db}
}

func (r *TaskJobRepo) Create(ctx context.Context, job *TaskJob) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *TaskJobRepo) Update(ctx context.Context, job *TaskJob) error {
	return r.db.WithContext(ctx).Save(job).Error
}

func (r *TaskJobRepo) DeleteByName(ctx context.Context, name string) error {
	return r.db.WithContext(ctx).Where("name = ?", name).Delete(&TaskJob{}).Error
}

func (r *TaskJobRepo) FindByName(ctx context.Context, name string) (*TaskJob, error) {
	var job TaskJob
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&job).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *TaskJobRepo) ListAllEnabled(ctx context.Context) ([]TaskJob, error) {
	var jobs []TaskJob
	err := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&jobs).Error
	return jobs, err
}

func (r *TaskJobRepo) List(ctx context.Context, page, pageSize int32) ([]TaskJob, int64, error) {
	var jobs []TaskJob
	var total int64

	query := r.db.WithContext(ctx).Model(&TaskJob{})
	query.Count(&total)

	err := query.Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Order("id DESC").Find(&jobs).Error
	return jobs, total, err
}

// TaskExecutionRepo 执行历史仓储
type TaskExecutionRepo struct {
	db *gorm.DB
}

func NewTaskExecutionRepo(db *gorm.DB) *TaskExecutionRepo {
	return &TaskExecutionRepo{db: db}
}

func (r *TaskExecutionRepo) Create(ctx context.Context, exec *TaskExecution) error {
	return r.db.WithContext(ctx).Create(exec).Error
}

func (r *TaskExecutionRepo) Update(ctx context.Context, exec *TaskExecution) error {
	return r.db.WithContext(ctx).Save(exec).Error
}

func (r *TaskExecutionRepo) FindByID(ctx context.Context, id int64) (*TaskExecution, error) {
	var exec TaskExecution
	err := r.db.WithContext(ctx).First(&exec, id).Error
	return &exec, err
}

func (r *TaskExecutionRepo) List(ctx context.Context, taskName string, status string, page, pageSize int32) ([]TaskExecution, int64, error) {
	var records []TaskExecution
	var total int64

	query := r.db.WithContext(ctx).Model(&TaskExecution{})
	if taskName != "" {
		query = query.Where("task_name = ?", taskName)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	err := query.Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Order("id DESC").Find(&records).Error
	return records, total, err
}

func (r *TaskExecutionRepo) FindPendingRetry(ctx context.Context, maxRetryCount int32) ([]TaskExecution, error) {
	var records []TaskExecution
	err := r.db.WithContext(ctx).
		Where("status IN ?", []string{"failed", "retrying"}).
		Where("retry_count < max_retries").
		Where("retry_count < ?", maxRetryCount).
		Find(&records).Error
	return records, err
}

// GetTaskStats 获取任务统计
func (r *TaskExecutionRepo) GetTaskStats(ctx context.Context, taskName string) (totalExecutions, successCount, failedCount, runningCount int64, lastExecutedAt *time.Time, lastStatus string, err error) {
	err = r.db.WithContext(ctx).Model(&TaskExecution{}).
		Select("COUNT(*) as total").
		Where("task_name = ?", taskName).
		Scan(&totalExecutions).Error
	if err != nil {
		return
	}

	r.db.WithContext(ctx).Model(&TaskExecution{}).
		Where("task_name = ? AND status = ?", taskName, "success").
		Count(&successCount)

	r.db.WithContext(ctx).Model(&TaskExecution{}).
		Where("task_name = ? AND status = ?", taskName, "failed").
		Count(&failedCount)

	r.db.WithContext(ctx).Model(&TaskExecution{}).
		Where("task_name = ? AND status = ?", taskName, "running").
		Count(&runningCount)

	var last TaskExecution
	if err := r.db.WithContext(ctx).Where("task_name = ?", taskName).
		Order("scheduled_at DESC").First(&last).Error; err == nil {
		lastExecutedAt = &last.ScheduledAt
		lastStatus = last.Status
	}

	return
}

// AutoMigrate 自动迁移表结构
func AutoMigrate(db *gorm.DB) {
	if err := db.AutoMigrate(&TaskJob{}, &TaskExecution{}); err != nil {
		panic("failed to auto migrate task tables: " + err.Error())
	}
}
