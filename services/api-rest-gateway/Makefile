BINARY_NAME=./main/api-rest-gateway-service
DOCKER_IMAGE_NAME=api-rest-gateway
DOCKER_IMAGE_TAG=latest

PROTO_DIR=proto
PROTO_FILE=$(PROTO_DIR)/gateway.proto
PROTOC=protoc
GO_OUT=paths=source_relative:$(PROTO_DIR)
GO_GRPC_OUT=paths=source_relative:$(PROTO_DIR)

GW_OUT=paths=source_relative:$(PROTO_DIR)
GOPATH_DIR=$(shell go env GOPATH)
GOOGLEAPIS=$(GOPATH_DIR)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@*/third_party/googleapis
M_AUTH=Mservices/auth/proto/auth.proto=github.com/Costin2000/GoChat---Schwarz-Internship---2025/services/auth/proto

.PHONY: all proto build run tidy up down docker-build docker-run

all: up

proto:
	@echo "==> Generating protobuf + grpc-gateway files..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	$(PROTOC) \
		-I $(PROTO_DIR) \
		-I $(GOOGLEAPIS) \
		-I ../.. \
		--go_out=$(GO_OUT) \
		--go_opt=$(M_AUTH) \
		--go-grpc_out=$(GO_GRPC_OUT) \
		--go-grpc_opt=$(M_AUTH) \
		--grpc-gateway_out=$(GW_OUT) \
		--grpc-gateway_opt=$(M_AUTH) \
		$(PROTO_FILE)

tidy:
	@echo "==> Tidying go modules (from root)..."
	cd ../.. && go mod tidy

build: proto tidy
	@echo "==> Building local binary..."
	cd ../.. && go build -o ./services/api-rest-gateway/$(BINARY_NAME) ./services/api-rest-gateway/main

run:
	@echo "==> Running service locally..."
	# Notă: Această comandă presupune că serviciul 'auth' rulează separat.
	GATEWAY_HTTP_ADDR=:8080 AUTH_ADDR=localhost:50053 go run ./main

docker-build:
	@echo "==> Building Docker image: $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)..."
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) -f ./Dockerfile ../..

docker-run:
	@echo "==> Running Docker container..."
	# Notă: Această comandă este pentru testare individuală. Folosește 'docker-compose up' pentru a rula totul.
	docker run -p 8080:8080 --rm --name $(DOCKER_IMAGE_NAME) $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

down:
	@echo "==> Stopping Docker container..."
	docker stop $(DOCKER_IMAGE_NAME) || true

up: docker-build docker-run