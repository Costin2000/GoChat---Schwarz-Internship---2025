BINARY_NAME=./main/message-base-service
DOCKER_IMAGE_NAME=message-base-service
DOCKER_IMAGE_TAG= latest

PROTO_DIR=./proto
PROTO_FILE=$(PROTO_DIR)/message.proto
PROTOC=protoc

GO_OUT=paths=source_relative:$(PROTO_DIR)
GO_GRPC_OUT=paths=source_relative:$(PROTO_DIR)

.PHONY: all proto build run tidy clean up down docker-build docker-run docker-stop test

all: build up down

proto:
	@echo "==> Generating protobuf files for message-base..."
	$(PROTOC) \
		-I $(PROTO_DIR) \
		-I ../.. \
		--go_out=$(GO_OUT) \
		--go-grpc_out=$(GO_GRPC_OUT) \
		$(PROTO_FILE)

tidy:
	@echo "==> Tidying go modules..."
	go mod tidy

build: clean proto tidy
	@echo "==> Building local binary..."
	go build -o $(BINARY_NAME) ./main

run: build 
	@echo "==> Running service locally..."
	@$(BINARY_NAME)

clean:
	@echo "==> Cleaning up compiled files..."
	rm -f $(BINARY_NAME) $(PROTO_DIR)/*.go

docker-build:
	@echo "==> Building Docker image: $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)..."
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) -f ./Dockerfile ../..

docker-run:
	@echo "==> Running Docker container..."
	docker run -p 50055:50055 --rm --name $(DOCKER_IMAGE_NAME) $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

down:
	@echo "==> Stopping Docker container..."
	docker stop $(DOCKER_IMAGE_NAME) || true

up: docker build docker-run

test:
	@echo "==> Running tests for main package..."
	@go test ./main -v