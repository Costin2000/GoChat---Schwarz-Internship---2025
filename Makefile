BINARY_NAME=./services/user-base/main/userBase-service
DOCKER_IMAGE_NAME=userbase-service
DOCKER_IMAGE_TAG=latest

PROTO_DIR=./services/user-base/proto
PROTO_FILE=$(PROTO_DIR)/api.proto
PROTOC=protoc
GO_MAIN_SOURCE=./services/user-base/main/service.go
GO_AUX_SOURCE=./services/user-base/main/user_base.go
GO_OUT=paths=source_relative:$(PROTO_DIR)
GO_GRPC_OUT=paths=source_relative:$(PROTO_DIR)

.PHONY: all proto build run tidy clean up down docker-build docker-run docker-stop

all: build up down

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

build: clean proto tidy
	@echo "==> Building local binary..."
	go build -o $(BINARY_NAME) ./$(GO_MAIN_SOURCE) ./$(GO_AUX_SOURCE)

run: build
	@echo "==> Running service locally..."
	@./$(BINARY_NAME) & \
	SERVER_PID=$$!; \
	echo "Server started with PID: $$SERVER_PID"; \
	echo "Waiting for server to initialize..."; \
	sleep 2; \
	echo "==> Sending request..."; \
	grpcurl -plaintext localhost:50051 user_base.UserBase/Ping; \
	echo "==> Shutting down server..."; \
	kill $$SERVER_PID;
	
clean:
	@echo "==> Cleaning up compiled files..."
	rm -f $(BINARY_NAME) $(PROTO_DIR)/*.go

docker-build:
	@echo "==> Building Docker image: $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)..."
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .

docker-run:
	@echo "==> Running Docker container..."
	docker run -p 50051:50051 --rm --name $(DOCKER_IMAGE_NAME) $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

down:
	@echo "==> Stopping Docker container..."
	docker stop $(DOCKER_IMAGE_NAME) || true

up: docker-build docker-run