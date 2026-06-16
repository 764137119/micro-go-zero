package types

// Request defines the interface that all API request DTOs must implement.
// RPCReq is the corresponding protobuf request type used by the RPC client.
type Request[RPCReq any] interface {
	ToRPC() *RPCReq
}
