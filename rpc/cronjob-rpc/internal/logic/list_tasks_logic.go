package logic

import (
	"context"
	"cronjob-rpc/cronjob"
	"cronjob-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTasksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTasksLogic {
	return &ListTasksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListTasksLogic) ListTasks(in *cronjob.ListTasksReq) (*cronjob.ListTasksResp, error) {
	page := in.Page
	if page < 1 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	jobs, total, err := l.svcCtx.TaskJobRepo.List(l.ctx, page, pageSize)
	if err != nil {
		logx.Errorf("Failed to list tasks: %v", err)
		return &cronjob.ListTasksResp{}, nil
	}

	var tasks []*cronjob.TaskInfo
	for _, job := range jobs {
		status := cronjob.TaskStatus_TASK_DISABLED
		if job.Enabled {
			status = cronjob.TaskStatus_TASK_ENABLED
		}

		task := &cronjob.TaskInfo{
			Name:        job.Name,
			CronExpr:    job.CronExpr,
			TargetType:  cronjob.TargetType(job.TargetType),
			Target:      job.Target,
			RequestBody: job.RequestBody,
			RetryPolicy: &cronjob.RetryPolicy{
				MaxRetries:  job.MaxRetries,
				IntervalSec: job.RetryInterval,
			},
			Description: job.Description,
			Status:      status,
			CreatedAt:   job.CreatedAt.Unix(),
			UpdatedAt:   job.UpdatedAt.Unix(),
		}
		tasks = append(tasks, task)
	}

	if tasks == nil {
		tasks = []*cronjob.TaskInfo{}
	}

	return &cronjob.ListTasksResp{
		Tasks: tasks,
		Total: int32(total),
	}, nil
}
