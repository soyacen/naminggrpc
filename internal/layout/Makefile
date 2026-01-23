# Makefile for grocer project

# Project name variable - customize as needed
PROJECT_NAME ?= $(shell echo $${PROJECT_NAME:-grocer})

# Image registry username variable - customize as needed
IMAGE_REGISTRY_USERNAME ?= $(shell echo $${IMAGE_REGISTRY_USERNAME:-username})

# Get the latest git tag as version, fallback to 'latest' if not tagged
GIT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "latest")
VERSION ?= $(shell echo $${VERSION:-$(GIT_TAG)})

# Generate protobuf files for all API subdirectories
protoc:
	@echo "Generating protobuf files for all API subdirectories..."
	@for dir in api/*; do \
		if [ -d "$$dir" ] && [ -f "$$dir/protoc.sh" ]; then \
			echo "Running $$dir/protoc.sh"; \
			bash "$$dir/protoc.sh"; \
		fi \
	done

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

help:
	@echo "Available targets:"
	@echo "  protoc                    - Generate protobuf files for all API subdirectories"
	@echo "  docker-build              - Build docker image with VERSION (default: latest git tag)"
	@echo "  docker-push               - Push docker image with VERSION (default: latest git tag)"
	@echo "  docker-release            - Build and push docker image (convenience target)"
	@echo "  help                      - Show this help message"

# Define a target that builds and pushes the image
docker-release: docker-build docker-push

.PHONY: protoc help docker-build docker-push docker-release