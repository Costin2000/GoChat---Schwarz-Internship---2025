BINARY_NAME=auth-service
DOCKER_IMAGE_NAME=auth-service
DOCKER_IMAGE_TAG=latest

PROTO_DIR=proto
PROTO_FILE=$(PROTO_DIR)/api.proto
PROTOC=protoc
GO_OUT=paths=source_relative:$(PROTO_DIR)
GO_GRPC_OUT=paths=source_relative:$(PROTO_DIR)

.PHONY: all proto build run tidy up down docker-build docker-run docker-stop

all: up

proto:
	@echo "==> Generating protobuf files..."
	$(PROTOC) \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(GO_OUT) \
		--go-grpc_out=$(GO_GRPC_OUT) \
		$(PROTO_FILE)

tidy:
	@echo "==> Tidying go modules..."
	go mod tidy

build: proto tidy
	@echo "==> Building local binary..."
	go build -o $(BINARY_NAME) ./main

run:
	@echo "==> Running service locally..."
	go run ./main

docker-build:
	@echo "==> Building Docker image: $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)..."
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .

docker-run:
	@echo "==> Running Docker container..."
	docker run -p 50053:50053 --rm --name $(DOCKER_IMAGE_NAME) $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

down:
	@echo "==> Stopping Docker container..."
	docker stop $(DOCKER_IMAGE_NAME) || true

up: docker-build docker-run