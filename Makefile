.PHONY: help docker-build docker-run docker-stop docker-clean docker-dev docker-logs docker-shell \
        lint test coverage generate tidy security build run clean fmt

help:
	@echo "GO Authorization Framework - Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Development:"
	@echo "  build            Build the application"
	@echo "  run              Run the application"
	@echo "  test             Run all tests"
	@echo "  test-verbose     Run tests with verbose output"
	@echo "  coverage         Run tests with coverage report"
	@echo "  coverage-html    Generate HTML coverage report"
	@echo "  lint             Run golangci-lint"
	@echo "  lint-fix         Run golangci-lint with auto-fix"
	@echo "  fmt              Format code with gofmt and goimports"
	@echo "  tidy             Run go mod tidy"
	@echo "  generate         Run go generate and templ generate"
	@echo "  security         Run security checks (gosec, govulncheck)"
	@echo "  clean            Clean build artifacts"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build     Build Docker image for production"
	@echo "  docker-run       Run Docker container in production mode"
	@echo "  docker-stop      Stop running Docker container"
	@echo "  docker-clean     Remove Docker container and volumes"
	@echo "  docker-dev       Run Docker container in development mode"
	@echo "  docker-logs      View Docker container logs"
	@echo "  docker-shell     Open shell inside running container"
	@echo "  docker-rebuild   Rebuild Docker image (forced)"
	@echo "  docker-test      Build and run tests in Docker"
	@echo ""

# =============================================================================
# Development Commands
# =============================================================================

build:
	@echo "Building application..."
	go build -o bin/azf ./cmd/...

run: build
	@echo "Running application..."
	./bin/azf

test:
	@echo "Running tests..."
	go test ./... -race

test-verbose:
	@echo "Running tests with verbose output..."
	go test -v ./... -race

coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

coverage-html: coverage
	@echo "Generating HTML coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

lint:
	@echo "Running linter..."
	golangci-lint run ./...

lint-fix:
	@echo "Running linter with auto-fix..."
	golangci-lint run --fix ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

tidy:
	@echo "Tidying modules..."
	go mod tidy
	go mod verify

generate:
	@echo "Running code generation..."
	go generate ./...
	templ generate

security:
	@echo "Running security checks..."
	@command -v gosec >/dev/null 2>&1 || { echo "Installing gosec..."; go install github.com/securego/gosec/v2/cmd/gosec@latest; }
	gosec ./...
	@command -v govulncheck >/dev/null 2>&1 || { echo "Installing govulncheck..."; go install golang.org/x/vuln/cmd/govulncheck@latest; }
	govulncheck ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# =============================================================================
# Pre-commit and CI
# =============================================================================

pre-commit: fmt lint test
	@echo "Pre-commit checks passed!"

ci: lint test coverage
	@echo "CI checks passed!"

# =============================================================================
# Docker Commands
# =============================================================================

docker-build:
	@echo "Building Docker image..."
	docker build -t AZFafz:latest .

docker-run: docker-build
	@echo "Running Docker container..."
	docker-compose up -d

docker-stop:
	@echo "Stopping Docker container..."
	docker-compose down

docker-clean: docker-stop
	@echo "Cleaning up Docker resources..."
	docker-compose down -v
	docker rmi AZF:latest 2>/dev/null || true

docker-dev:
	@echo "Running Docker container in development mode..."
	docker-compose -f docker-compose.dev.yml up -d

docker-logs:
	@echo "Showing Docker container logs..."
	docker-compose logs -f

docker-shell:
	@echo "Opening shell in running container..."
	docker exec -it azf sh

docker-rebuild:
	@echo "Rebuilding Docker image..."
	docker build -t AZFazf:latest --no-cache .

docker-test:
	@echo "Running tests in Docker..."
	docker build -t AZFazf:test --target builder .
	docker run --rm AZFazf:test go test ./...

docker-ps:
	@echo "Docker container status:"
	docker-compose ps

docker-health:
	@echo "Checking container health..."
	docker-compose exec authz-framework sh -c "wget --quiet --tries=1 --spider http://localhost:8080/admin-ui/login && echo 'OK' || echo 'UNHEALTHY'"
