FROM golang:1.26-alpine AS builder

WORKDIR /build

# 先拷贝 common 公共模块（支持 go.mod replace: common => ../common）
COPY rpc/common /common

# 拷贝本服务代码
COPY rpc/stock-rpc/go.mod rpc/stock-rpc/go.sum ./
RUN go mod download

COPY rpc/stock-rpc/ ./
RUN CGO_ENABLED=0 go build -o stock-rpc-app .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /build/stock-rpc-app .
COPY --from=builder /build/etc/stock.yaml etc/

EXPOSE 8082

ENTRYPOINT ["./stock-rpc-app", "-f", "etc/stock.yaml"]