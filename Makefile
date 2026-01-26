# Makefile for grocer project

# Variables
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

# Copy proto files from pkg to internal/layout/third_party/pkg
update-proto:
	@echo "Copying proto files from pkg to internal/layout/third_party/pkg..."
	bash ./scripts/update-proto.sh
	@echo "Proto files copied successfully!"

# Generate protobuf files
gen-proto:
	@echo "Generating protobuf files..."
	bash ./scripts/gen-proto.sh

# Build the project
build:
	@echo "Building the project..."
	go build -ldflags="-X github.com/soyacen/grocer/cmd.Version=$(VERSION)" -o bin/grocer .

# Install the project
install:
	@echo "Installing the project..."
	go install -ldflags="-X github.com/soyacen/grocer/cmd.Version=$(VERSION)" .

help:
	@echo "Available targets:"
	@echo "  update-proto 	- Copy proto files from pkg to internal/layout/third_party/pkg"
	@echo "  gen-proto  	- Generate protobuf files"
	@echo "  build      	- Build the project"
	@echo "  install    	- Install the project"
	@echo "  help       	- Show this help message"

.PHONY: update-proto gen-proto build install help