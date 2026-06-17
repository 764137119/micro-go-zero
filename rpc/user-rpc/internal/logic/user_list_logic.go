package logic

import (
	"context"

	"user-rpc/internal/svc"
	"user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserListLogic {
	return &UserListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户列表
func (l *UserListLogic) UserList(in *user.UserListReq) (*user.UserListResp, error) {
	// todo: add your logic here and delete this line

	return &user.UserListResp{}, nil
}
