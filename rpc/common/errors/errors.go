package errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrNotFound 记录不存在
	ErrNotFound = status.Error(codes.NotFound, "记录不存在")
	// ErrDB 数据库错误
	ErrDB = status.Error(codes.Internal, "数据库错误")
	// ErrInvalidParam 参数错误
	ErrInvalidParam = status.Error(codes.InvalidArgument, "参数错误")
	// ErrDuplicate 重复记录
	ErrDuplicate = status.Error(codes.AlreadyExists, "重复记录")
)

// --- 工厂函数（支持动态消息） ---

// NewNotFound 返回记录不存在的错误
func NewNotFound(msg string) error {
	return status.Error(codes.NotFound, msg)
}

// NewInvalidParam 返回参数错误的错误
func NewInvalidParam(msg string) error {
	return status.Error(codes.InvalidArgument, msg)
}

// NewDuplicate 返回重复记录的错误
func NewDuplicate(msg string) error {
	return status.Error(codes.AlreadyExists, msg)
}

// NewUnauthenticated 返回未认证的错误
func NewUnauthenticated(msg string) error {
	return status.Error(codes.Unauthenticated, msg)
}

// NewPermissionDenied 返回无权限的错误
func NewPermissionDenied(msg string) error {
	return status.Error(codes.PermissionDenied, msg)
}

// NewInternal 返回服务端内部错误
func NewInternal(msg string) error {
	return status.Error(codes.Internal, msg)
}

// NewUnavailable 返回服务不可用的错误
func NewUnavailable(msg string) error {
	return status.Error(codes.Unavailable, msg)
}
