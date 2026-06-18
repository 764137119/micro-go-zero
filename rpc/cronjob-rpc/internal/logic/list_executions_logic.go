package logic

import (
	"context"
	"cronjob-rpc/cronjob"
	"cronjob-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListExecutionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListExecutionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListExecutionsLogic {
	return &ListExecutionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListExecutionsLogic) ListExecutions(in *cronjob.ListExecutionsReq) (*cronjob.ListExecutionsResp, error) {
	page := in.Page
	if page < 1 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	status := ""
	if in.Status != cronjob.ExecStatus_PENDING {
		status = in.Status.String()
	}

	records, total, err := l.svcCtx.TaskExecRepo.List(l.ctx, in.TaskName, status, page, pageSize)
	if err != nil {
		logx.Errorf("Failed to list executions: %v", err)
		return &cronjob.ListExecutionsResp{}, nil
	}

	var execRecords []*cronjob.ExecutionRecord
	for _, r := range records {
		execStatus := cronjob.ExecStatus_PENDING
		switch r.Status {
		case "running":
			execStatus = cronjob.ExecStatus_RUNNING
		case "success":
			execStatus = cronjob.ExecStatus_SUCCESS
		case "failed":
			execStatus = cronjob.ExecStatus_FAILED
		case "retrying":
			execStatus = cronjob.ExecStatus_RETRYING
		}

		record := &cronjob.ExecutionRecord{
			Id:          r.ID,
			TaskName:    r.TaskName,
			ScheduledAt: r.ScheduledAt.Unix(),
			StartedAt:   r.StartedAt.Unix(),
			FinishedAt:  r.FinishedAt.Unix(),
			Status:      execStatus,
			RetryCount:  r.RetryCount,
			MaxRetries:  r.MaxRetries,
			Result:      r.Result,
			TraceId:     r.TraceID,
			ExecNode:    r.ExecNode,
		}
		execRecords = append(execRecords, record)
	}

	if execRecords == nil {
		execRecords = []*cronjob.ExecutionRecord{}
	}

	return &cronjob.ListExecutionsResp{
		Records: execRecords,
		Total:   int32(total),
	}, nil
}
