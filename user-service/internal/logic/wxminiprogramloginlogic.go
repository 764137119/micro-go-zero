// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"user-service/internal/svc"
	"user-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WxMiniProgramLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWxMiniProgramLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WxMiniProgramLoginLogic {
	return &WxMiniProgramLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WxMiniProgramLoginLogic) WxMiniProgramLogin(req *types.WxMiniProgramLoginReq) (resp *types.WxMiniProgramLoginResp, err error) {
	// todo: add your logic here and delete this line

	return
}
