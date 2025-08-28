BINARY_NAME=./main/userBase-service
DOCKER_IMAGE_NAME=userbase-service
DOCKER_IMAGE_TAG=latest

PROTO_DIR=./proto
PROTO_FILE=$(PROTO_DIR)/userbase.proto
PROTOC_INCLUDE=/usr/local/include
PROTOC=protoc
GO_OUT=paths=source_relative:$(PROTO_DIR)
GO_GRPC_OUT=paths=source_relative:$(PROTO_DIR)

.PHONY: all proto build run tidy clean up down docker-build docker-run docker-stop test test-unit test-integration

all: build up down

proto:
	@echo "==> Generating protobuf files..."
	$(PROTOC) \
		--proto_path=$(PROTO_DIR) \
		-I=/usr/local/include \
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
	@echo "==> Starting local database..."
	@(cd ./scripts && ./db-start.sh)
	@echo "==> Running service locally..."
	@./main/userBase-service & \
	PID=$$!; \
	trap "echo '==> Shutting down server...'; kill $$PID; echo '==> Stopping local database...'; (cd ./scripts && ./db-stop.sh)" INT TERM; \
	echo "Server started with PID: $$PID"; \
	sleep 2; \
	wait $$PID;
    
clean:
	@echo "==> Cleaning up compiled files..."
	rm -f $(BINARY_NAME) $(PROTO_DIR)/*.go

docker-build:
	@echo "==> Building Docker image: $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)..."
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) -f ./Dockerfile ../..

docker-run:
	@echo "==> Running Docker container..."
	docker run -p 50051:50051 --rm --name $(DOCKER_IMAGE_NAME) $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

down:
	@echo "==> Stopping Docker container..."
	docker stop $(DOCKER_IMAGE_NAME) || true

up: docker-build docker-run

test:
	@echo "==> Running tests for main package..."
	@go test ./main -v
