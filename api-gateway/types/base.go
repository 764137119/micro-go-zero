package types

// Request defines the interface that all API request DTOs must implement.
// RPCReq is the corresponding protobuf request type used by the RPC client.
type Request[RPCReq any] interface {
	ToRPC() *RPCReq
}

// Response 统一的 API 响应结构体
type Response struct {
	Code int         `json:"code"`           // 业务码（gRPC codes.Code）
	Msg  string      `json:"msg"`            // 提示信息
	Data interface{} `json:"data,omitempty"` // 响应数据
}
