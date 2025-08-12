FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY services/auth/go.mod services/auth/go.sum ./services/auth/
COPY services/api-rest-gateway/go.mod services/api-rest-gateway/go.sum ./services/api-rest-gateway/

WORKDIR /app/services/api-rest-gateway
RUN go mod tidy


COPY . .

WORKDIR /app/services/api-rest-gateway

RUN CGO_ENABLED=0 go build -o /app/service ./main

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/service .

EXPOSE 8080

CMD ["./service"]
