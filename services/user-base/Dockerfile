FROM golang:1.24-alpine

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o service ./main/service.go ./main/user_base.go

EXPOSE 50051

CMD ["./service"]