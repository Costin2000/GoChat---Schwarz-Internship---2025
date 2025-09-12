FROM golang:1.24-alpine

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./ 
RUN go mod tidy

COPY . .

RUN go build -o service ./services/conversation-base/main

EXPOSE 50056

CMD ["./service"]