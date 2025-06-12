FROM golang:1.24.1 AS builder

WORKDIR /app

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o healthcheck-service cmd/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates && update-ca-certificates

WORKDIR /app

COPY --from=builder /app/healthcheck-service .
COPY config/ ./config/

CMD ["./healthcheck-service"]