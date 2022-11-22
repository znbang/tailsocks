FROM golang:alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY . .
RUN go build -trimpath -ldflags "-s -w"

FROM alpine:latest
COPY --from=builder /app/tailsocks5 .
ENTRYPOINT /tailsocks5 -h "$FLY_APP_NAME-$FLY_REGION"
