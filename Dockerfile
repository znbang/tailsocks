FROM golang:alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY . .
RUN go build -trimpath -ldflags "-s -w"

FROM alpine:latest
COPY --from=builder /app/tailproxy .
ENTRYPOINT /tailproxy -hostname "$FLY_APP_NAME-$FLY_REGION"
