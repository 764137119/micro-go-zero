FROM golang:1.26-alpine AS builder

WORKDIR /build

# 先拷贝 common 公共模块（支持 go.mod replace: common => ../common）
COPY rpc/common /common

# 拷贝本服务代码
COPY rpc/cronjob-rpc/go.mod rpc/cronjob-rpc/go.sum ./
RUN go mod download

COPY rpc/cronjob-rpc/ ./
RUN CGO_ENABLED=0 go build -o cronjob-rpc-app .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /build/cronjob-rpc-app .
COPY --from=builder /build/etc/cronjob.yaml etc/

EXPOSE 8083

ENTRYPOINT ["./cronjob-rpc-app", "-f", "etc/cronjob.yaml"]