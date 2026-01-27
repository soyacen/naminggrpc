# Makefile for grocer project

# Project name variable - customize as needed
PROJECT_NAME ?= $(shell echo $${PROJECT_NAME:-grocer})

# Image registry username variable - customize as needed
IMAGE_REGISTRY_USERNAME ?= $(shell echo $${IMAGE_REGISTRY_USERNAME:-username})

# Get the latest git tag as version, fallback to 'latest' if not tagged
GIT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "latest")
VERSION ?= $(shell echo $${VERSION:-$(GIT_TAG)})

NACOS_CONFIG_IP=127.0.0.1
NACOS_CONFIG_PORT=8848
NACOS_CONFIG_NAMESPACE=public
NACOS_CONFIG_GROUP=DEFAULT_GROUP
NACOS_CONFIG_DATA_ID=vibeme-micro.yaml

# Build binary to bin directory
build:
	@echo "Building binary to bin directory..."
	@mkdir -p bin
	@go build -o bin/$(PROJECT_NAME) .

# Run grpc script to compile all proto files in api directory and subdirectories
grpc:
	@echo "Compiling proto files using grpc.sh..."
	@bash scripts/grpc.sh

# Run config script to compile config proto files
config:
	@echo "Compiling config proto files using config.sh..."
	@bash scripts/config.sh

# Run goose script to compile all proto files in api directory and subdirectories
goose:
	@echo "Compiling proto files using goose.sh..."
	@bash scripts/goose.sh

# Build docker image, optionally takes TAG argument (defaults to latest)
docker-build:
	@if [ "$(IMAGE_REGISTRY_USERNAME)" = "username" ]; then \
		echo "ERROR: IMAGE_REGISTRY_USERNAME is set to default value."; \
		exit 1; \
	fi
	@echo "Building docker image: $(IMAGE_REGISTRY_USERNAME)/$(PROJECT_NAME):$(VERSION)"
	docker build -t $(IMAGE_REGISTRY_USERNAME)/$(PROJECT_NAME):$(VERSION) .
	@echo "Successfully built image: $(IMAGE_REGISTRY_USERNAME)/$(PROJECT_NAME):$(VERSION)"

# Push docker image, optionally takes TAG argument (defaults to latest)
docker-push:
	@if [ "$(IMAGE_REGISTRY_USERNAME)" = "username" ]; then \
		echo "ERROR: IMAGE_REGISTRY_USERNAME is set to default value."; \
		exit 1; \
	fi
	@echo "Pushing docker image: $(IMAGE_REGISTRY_USERNAME)/$(PROJECT_NAME):$(VERSION)"
	docker push $(IMAGE_REGISTRY_USERNAME)/$(PROJECT_NAME):$(VERSION)
	@echo "Successfully pushed image: $(IMAGE_REGISTRY_USERNAME)/$(PROJECT_NAME):$(VERSION)"

# Run cronjob locally with required environment variables
run-cronjob:
	@echo "Running cronjob locally..."
	NACOS_CONFIG_IP=$(NACOS_CONFIG_IP) \
	NACOS_CONFIG_PORT=$(NACOS_CONFIG_PORT) \
	NACOS_CONFIG_NAMESPACE=$(NACOS_CONFIG_NAMESPACE) \
	NACOS_CONFIG_GROUP=$(NACOS_CONFIG_GROUP) \
	NACOS_CONFIG_DATA_ID=$(NACOS_CONFIG_DATA_ID) \
	go run main.go cronjob

# Run job locally
run-job:
	@echo "Running job locally..."
	go run main.go job

# Run grpc service locally
run-grpc:
	@echo "Running grpc service locally..."
	go run main.go grpc

# Run http service locally
run-http:
	@echo "Running http service locally..."
	go run main.go http

help:
	@echo "Available targets:"
	@echo "  build                     - Build binary to bin directory"
	@echo "  grpc                      - Compile proto files using grpc.sh script"
	@echo "  config                    - Compile config proto files"
	@echo "  goose                     - Compile proto files using goose.sh script"
	@echo "  docker-build              - Build docker image with VERSION (default: latest git tag)"
	@echo "  docker-push               - Push docker image with VERSION (default: latest git tag)"
	@echo "  docker-release            - Build and push docker image (convenience target)"
	@echo "  run-cronjob               - Run cronjob locally with env vars (set NACOS_CONFIG_*, NACOS_CONFIG_GROUP, NACOS_CONFIG_DATA_ID)"
	@echo "  run-job                   - Run job locally"
	@echo "  run-grpc                  - Run grpc service locally"
	@echo "  run-http                  - Run http service locally"
	@echo "  help                      - Show this help message"

# Define a target that builds and pushes the image
docker-release: docker-build docker-push

.PHONY: build protoc help docker-build docker-push docker-release grpc config goose run-cronjob run-job run-grpc run-http