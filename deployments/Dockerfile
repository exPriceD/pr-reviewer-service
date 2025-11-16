FROM golang:1.24.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pr-reviewer-service ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /root/

COPY --from=builder /app/pr-reviewer-service .
COPY --from=builder /app/configs ./configs

EXPOSE 8080

CMD ["./pr-reviewer-service"]

