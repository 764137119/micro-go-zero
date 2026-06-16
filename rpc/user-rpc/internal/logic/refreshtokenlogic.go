package logic

import (
	"context"

	"user-rpc/internal/svc"
	"user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 刷新token
func (l *RefreshTokenLogic) RefreshToken(in *user.RefreshTokenReq) (*user.RefreshTokenResp, error) {
	// todo: add your logic here and delete this line

	return &user.RefreshTokenResp{}, nil
}
