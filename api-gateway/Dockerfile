FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
COPY config ./config
COPY svc ./svc
COPY handler ./handler
COPY middleware ./middleware
COPY main.go router.go ./

RUN go mod download && \
    CGO_ENABLED=0 go build -o api-gateway .

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/api-gateway .
COPY etc/config.yaml ./etc/

EXPOSE 8888

ENTRYPOINT ["./api-gateway"]