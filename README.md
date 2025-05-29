# Agent Task Manager

A simple Go service with health check endpoints for Kubernetes.

## Quick Start with Makefile

This project includes a comprehensive Makefile for easy development and deployment:

```bash
# Show all available commands
make help

# Run locally for development
make run

# Build and test locally
make quick-build

# Build multi-platform image and push to Docker Hub
make DOCKER_USERNAME=yourusername build-and-push

# Full release with formatting, vetting, building and pushing
make DOCKER_USERNAME=yourusername release
```

## Installation and Setup

```bash
# Install dependencies
go mod download
# or
make deps

# Run the service
go run .
# or 
make run
```

The service will start on port 8081.

## Docker & Kubernetes Deployment

### Using Makefile (Recommended)

```bash
# 1. Build and push multi-platform Docker image
make DOCKER_USERNAME=yourusername build-and-push

# 2. Update your Kubernetes manifests with your image name
# 3. Deploy to Kubernetes
kubectl apply -f your-k8s-manifests.yaml
```

### Manual Docker Commands

```bash
# Build multi-platform image
docker buildx build --platform linux/amd64,linux/arm64 \
  -t yourusername/agent-task-manager:latest --push .

# Test locally
docker run -p 8081:8081 yourusername/agent-task-manager:latest
```

## API Endpoints

### Health Check
- **GET** `/health` - Service liveness check (liveness probe)
  
  Response:
  ```json
  {
    "status": "alive",
    "message": "Service is running"
  }
  ```

### Ready Check
- **GET** `/ready` - Service readiness check (readiness probe)
  
  Response:
  ```json
  {
    "status": "ready",
    "message": "Service is ready to accept requests"
  }
  ```

## Kubernetes Configuration Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agent-task-manager
spec:
  replicas: 3
  selector:
    matchLabels:
      app: agent-task-manager
  template:
    metadata:
      labels:
        app: agent-task-manager
    spec:
      containers:
      - name: agent-task-manager
        image: yourusername/agent-task-manager:latest  # Update with your Docker Hub username
        ports:
        - containerPort: 8081
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 5
```

## Available Makefile Commands

### Development
- `make run` - Run application locally
- `make dev` - Run in development mode
- `make deps` - Download Go dependencies
- `make fmt` - Format Go code
- `make vet` - Run Go vet
- `make tidy` - Tidy Go modules

### Docker
- `make build` - Build Docker image for current platform
- `make build-multi` - Build multi-platform Docker image
- `make test-local` - Build and test image locally
- `make push` - Push image to Docker Hub
- `make build-and-push` - Build multi-platform and push to Docker Hub

### Utilities
- `make clean` - Clean up Docker images and containers
- `make inspect-image` - Inspect multi-platform image details
- `make release` - Full release process (format + vet + build + push)

## Environment Variables

- `DOCKER_USERNAME` - Your Docker Hub username
- `TAG` - Image tag (default: latest)

## Testing the Service

```bash
# Test health endpoint
curl http://localhost:8081/health

# Test ready endpoint
curl http://localhost:8081/ready

# Or use Makefile for complete testing
make test-local
```

## Project Structure

- `main.go` - Main application file with Gin router setup
- `health.go` - Health check handlers for Kubernetes probes
- `Dockerfile` - Multi-stage Docker build configuration
- `Makefile` - Build automation and deployment commands
- `go.mod` / `go.sum` - Go module dependencies 