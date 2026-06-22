FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY rpc/user-rpc/user-rpc .
COPY rpc/user-rpc/etc/user.yaml ./etc/user.yaml

EXPOSE 8080

ENV TZ=Asia/Shanghai
RUN ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

CMD ["./user-rpc", "-f", "etc/user.yaml"]