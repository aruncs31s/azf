# AZF Authorization Framework

## Environment Configuration

Before running the application, copy `.env.example` to `.env` and configure your environment variables:

```bash
cp .env.example .env
```

### Required Configuration

1. **JWT_SECRET**: Must be at least 32 characters. Generate with:
   ```bash
   openssl rand -base64 64
   ```

### Development

```bash
# Install dependencies
go mod download

# Run the application
make run

# Run tests
make test

# Run tests with coverage
make coverage

# Run linter
make lint

# Format code
make fmt
```

### Docker

```bash
# Build and run with Docker
make docker-run

# Stop Docker container
make docker-stop

# View logs
make docker-logs
```

See the [full documentation](README.md) for more details.
