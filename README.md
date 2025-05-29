# Agent Task Manager

A simple Go service with health check endpoints for Kubernetes.

## Installation and Setup

```bash
# Install dependencies
go mod download

# Run the service
go run .
```

The service will start on port 8081.

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
        image: agent-task-manager:latest
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

## Docker Build

```dockerfile
FROM golang:1.24.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o agent-task-manager .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/agent-task-manager .
EXPOSE 8081
CMD ["./agent-task-manager"]
```

## Testing the Service

```bash
# Test health endpoint
curl http://localhost:8081/health

# Test ready endpoint
curl http://localhost:8081/ready
```

## Project Structure

- `main.go` - Main application file with Gin router setup
- `health.go` - Health check handlers for Kubernetes probes
- `Dockerfile` - Multi-stage Docker build configuration
- `go.mod` / `go.sum` - Go module dependencies 