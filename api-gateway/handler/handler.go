package handler

import (
	"context"
	"net/http"

	"api-gateway/types"
	"common/errors"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

// Success 统一成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, types.Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

// Error 统一错误响应（从 gRPC error 中提取状态码和消息）
func Error(c *gin.Context, err error) {
	httpStatus := errors.GRPCStatusToHTTPCode(err)
	st, _ := status.FromError(err)
	c.JSON(httpStatus, types.Response{
		Code: int(st.Code()),
		Msg:  st.Message(),
	})
}

// BadRequest 参数绑定失败的统一响应
func BadRequest(c *gin.Context, err error) {
	paramErr := errors.NewInvalidParam(err.Error())
	Error(c, paramErr)
}

// HandleJSON is a generic wrapper for JSON body request handlers.
// It eliminates the boilerplate of binding JSON, calling RPC, and returning the response.
//
// Usage:
//
//	func MyHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
//	    return HandleJSON(
//	        func(ctx context.Context, req *types.MyReq) (*orderpb.MyResp, error) {
//	            return svcCtx.OrderRpc.MyMethod(ctx, req.ToRPC())
//	        },
//	    )
//	}
func HandleJSON[Req any, Resp any](
	rpcFunc func(context.Context, *Req) (*Resp, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequest(c, err)
			return
		}
		resp, err := rpcFunc(c.Request.Context(), &req)
		if err != nil {
			Error(c, err)
			return
		}
		Success(c, resp)
	}
}

// HandleQuery is a generic wrapper for URL query parameter handlers.
func HandleQuery[Req any, Resp any](
	rpcFunc func(context.Context, *Req) (*Resp, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req
		if err := c.ShouldBindQuery(&req); err != nil {
			BadRequest(c, err)
			return
		}
		resp, err := rpcFunc(c.Request.Context(), &req)
		if err != nil {
			Error(c, err)
			return
		}
		Success(c, resp)
	}
}
