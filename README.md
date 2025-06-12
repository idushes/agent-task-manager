# Agent Task Manager

A hierarchical task management API service built with Go, Gin framework, and GORM ORM. Features JWT authentication, task lifecycle management, and full Kubernetes readiness with health check endpoints.

## Key Features

- 🔐 **JWT Authentication** - Secure API access with configurable token expiration
- 🌳 **Hierarchical Tasks** - Support for parent-child task relationships
- 🔄 **Task Lifecycle Management** - Multiple statuses: submitted, working, waiting, completed, failed, canceled
- 🚀 **Auto Status Transitions** - Smart status updates based on subtask completion
- 🗑️ **Auto Cleanup** - Tasks automatically deleted after configurable period
- 🔒 **Role-based Access** - Different permissions for assignee and task creator
- 🏥 **Kubernetes Ready** - Built-in health and readiness probes
- 🌍 **Multi-platform Docker** - Supports linux/amd64 and linux/arm64

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

## Quick API Example

```bash
# 1. Generate JWT token
TOKEN=$(curl -s -X POST "http://localhost:8081/generate-jwt" \
  -H "Content-Type: application/json" \
  -d '{"secret":"your-secret-key","user_id":"agent1"}' | jq -r .token)

# 2. Create a task
curl -X POST http://localhost:8081/task \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Analyze sales data for Q1",
    "assignee": "agent1"
  }'

# 3. Get next task (automatically sets status to "working")
curl http://localhost:8081/task \
  -H "Authorization: Bearer $TOKEN"

# 4. Complete the task
curl -X POST http://localhost:8081/task/{task-id}/complete \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Analysis complete. Sales increased by 15%"
  }'
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

### Running with PostgreSQL

The service requires PostgreSQL database. You can run it locally with Docker:

```bash
# Start PostgreSQL with Docker
docker run -d \
  --name postgres-task-manager \
  -e POSTGRES_USER=taskuser \
  -e POSTGRES_PASSWORD=taskpass \
  -e POSTGRES_DB=taskdb \
  -p 5432:5432 \
  postgres:15

# Set environment variables
export POSTGRES_URL="postgres://taskuser:taskpass@localhost:5432/taskdb?sslmode=disable"
export SECRET_KEY="your-secure-secret-key"

# Run the service
make run
```

### Using Docker Compose (Alternative)

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: taskuser
      POSTGRES_PASSWORD: taskpass
      POSTGRES_DB: taskdb
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  app:
    build: .
    ports:
      - "8081:8081"
    environment:
      POSTGRES_URL: postgres://taskuser:taskpass@postgres:5432/taskdb?sslmode=disable
      SECRET_KEY: your-secure-secret-key
    depends_on:
      - postgres

volumes:
  postgres_data:
```

Then run: `docker-compose up`

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

### Health & Status

#### Health Check
- **GET** `/health` - Service liveness check (liveness probe)
  ```json
  {
    "status": "healthy"
  }
  ```

#### Ready Check
- **GET** `/ready` - Service readiness check (readiness probe)
  ```json
  {
    "status": "ready"
  }
  ```

#### API Info
- **GET** `/info` - Get detailed API documentation
  - Returns comprehensive API documentation with all endpoints

### Authentication

#### Generate JWT Token
- **POST** `/generate-jwt`
  - Generates JWT token for authentication
  - Rate limit: 5 requests per minute per IP
  - Request body:
    ```json
    {
      "secret": "your-secret-key",
      "user_id": "user123",
      "expires_in": 24
    }
    ```
  - Parameters:
    - `secret` (required) - Must match server's SECRET_KEY
    - `user_id` (optional) - Default: "anonymous"
    - `expires_in` (optional) - Token lifetime in hours, default: 8760 (1 year)

#### Get Current User
- **GET** `/me` - Get current user info (requires auth)
  - Headers: `Authorization: Bearer {token}`

### Task Management (Requires Authentication)

All task endpoints require JWT authentication via `Authorization: Bearer {token}` header.

#### Create Task
- **POST** `/task` - Create a new task
  ```json
  {
    "description": "Task description",
    "assignee": "user123",
    "parent_task_id": "uuid-of-parent-task",
    "delete_at": "2024-04-20T10:30:00Z",
    "credentials": {
      "service_name": {
        "ENV_VAR": "value"
      }
    }
  }
  ```

#### Get Next Task
- **GET** `/task` - Get next available task for current user
  - Returns first task where assignee = current user and status = "submitted"
  - Automatically changes task status to "working"
  - Includes completed first-level subtasks in the response
  ```json
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "status": "working",
    "description": "Main task description",
    "completed_subtasks": [
      {
        "id": "456e7890-e89b-12d3-a456-426614174001",
        "description": "Subtask 1",
        "status": "completed",
        "result": "Subtask completed successfully"
      }
    ]
  }
  ```

#### Complete Task
- **POST** `/task/:id/complete` - Mark task as completed
  ```json
  {
    "description": "Result of the task",
    "delete_at": "2024-04-20T10:30:00Z"
  }
  ```
  - Only assignee can complete the task
  - Cancels all active subtasks recursively
  - Updates parent task status if all subtasks are done

#### Cancel Task
- **POST** `/task/:id/cancel` - Cancel task and all subtasks
  - No request body required
  - Can be done by assignee or task creator
  - Recursively cancels all active subtasks
  - Updates parent task status if all subtasks are done

#### Fail Task
- **POST** `/tasks/:id/fail` - Mark task as failed
  ```json
  {
    "reason": "Reason for failure"
  }
  ```
  - Only assignee can fail the task
  - Sets result to "FAILURE REASON: {reason}"
  - Parent task remains in "waiting" status

#### Get Root Tasks
- **GET** `/root-task/:id/tasks` - Get all tasks by root_task_id
  - Returns flat list of all tasks with specified root_task_id
  - Access control: Only the creator of the root task can access this endpoint
  - Credentials field is excluded from the response
  ```json
  [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "created_at": "2024-01-20T10:30:00Z",
      "created_by": "user123",
      "assignee": "agent1",
      "description": "Main task",
      "root_task_id": "123e4567-e89b-12d3-a456-426614174000",
      "parent_task_id": null,
      "result": "",
      "status": "submitted"
    },
    {
      "id": "456e7890-e89b-12d3-a456-426614174001",
      "created_at": "2024-01-20T10:35:00Z",
      "created_by": "user123",
      "assignee": "agent2",
      "description": "Subtask",
      "root_task_id": "123e4567-e89b-12d3-a456-426614174000",
      "parent_task_id": "123e4567-e89b-12d3-a456-426614174000",
      "result": "",
      "status": "working"
    }
  ]
  ```

## Task Lifecycle & Business Logic

### Task Statuses
- `submitted` - Task created and waiting to be taken
- `working` - Task in progress by assignee
- `waiting` - Task waiting for subtasks to complete
- `completed` - Task successfully completed
- `failed` - Task failed with error
- `canceled` - Task was canceled
- `rejected` - Task was rejected
- `input-required` - Task requires additional input

### Business Rules
1. When creating a subtask, parent task automatically transitions to `waiting` status
2. Subtasks can only be created for tasks in statuses: `waiting`, `working`, `submitted`
3. When all subtasks are `completed` or `canceled`, parent task transitions to `submitted`
4. When completing or canceling a task, all active subtasks (`submitted`, `working`, `waiting`) are recursively canceled
5. Only assignee can take task to work, complete or fail it
6. Assignee or task creator can cancel a task
7. Tasks are automatically deleted after 3 months (configurable via `delete_at`)
8. Each task has `root_task_id` for hierarchy tracking
9. When getting a task (GET /task), completed first-level subtasks are included in the response
10. Only the creator of a root task can view all tasks in its hierarchy (GET /root-task/:id/tasks)

### Task Hierarchy Example
```
Root Task A
├── Subtask B (assignee: agent1)
│   └── Subtask D (assignee: agent2)
└── Subtask C (assignee: agent3)
    └── Subtask E (assignee: agent4)
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

## Configuration

The application supports configuration through environment variables or a `.env` file. Environment variables take precedence over `.env` file values.

### Using .env file

1. Copy the example configuration:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` file with your values:
   ```bash
   # Required
   SECRET_KEY=your-secure-secret-key-here
   
   # Optional
   PORT=8081
   POSTGRES_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
   REDIS_URL=redis://localhost:6379
   ```

3. Run the application:
   ```bash
   go run .
   ```

### Configuration Priority

1. Environment variables (highest priority)
2. `.env` file values
3. Default values (lowest priority)

### Required Configuration

- `SECRET_KEY` - **Required** for JWT token signing. Application will not start without it.
- `POSTGRES_URL` - **Required** for database connection. Application will not start without it.

### Database Configuration

The application requires PostgreSQL:
- `POSTGRES_URL` is required for database connection
- Database tables are automatically migrated on startup

#### Database Schema

The service automatically creates and migrates the following table:

```sql
-- tasks table
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    delete_at TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    assignee VARCHAR(255),
    description TEXT,
    root_task_id UUID,
    parent_task_id UUID,
    result TEXT,
    credentials JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'submitted',
    FOREIGN KEY (root_task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_task_id) REFERENCES tasks(id) ON DELETE CASCADE
);
```

## Environment Variables

### Application Configuration
- `SECRET_KEY` - **Required** - Secret key for JWT token signing
- `POSTGRES_URL` - **Required** - PostgreSQL connection URL
- `PORT` - Port to run the server on (default: 8081)
- `REDIS_URL` - Redis connection URL (optional)
- `BLACKLISTED_USERS` - Comma-separated list of blocked user IDs (optional)
- `ALLOWED_ORIGINS` - Comma-separated list of allowed CORS origins (default: "*")

### Build/Deployment Configuration
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

### Testing Task Hierarchy

Example of creating and managing a task hierarchy:

```bash
# Get token
TOKEN=$(curl -s -X POST "http://localhost:8081/generate-jwt" \
  -H "Content-Type: application/json" \
  -d '{"secret":"your-secret-key","user_id":"manager1"}' | jq -r .token)

# 1. Create root task
ROOT_TASK=$(curl -s -X POST http://localhost:8081/task \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Complete project X",
    "assignee": "manager1"
  }' | jq -r .id)

# 2. Create subtasks
SUBTASK1=$(curl -s -X POST http://localhost:8081/task \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Design database schema",
    "assignee": "dev1",
    "parent_task_id": "'$ROOT_TASK'"
  }' | jq -r .id)

# 3. Get token for dev1 and work on task
TOKEN_DEV1=$(curl -s -X POST "http://localhost:8081/generate-jwt" \
  -H "Content-Type: application/json" \
  -d '{"secret":"your-secret-key","user_id":"dev1"}' | jq -r .token)

# 4. Dev1 gets their task (changes status to working)
curl -H "Authorization: Bearer $TOKEN_DEV1" http://localhost:8081/task

# 5. Dev1 completes the task
curl -X POST http://localhost:8081/task/$SUBTASK1/complete \
  -H "Authorization: Bearer $TOKEN_DEV1" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Schema designed with 5 tables"
  }'

# 6. Manager can view all tasks in the hierarchy
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8081/root-task/$ROOT_TASK/tasks | jq
```

## Project Structure

- `main.go` - Main application file with Gin router setup
- `config/config.go` - Configuration management
- `database/database.go` - Database connection and initialization
- `handlers/`
  - `health.go` - Health check handlers for Kubernetes probes
  - `jwt_auth.go` - JWT authentication middleware and handlers
  - `info.go` - API documentation endpoint
  - `tasks/` - Task management handlers
    - `create.go` - Create task handler
    - `get.go` - Get next task handler
    - `get_root_tasks.go` - Get all tasks by root_task_id handler
    - `complete.go` - Complete task handler
    - `cancel.go` - Cancel task handler
    - `fail.go` - Fail task handler
    - `types.go` - Request/response types
    - `validation.go` - Input validation
- `models/task.go` - Task model with GORM definitions (supports cascade deletion)
- `Dockerfile` - Multi-stage Docker build configuration
- `Makefile` - Build automation and deployment commands
- `go.mod` / `go.sum` - Go module dependencies 