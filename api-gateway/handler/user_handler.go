package handler

import (
	"context"

	"api-gateway/svc"
	"api-gateway/types"
	userpb "user-rpc/user"

	"github.com/gin-gonic/gin"
)

// Login 用户登录
func Login(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.LoginReq) (*userpb.LoginResp, error) {
			return svcCtx.UserRpc.Login(ctx, req.ToRPC())
		},
	)
}

// Logout 用户登出
func Logout(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.LogoutReq) (*userpb.LogoutResp, error) {
			return svcCtx.UserRpc.Logout(ctx, req.ToRPC())
		},
	)
}

// CreateUser 创建用户
func CreateUser(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.CreateOrUpdateUserReq) (*userpb.CreateOrUpdateUserResp, error) {
			return svcCtx.UserRpc.CreateUser(ctx, req.ToRPC())
		},
	)
}

// UpdateUser 更新用户信息
func UpdateUser(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.CreateOrUpdateUserReq) (*userpb.CreateOrUpdateUserResp, error) {
			return svcCtx.UserRpc.UpdateUser(ctx, req.ToRPC())
		},
	)
}

// WxMiniProgramLogin 微信小程序登录
func WxMiniProgramLogin(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.WxMiniProgramLoginReq) (*userpb.WxMiniProgramLoginResp, error) {
			return svcCtx.UserRpc.WxMiniProgramLogin(ctx, req.ToRPC())
		},
	)
}

// RefreshToken 刷新token
func RefreshToken(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.RefreshTokenReq) (*userpb.RefreshTokenResp, error) {
			return svcCtx.UserRpc.RefreshToken(ctx, req.ToRPC())
		},
	)
}

// UserList 用户列表（需要认证）
func UserList(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.UserListReq) (*userpb.UserListResp, error) {
			return svcCtx.UserRpc.UserList(ctx, req.ToRPC())
		},
	)
}

// UserInfo 用户详情查询（需要认证）
func UserInfo(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.UserInfoReq) (*userpb.UserInfoRsp, error) {
			return svcCtx.UserRpc.UserInfo(ctx, req.ToRPC())
		},
	)
}
