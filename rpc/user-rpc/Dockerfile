FROM golang:1.26-alpine AS builder

WORKDIR /build

# 先拷贝 common 公共模块（支持 go.mod replace: common => ../common）
COPY rpc/common /common

# 拷贝本服务代码
COPY rpc/user-rpc/go.mod rpc/user-rpc/go.sum ./
RUN go mod download

COPY rpc/user-rpc/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o user-rpc-app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /build/user-rpc-app .
COPY --from=builder /build/etc/user.yaml ./etc/user.yaml

EXPOSE 8080

ENV TZ=Asia/Shanghai
RUN ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

CMD ["./user-rpc-app", "-f", "etc/user.yaml"]