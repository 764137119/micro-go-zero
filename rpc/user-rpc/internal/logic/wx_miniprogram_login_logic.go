package logic

import (
	"context"

	"user-rpc/internal/svc"
	"user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type WxMiniProgramLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewWxMiniProgramLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WxMiniProgramLoginLogic {
	return &WxMiniProgramLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 微信小程序登录
func (l *WxMiniProgramLoginLogic) WxMiniProgramLogin(in *user.WxMiniProgramLoginReq) (*user.WxMiniProgramLoginResp, error) {
	// todo: add your logic here and delete this line

	return &user.WxMiniProgramLoginResp{}, nil
}
