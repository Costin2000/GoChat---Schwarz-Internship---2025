BINARY_NAME=api-rest-gateway
DOCKER_IMAGE_NAME=api-rest-gateway
DOCKER_IMAGE_TAG=latest

PROTO_DIR=proto
PROTO_FILE=$(PROTO_DIR)/api.proto
PROTOC=protoc

GO_OUT=paths=source_relative:$(PROTO_DIR)
GO_GRPC_OUT=paths=source_relative:$(PROTO_DIR)
GW_OUT=paths=source_relative:$(PROTO_DIR)

GOPATH_DIR=$(shell go env GOPATH)
GOOGLEAPIS=$(GOPATH_DIR)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@*/third_party/googleapis

# mapping: imported proto -> valid Go import path
M_AUTH=Mservices/auth/proto/api.proto=github.com/Costin2000/GoChat---Schwarz-Internship---2025/auth/proto

.PHONY: all deps proto tidy build run docker-build docker-run up down clean
all: up

deps:
	@echo "==> Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

proto: deps
	@echo "==> Generating protobuf + grpc-gateway..."
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
	@echo "==> go mod tidy..."
	go mod tidy

build: proto tidy
	@echo "==> Building local binary..."
	GO111MODULE=on CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o $(BINARY_NAME) ./main

run:
	@echo "==> Running service locally..."
	GATEWAY_HTTP_ADDR=:8080 AUTH_ADDR=auth:50053 ./$(BINARY_NAME)

docker-build:
	@echo "==> Building Docker image: $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)..."
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .

docker-run:
	@echo "==> Running Docker container..."
	docker run -p 8080:8080 --rm --name $(DOCKER_IMAGE_NAME) \
	  -e GATEWAY_HTTP_ADDR=:8080 \
	  -e AUTH_ADDR=auth:50053 \
	  $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

up: docker-build docker-run

down:
	@echo "==> Stopping Docker container..."
	docker stop $(DOCKER_IMAGE_NAME) || true

clean:
	rm -f $(BINARY_NAME)
	# păstrăm .pb.go generate în proto/