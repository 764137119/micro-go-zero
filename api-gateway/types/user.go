package types

import (
	userpb "user-rpc/user"
)

// LoginReq 用户登录请求
type LoginReq struct {
	Mobile   string `json:"mobile" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (r *LoginReq) ToRPC() *userpb.LoginReq {
	return &userpb.LoginReq{
		Mobile:   r.Mobile,
		Password: r.Password,
	}
}

// LogoutReq 用户登出请求
type LogoutReq struct {
	UserID int64 `json:"user_id" binding:"required,min=1"`
}

func (r *LogoutReq) ToRPC() *userpb.LogoutReq {
	return &userpb.LogoutReq{
		UserId: r.UserID,
	}
}

// CreateOrUpdateUserReq 创建/更新用户请求
type CreateOrUpdateUserReq struct {
	UserID   int64  `json:"user_id"`
	Mobile   string `json:"mobile" binding:"required"`
	NickName string `json:"nick_name"`
	Sex      int32  `json:"sex"`
	Avatar   string `json:"avatar"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (r *CreateOrUpdateUserReq) ToRPC() *userpb.CreateOrUpdateUserReq {
	return &userpb.CreateOrUpdateUserReq{
		UserId:   r.UserID,
		Mobile:   r.Mobile,
		NickName: r.NickName,
		Sex:      r.Sex,
		Avatar:   r.Avatar,
		Password: r.Password,
		Email:    r.Email,
	}
}

// WxMiniProgramLoginReq 微信小程序登录请求
type WxMiniProgramLoginReq struct {
	Code     string `json:"code" binding:"required"`
	NickName string `json:"nick_name"`
	Avatar   string `json:"avatar"`
	Sex      int32  `json:"sex"`
}

func (r *WxMiniProgramLoginReq) ToRPC() *userpb.WxMiniProgramLoginReq {
	return &userpb.WxMiniProgramLoginReq{
		Code:     r.Code,
		NickName: r.NickName,
		Avatar:   r.Avatar,
		Sex:      r.Sex,
	}
}

// RefreshTokenReq 刷新token请求
type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (r *RefreshTokenReq) ToRPC() *userpb.RefreshTokenReq {
	return &userpb.RefreshTokenReq{
		RefreshToken: r.RefreshToken,
	}
}

// UserListReq 用户列表请求
type UserListReq struct {
	Page     int32  `json:"page" binding:"required,min=1"`
	PageSize int32  `json:"page_size" binding:"required,min=1,max=100"`
	Mobile   string `json:"mobile"`
	NickName string `json:"nick_name"`
}

func (r *UserListReq) ToRPC() *userpb.UserListReq {
	return &userpb.UserListReq{
		Page:     r.Page,
		PageSize: r.PageSize,
		Mobile:   r.Mobile,
		NickName: r.NickName,
	}
}

// UserInfoReq 用户详情查询请求
type UserInfoReq struct {
	UserID int64 `json:"user_id" binding:"required,min=1"`
}

func (r *UserInfoReq) ToRPC() *userpb.UserInfoReq {
	return &userpb.UserInfoReq{
		UserId: r.UserID,
	}
}
