# Makefile for User API
# Usage: make [target]

# Variables
BINARY_NAME=user-api
BUILD_DIR=bin
CMD_PATH=cmd/app
CONFIG_PATH=configs/config.yaml
MIGRATIONS_DIR=docs/migrations
DOCKER_COMPOSE_FILE=docker-compose.yml
SWAGGER_DIR=docs/swagger

# Go variables
GO=go
GO_BUILD=$(GO) build
GO_TEST=$(GO) test
GO_MOD=$(GO) mod
GO_RUN=$(GO) run
GO_CLEAN=$(GO) clean

# Build flags
BUILD_FLAGS=-v
LDFLAGS=-w -s
DEBUG_LDFLAGS=

# Default target
all: build

## Development targets
.PHONY: run
run:
	@echo "Running application in development mode..."
	$(GO_RUN) $(CMD_PATH)/main.go

.PHONY: run-with-config
run-with-config:
	@echo "Running application with custom config..."
	CONFIG_PATH=$(CONFIG_PATH) $(GO_RUN) $(CMD_PATH)/main.go

.PHONY: run-debug
run-debug:
	@echo "Running application in debug mode..."
	DEBUG=true $(GO_RUN) $(CMD_PATH)/main.go

## Build targets
.PHONY: build
build: clean
	@echo "Building application..."
	$(GO_BUILD) $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)/main.go

.PHONY: build-debug
build-debug: clean
	@echo "Building application with debug information..."
	$(GO_BUILD) $(BUILD_FLAGS) -ldflags '$(DEBUG_LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-debug $(CMD_PATH)/main.go
	@echo "Debug build complete: $(BUILD_DIR)/$(BINARY_NAME)-debug"

.PHONY: build-linux
build-linux: clean
	@echo "Building Linux binary..."
	GOOS=linux GOARCH=amd64 $(GO_BUILD) $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(CMD_PATH)/main.go

.PHONY: build-windows
build-windows: clean
	@echo "Building Windows binary..."
	GOOS=windows GOARCH=amd64 $(GO_BUILD) $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME).exe $(CMD_PATH)/main.go

.PHONY: build-darwin
build-darwin: clean
	@echo "Building macOS binary..."
	GOOS=darwin GOARCH=amd64 $(GO_BUILD) $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-macos $(CMD_PATH)/main.go

## Test targets
.PHONY: test
test:
	@echo "Running tests..."
	$(GO_TEST) -v ./...

.PHONY: test-cover
test-cover:
	@echo "Running tests with coverage..."
	$(GO_TEST) -cover -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

.PHONY: test-race
test-race:
	@echo "Running tests with race detector..."
	$(GO_TEST) -race ./...

.PHONY: test-short
test-short:
	@echo "Running short tests..."
	$(GO_TEST) -short ./...

## Database targets
.PHONY: migrate-up
migrate-up:
	@echo "Running database migrations..."
	@if [ -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		$(BUILD_DIR)/$(BINARY_NAME) migrate; \
	else \
		$(GO_RUN) $(CMD_PATH)/main.go migrate; \
	fi

.PHONY: migrate-down
migrate-down:
	@echo "Rolling back database migrations..."
	@if [ -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		$(BUILD_DIR)/$(BINARY_NAME) migrate down; \
	else \
		$(GO_RUN) $(CMD_PATH)/main.go migrate down; \
	fi

.PHONY: migrate-create
migrate-create:
	@echo "Creating new migration file..."
	@read -p "Enter migration name: " name; \
	timestamp=$$(date +%Y%m%d%H%M%S); \
	up_file="$(MIGRATIONS_DIR)/$${timestamp}_$${name}.up.sql"; \
	down_file="$(MIGRATIONS_DIR)/$${timestamp}_$${name}.down.sql"; \
	touch "$$up_file" "$$down_file"; \
	echo "Created migration files: $${up_file}, $${down_file}"

## Docker targets
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest .

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(BINARY_NAME):latest

.PHONY: docker-compose-up
docker-compose-up:
	@echo "Starting services with Docker Compose..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) up --build

.PHONY: docker-compose-down
docker-compose-down:
	@echo "Stopping services with Docker Compose..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

.PHONY: docker-compose-logs
docker-compose-logs:
	@echo "Showing Docker Compose logs..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

## Development utilities
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GO_MOD) download

.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	$(GO_MOD) download
	$(GO_MOD) tidy

.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Installing..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2; \
		golangci-lint run; \
	fi

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

.PHONY: vet
vet:
	@echo "Vetting code..."
	$(GO) vet ./...

.PHONY: swaggo
swaggo:
	@echo "Generating Swagger documentation..."
	@if command -v swag >/dev/null; then \
		swag init -g $(CMD_PATH)/main.go -o $(SWAGGER_DIR); \
	else \
		echo "swag not installed. Installing..."; \
		$(GO) install github.com/swaggo/swag/cmd/swag@latest; \
		swag init -g $(CMD_PATH)/main.go -o $(SWAGGER_DIR); \
	fi

.PHONY: dev
dev: deps fmt vet lint test build
	@echo "Development build complete!"

## Debug targets
.PHONY: debug-attach
debug-attach: build-debug
	@echo "Starting application for debugging..."
	@echo "Build with debug symbols: $(BUILD_DIR)/$(BINARY_NAME)-debug"
	@echo "Use: dlv exec $(BUILD_DIR)/$(BINARY_NAME)-debug --headless --listen=:2345 --api-version=2 --accept-multiclient"

.PHONY: debug-test
debug-test:
	@echo "Debugging tests..."
	$(GO_TEST) -exec delve ./...

## Monitoring targets
.PHONY: monitor
monitor:
	@echo "Starting monitoring..."
	@if command -v htop >/dev/null; then \
		htop; \
	else \
		top; \
	fi

.PHONY: ps
ps:
	@echo "Checking running processes..."
	ps aux | grep $(BINARY_NAME) | grep -v grep

## Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Development:"
	@echo "  run           - Run application in development mode"
	@echo "  run-debug     - Run application in debug mode"
	@echo "  dev           - Full development build (deps, fmt, vet, lint, test, build)"
	@echo ""
	@echo "Build:"
	@echo "  build         - Build production binary"
	@echo "  build-debug   - Build with debug information"
	@echo "  build-linux   - Build Linux binary"
	@echo "  build-windows - Build Windows binary"
	@echo "  build-darwin  - Build macOS binary"
	@echo ""
	@echo "Testing:"
	@echo "  test          - Run all tests"
	@echo "  test-cover    - Run tests with coverage report"
	@echo "  test-race     - Run tests with race detector"
	@echo "  test-short    - Run only short tests"
	@echo ""
	@echo "Database:"
	@echo "  migrate-up    - Run database migrations"
	@echo "  migrate-down  - Rollback database migrations"
	@echo "  migrate-create - Create new migration files"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build        - Build Docker image"
	@echo "  docker-run          - Run Docker container"
	@echo "  docker-compose-up   - Start services with Docker Compose"
	@echo "  docker-compose-down - Stop Docker Compose services"
	@echo ""
	@echo "Utilities:"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  deps-update   - Update dependencies"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo "  swagger       - Generate Swagger documentation"
	@echo ""
	@echo "Debug:"
	@echo "  debug-attach  - Build and prepare for debugger attachment"
	@echo "  debug-test    - Debug tests with delve"
	@echo ""
	@echo "Monitoring:"
	@echo "  monitor       - System monitoring"
	@echo "  ps            - Check running processes"
	@echo ""
	@echo "Help:"
	@echo "  help          - Show this help message"

# Create build directory if it doesn't exist
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Default target
.DEFAULT_GOAL := help