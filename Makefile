# Makefile in the root directory

# Define all service directories. Add new services to this list.
SERVICES := services/api-rest-gateway services/auth services/friend-request-base services/user-base

# Phony targets prevent conflicts with file names
.PHONY: all proto build tidy docker-build clean up down down-hard logs help

# Set the default goal to 'help' so that running 'make' shows the user the available commands.
.DEFAULT_GOAL := help

# --- Docker Compose Commands ---
up:
	@echo "==> Starting all services with Docker Compose (builds if necessary)..."
	docker-compose up --build -d

down:
	@echo "==> Stopping all Docker Compose services..."
	docker-compose down

down-hard:
	@echo "==> Stopping all Docker Compose services and deleting the volumes..."
	docker-compose down -v

logs:
	@echo "==> Tailing logs for all running services..."
	docker-compose logs -f

# --- Service-Specific Commands (Loops) ---
all: build

proto:
	@echo "==> Generating protobuf files for all services..."
	@for service in $(SERVICES); do \
		echo "--> Generating protos for $$service"; \
		$(MAKE) -C $$service proto; \
	done

build:
	@echo "==> Building all services..."
	@for service in $(SERVICES); do \
		echo "--> Building $$service"; \
		$(MAKE) -C $$service build; \
	done

tidy:
	@echo "==> Tidying go modules for all services..."
	@for service in $(SERVICES); do \
		echo "--> Tidying $$service"; \
		$(MAKE) -C $$service tidy; \
	done

docker-build:
	@echo "==> Building Docker images for all services individually..."
	@for service in $(SERVICES); do \
		echo "--> Building Docker image for $$service"; \
		$(MAKE) -C $$service docker-build; \
	done


# The '|| true' prevents the command from failing if a service doesn't have a 'clean' rule.
clean:
	@echo "==> Cleaning all services..."
	@for service in $(SERVICES); do \
		echo "--> Cleaning $$service"; \
		$(MAKE) -C $$service clean || true; \
	done

# --- Help ---
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Docker Compose Targets:"
	@echo "  up             	 Start all services with docker-compose in detached mode."
	@echo "  down           	 Stop all services with docker-compose."
	@echo "  down-hard           Stop all services with docker-compose and deletes associated volumes."
	@echo "  logs           	 Tail the logs of all running services."
	@echo ""
	@echo "Individual Service Targets:"
	@echo "  all            	 Build all services (default)."
	@echo "  proto          	 Generate protobuf files for all services."
	@echo "  build          	 Build binaries for all services."
	@echo "  tidy           	 Run 'go mod tidy' for all services."
	@echo "  docker-build   	 Build Docker images for all services individually."
	@echo "  clean          	 Clean build artifacts for all services."