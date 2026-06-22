FROM golang:1.26-alpine AS builder

WORKDIR /build

# 先拷贝 common 公共模块（支持 go.mod replace: common => ../common）
COPY rpc/common /common

# 拷贝本服务代码
COPY rpc/order-rpc/go.mod rpc/order-rpc/go.sum ./
RUN go mod download

COPY rpc/order-rpc/ ./
RUN CGO_ENABLED=0 go build -o order-rpc-app .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /build/order-rpc-app .
COPY --from=builder /build/etc/order.yaml etc/

EXPOSE 8081

ENTRYPOINT ["./order-rpc-app", "-f", "etc/order.yaml"]