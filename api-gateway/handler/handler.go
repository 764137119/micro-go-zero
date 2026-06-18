package handler

import (
	"context"
	"net/http"

	"common/errors"

	"github.com/gin-gonic/gin"
)

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
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		resp, err := rpcFunc(c.Request.Context(), &req)
		if err != nil {
			httpStatus := errors.GRPCStatusToHTTPCode(err)
			c.JSON(httpStatus, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// HandleQuery is a generic wrapper for URL query parameter handlers.
func HandleQuery[Req any, Resp any](
	rpcFunc func(context.Context, *Req) (*Resp, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		resp, err := rpcFunc(c.Request.Context(), &req)
		if err != nil {
			httpStatus := errors.GRPCStatusToHTTPCode(err)
			c.JSON(httpStatus, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}
