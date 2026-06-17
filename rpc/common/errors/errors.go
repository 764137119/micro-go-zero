package errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNotFound     = status.Error(codes.NotFound, "记录不存在")
	ErrDB           = status.Error(codes.Internal, "数据库错误")
	ErrInvalidParam = status.Error(codes.InvalidArgument, "参数错误")
	ErrDuplicate    = status.Error(codes.AlreadyExists, "重复记录")
)
