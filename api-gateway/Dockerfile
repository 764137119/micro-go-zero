FROM golang:1.26-alpine AS builder

WORKDIR /build

# 拷贝 go.mod replace 引用的本地模块（api-gateway import 了这些服务的包）
COPY rpc /rpc

# 拷贝本服务代码
COPY api-gateway/go.mod api-gateway/go.sum ./
RUN go mod download

COPY api-gateway/ ./
RUN CGO_ENABLED=0 go build -o api-gateway-app .

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/api-gateway-app .
COPY --from=builder /build/etc/config.yaml ./etc/

EXPOSE 8888

ENTRYPOINT ["./api-gateway-app"]