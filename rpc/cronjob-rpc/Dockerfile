FROM golang:1.26-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o cronjob-rpc .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /build/cronjob-rpc .
COPY --from=builder /build/etc/cronjob.yaml etc/

EXPOSE 8083

ENTRYPOINT ["./cronjob-rpc", "-f", "etc/cronjob.yaml"]