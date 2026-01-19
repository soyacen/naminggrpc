# Makefile for grocer project

# Generate protobuf files
protoc:
	@echo "Generating protobuf files..."
	bash ./scripts/protoc-all.sh

help:
	@echo "Available targets:"
	@echo "  protoc  - Generate protobuf files"
	@echo "  help    - Show this help message"

.PHONY: protoc help