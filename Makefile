.PHONY: help docker-build docker-run docker-stop docker-clean docker-dev docker-logs docker-shell

help:
	@echo "GO Authorization Framework - Docker Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  docker-build       Build Docker image for production"
	@echo "  docker-run         Run Docker container in production mode"
	@echo "  docker-stop        Stop running Docker container"
	@echo "  docker-clean       Remove Docker container and volumes"
	@echo "  docker-dev         Run Docker container in development mode"
	@echo "  docker-logs        View Docker container logs"
	@echo "  docker-shell       Open shell inside running container"
	@echo "  docker-rebuild     Rebuild Docker image (forced)"
	@echo "  docker-test        Build and run tests in Docker"
	@echo ""

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
