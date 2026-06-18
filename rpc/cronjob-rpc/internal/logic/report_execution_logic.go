package logic

import (
	"context"

	"cronjob-rpc/cronjob"
	"cronjob-rpc/internal/model"
	"cronjob-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReportExecutionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReportExecutionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportExecutionLogic {
	return &ReportExecutionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 业务方回调
func (l *ReportExecutionLogic) ReportExecution(in *cronjob.ReportExecutionReq) (*cronjob.ReportExecutionResp, error) {
	// 将 proto 状态转为 model 状态
	status := ""
	switch in.Status {
	case cronjob.ExecStatus_SUCCESS:
		status = model.ExecStatusSuccess
	case cronjob.ExecStatus_FAILED:
		status = model.ExecStatusFailed
	default:
		logx.Errorf("Invalid report status: %v for execution %d", in.Status, in.ExecutionId)
		return &cronjob.ReportExecutionResp{Ok: false}, nil
	}

	if err := l.svcCtx.Executor.ReportExecution(l.ctx, in.ExecutionId, status, in.Result); err != nil {
		logx.Errorf("Failed to report execution %d: %v", in.ExecutionId, err)
		return &cronjob.ReportExecutionResp{Ok: false}, nil
	}

	logx.Infof("Execution %d reported: status=%s", in.ExecutionId, status)
	return &cronjob.ReportExecutionResp{Ok: true}, nil
}
