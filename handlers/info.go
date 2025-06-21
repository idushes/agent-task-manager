package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIInfo structure for API documentation
type APIInfo struct {
	Title       string                `json:"title"`
	Version     string                `json:"version"`
	Description string                `json:"description"`
	BaseURL     string                `json:"base_url"`
	Auth        AuthInfo              `json:"authentication"`
	Endpoints   map[string][]Endpoint `json:"endpoints"`
}

// AuthInfo authentication information
type AuthInfo struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

// Endpoint endpoint description
type Endpoint struct {
	Method      string      `json:"method"`
	Path        string      `json:"path"`
	Description string      `json:"description"`
	Auth        bool        `json:"requires_auth"`
	Request     interface{} `json:"request,omitempty"`
	Response    interface{} `json:"response,omitempty"`
	Errors      []ErrorInfo `json:"possible_errors,omitempty"`
}

// ErrorInfo error information
type ErrorInfo struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

// InfoHandler handler for getting API information
func InfoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		info := APIInfo{
			Title:       "Agent Task Manager API",
			Version:     "1.0.0",
			Description: "API for managing agent tasks with task hierarchy support and JWT authentication",
			BaseURL:     "https://task.agent.lisacorp.com",
			Auth: AuthInfo{
				Type:        "Bearer Token (JWT)",
				Description: "JWT token in Authorization header is required for protected endpoints",
				Example:     "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			},
			Endpoints: map[string][]Endpoint{
				"Health": {
					{
						Method:      "GET",
						Path:        "/health",
						Description: "Service health check",
						Auth:        false,
						Response: map[string]string{
							"status": "healthy",
						},
					},
					{
						Method:      "GET",
						Path:        "/ready",
						Description: "Service readiness check (including DB)",
						Auth:        false,
						Response: map[string]string{
							"status": "ready",
						},
					},
				},
				"Authentication": {
					{
						Method:      "POST",
						Path:        "/generate-jwt",
						Description: "Generate JWT token for authentication (Rate limit: 5 requests per minute per IP)",
						Auth:        false,
						Request: map[string]interface{}{
							"body": map[string]interface{}{
								"secret":     "Service secret key (required)",
								"user_id":    "User ID (optional, default 'anonymous')",
								"expires_in": "Token lifetime in hours (optional, default 8760)",
							},
							"example": map[string]interface{}{
								"secret":     "your-secret-key",
								"user_id":    "user123",
								"expires_in": 24,
							},
						},
						Response: map[string]interface{}{
							"token":      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
							"expires_at": 1735689600,
							"user_id":    "user123",
						},
						Errors: []ErrorInfo{
							{Code: 400, Description: "Invalid JSON format or missing required field 'secret'"},
							{Code: 401, Description: "Invalid secret"},
							{Code: 429, Description: "Rate limit exceeded"},
						},
					},
					{
						Method:      "GET",
						Path:        "/me",
						Description: "Get current user information",
						Auth:        true,
						Response: map[string]interface{}{
							"user_id":    "user123",
							"expires_at": 1735689600,
						},
						Errors: []ErrorInfo{
							{Code: 401, Description: "Missing or invalid token"},
						},
					},
				},
				"Tasks": {
					{
						Method:      "POST",
						Path:        "/task",
						Description: "Create new task",
						Auth:        true,
						Request: map[string]interface{}{
							"description":    "Task description (required)",
							"assignee":       "Assignee ID (optional)",
							"parent_task_id": "Parent task UUID (optional)",
							"delete_at":      "Task deletion date ISO 8601 (optional, default +3 months)",
							"credentials": map[string]interface{}{
								"service_name": map[string]string{
									"ENV_VAR": "value",
								},
								"example": map[string]interface{}{
									"postgres": map[string]string{
										"DB_PASSWORD": "secret123",
										"DB_HOST":     "localhost",
									},
									"redis": map[string]string{
										"REDIS_URL": "redis://localhost:6379",
									},
								},
							},
						},
						Response: map[string]interface{}{
							"id":             "123e4567-e89b-12d3-a456-426614174000",
							"created_at":     "2024-01-20T10:30:00Z",
							"delete_at":      "2024-04-20T10:30:00Z",
							"created_by":     "user123",
							"assignee":       "agent456",
							"description":    "Analyze sales data",
							"root_task_id":   "123e4567-e89b-12d3-a456-426614174000",
							"parent_task_id": nil,
							"result":         "",
							"credentials":    "{}",
							"status":         "submitted",
						},
						Errors: []ErrorInfo{
							{Code: 400, Description: "Invalid data format or parent task in invalid status"},
							{Code: 401, Description: "Authorization required"},
						},
					},
					{
						Method:      "GET",
						Path:        "/task",
						Description: "Get task for work (takes first available task where assignee = current user and status = submitted)",
						Auth:        true,
						Response: map[string]interface{}{
							"id":          "123e4567-e89b-12d3-a456-426614174000",
							"status":      "working",
							"description": "Analyze data",
							"_note":       "Status automatically changes to 'working'",
							"completed_subtasks": []map[string]interface{}{
								{
									"id":          "456e7890-e89b-12d3-a456-426614174001",
									"description": "Subtask 1",
									"status":      "completed",
									"result":      "Subtask completed successfully",
								},
							},
						},
						Errors: []ErrorInfo{
							{Code: 404, Description: "No available tasks for this user"},
							{Code: 401, Description: "Authorization required"},
						},
					},
					{
						Method:      "POST",
						Path:        "/task/:id/complete",
						Description: "Complete task (only assignee can complete task)",
						Auth:        true,
						Request: map[string]interface{}{
							"description": "Task execution result (required)",
							"delete_at":   "New deletion date ISO 8601 (optional)",
						},
						Response: map[string]interface{}{
							"id":     "123e4567-e89b-12d3-a456-426614174000",
							"status": "completed",
							"result": "Analysis completed. Sales growth is 15%",
							"_note":  "When completing a task, all active subtasks (submitted, working, waiting) are automatically canceled. If all subtasks are completed or canceled, parent task is moved to status = submitted",
						},
						Errors: []ErrorInfo{
							{Code: 400, Description: "Task not in 'working' status or invalid data format"},
							{Code: 403, Description: "Only assignee can complete task"},
							{Code: 404, Description: "Task not found"},
							{Code: 401, Description: "Authorization required"},
						},
					},
					{
						Method:      "POST",
						Path:        "/task/:id/cancel",
						Description: "Cancel task and all its subtasks (can be performed by assignee or creator)",
						Auth:        true,
						Request: map[string]interface{}{
							"_note": "Request body not required",
						},
						Response: map[string]interface{}{
							"id":     "123e4567-e89b-12d3-a456-426614174000",
							"status": "canceled",
							"_note":  "When canceling a task, all active subtasks (submitted, working, waiting) are recursively canceled. If all parent's subtasks are now completed or canceled, parent is moved to status = submitted",
						},
						Errors: []ErrorInfo{
							{Code: 400, Description: "Cannot cancel task with status completed, canceled or failed"},
							{Code: 403, Description: "Only assignee or creator can cancel task"},
							{Code: 404, Description: "Task not found"},
							{Code: 401, Description: "Authorization required"},
						},
					},
					{
						Method:      "POST",
						Path:        "/tasks/:id/fail",
						Description: "Mark task as failed (only assignee can fail task)",
						Auth:        true,
						Request: map[string]interface{}{
							"reason": "Failure reason (required)",
						},
						Response: map[string]interface{}{
							"id":     "123e4567-e89b-12d3-a456-426614174000",
							"status": "failed",
							"result": "FAILURE REASON: Could not connect to database",
							"_note":  "Parent task remains in waiting status and waits for completion of other subtasks",
						},
						Errors: []ErrorInfo{
							{Code: 400, Description: "Task not in 'working' status or invalid data format"},
							{Code: 403, Description: "Only assignee can fail task"},
							{Code: 404, Description: "Task not found"},
							{Code: 401, Description: "Authorization required"},
						},
					},
					{
						Method:      "GET",
						Path:        "/root-task/:id/tasks",
						Description: "Get all tasks by root_task_id (available only to root task creator)",
						Auth:        true,
						Response: []map[string]interface{}{
							{
								"id":             "123e4567-e89b-12d3-a456-426614174000",
								"created_at":     "2024-01-20T10:30:00Z",
								"created_by":     "user123",
								"assignee":       "agent1",
								"description":    "Main task",
								"root_task_id":   "123e4567-e89b-12d3-a456-426614174000",
								"parent_task_id": nil,
								"result":         "",
								"status":         "submitted",
								"_note":          "Credentials field excluded from output",
							},
							{
								"id":             "456e7890-e89b-12d3-a456-426614174001",
								"created_at":     "2024-01-20T10:35:00Z",
								"created_by":     "user123",
								"assignee":       "agent2",
								"description":    "Subtask",
								"root_task_id":   "123e4567-e89b-12d3-a456-426614174000",
								"parent_task_id": "123e4567-e89b-12d3-a456-426614174000",
								"result":         "",
								"status":         "working",
							},
						},
						Errors: []ErrorInfo{
							{Code: 400, Description: "Invalid ID format"},
							{Code: 403, Description: "Access denied: you are not the creator of the root task"},
							{Code: 404, Description: "Root task not found"},
							{Code: 401, Description: "Authorization required"},
						},
					},
					{
						Method:      "GET",
						Path:        "/root-task",
						Description: "Get list of current user's root tasks (where id == root_task_id and created_by == current user)",
						Auth:        true,
						Response: []map[string]interface{}{
							{
								"root_task_id": "123e4567-e89b-12d3-a456-426614174000",
								"created_at":   "2024-01-20T10:30:00Z",
								"delete_at":    "2024-04-20T10:30:00Z",
								"assignee":     "agent1",
								"description":  "Main task 1",
								"status":       "submitted",
							},
							{
								"root_task_id": "789a0123-e89b-12d3-a456-426614174002",
								"created_at":   "2024-01-21T14:00:00Z",
								"delete_at":    nil,
								"assignee":     "agent2",
								"description":  "Main task 2",
								"status":       "completed",
							},
						},
						Errors: []ErrorInfo{
							{Code: 401, Description: "Authorization required"},
							{Code: 500, Description: "Error retrieving tasks from database"},
						},
					},
				},
				"Statistics": {
					{
						Method:      "GET",
						Path:        "/stat",
						Description: "Get user task statistics for specified period",
						Auth:        true,
						Request: map[string]interface{}{
							"query_params": map[string]string{
								"period": "Statistics period (optional, default 'all-time'). Possible values: today, yesterday, week, month, year, all-time",
							},
						},
						Response: map[string]interface{}{
							"period":        "week",
							"pending_tasks": 15,
							"in_progress":   3,
							"new_tasks":     25,
							"failed_tasks":  2,
							"_note":         "pending_tasks and in_progress show current state, new_tasks and failed_tasks are counted for specified period",
						},
						Errors: []ErrorInfo{
							{Code: 400, Description: "Invalid period parameter"},
							{Code: 401, Description: "Authorization required"},
							{Code: 500, Description: "Error calculating statistics"},
						},
					},
				},
				"Users": {
					{
						Method:      "GET",
						Path:        "/users-with-tasks",
						Description: "Get list of users with active tasks (from in-memory cache)",
						Auth:        true,
						Request: map[string]interface{}{
							"query_params": map[string]string{
								"filter": "Comma-separated list of users for filtering (optional). Example: user1,user2,user3",
							},
						},
						Response: map[string]interface{}{
							"users": []string{"user1", "user2", "user3"},
							"count": 3,
							"_note": "If filter parameter is specified, only users from the list who have active tasks are returned",
						},
						Errors: []ErrorInfo{
							{Code: 401, Description: "Authorization required"},
						},
					},
				},
			},
		}

		// Add additional information about task statuses
		info.Endpoints["Task Statuses"] = []Endpoint{
			{
				Method:      "INFO",
				Path:        "",
				Description: "Possible task statuses",
				Auth:        false,
				Response: map[string]string{
					"submitted":      "Task created and waiting to be taken for work",
					"working":        "Task in progress",
					"waiting":        "Task waiting for subtasks completion",
					"completed":      "Task successfully completed",
					"failed":         "Task completed with error",
					"canceled":       "Task canceled",
					"rejected":       "Task rejected",
					"input-required": "Additional input required",
				},
			},
		}

		// Add information about business logic
		info.Endpoints["Business Logic"] = []Endpoint{
			{
				Method:      "INFO",
				Path:        "",
				Description: "Main rules for working with tasks",
				Auth:        false,
				Response: map[string]interface{}{
					"rules": []string{
						"1. When creating a subtask, parent task is automatically moved to 'waiting' status",
						"2. Subtasks can only be created for tasks in statuses: waiting, working, submitted",
						"3. When all subtasks are completed or canceled, parent task is automatically moved to 'submitted' status",
						"4. When completing or canceling a task, all its active subtasks (submitted, working, waiting) are recursively canceled",
						"5. Only assignee can take task for work, complete or fail it",
						"6. Assignee or task creator can cancel task",
						"7. Tasks are automatically deleted after 3 months (can be changed during creation)",
						"8. Each task has root_task_id for hierarchy tracking",
						"9. When getting task (GET /task), response includes completed first-level subtasks",
						"10. Only root task creator can view all tasks in its hierarchy (GET /root-task/:id/tasks)",
						"11. In-memory cache is used to store list of users with active tasks",
						"12. Cache is synchronized with database on application startup",
						"13. Cache is automatically synchronized with DB every 10 minutes (configurable via CACHE_SYNC_INTERVAL)",
						"14. Automatic cleanup of tasks with expired DeleteAt runs every hour (configurable via CLEANUP_INTERVAL)",
					},
				},
			},
		}

		// Add security information
		info.Endpoints["Security"] = []Endpoint{
			{
				Method:      "INFO",
				Path:        "",
				Description: "Security mechanisms",
				Auth:        false,
				Response: map[string]interface{}{
					"features": []string{
						"1. JWT tokens for authentication with configurable lifetime",
						"2. Rate limiting on /generate-jwt endpoint (5 requests per minute per IP)",
						"3. User blacklist via BLACKLISTED_USERS environment variable (comma-separated)",
						"4. Secret key passed through POST body, not URL",
						"5. JWT signature algorithm verification to protect against algorithm confusion attacks",
						"6. All tokens of blacklisted user are automatically blocked",
					},
					"environment_variables": map[string]string{
						"SECRET_KEY":          "Secret key for JWT token signing (required)",
						"BLACKLISTED_USERS":   "Comma-separated list of blacklisted users (optional)",
						"CACHE_SYNC_INTERVAL": "Cache synchronization interval with DB (optional, default 10m)",
					},
				},
			},
		}

		// Add caching information
		info.Endpoints["In-Memory Cache"] = []Endpoint{
			{
				Method:      "INFO",
				Path:        "",
				Description: "In-memory caching of users with active tasks",
				Auth:        false,
				Response: map[string]interface{}{
					"purpose":         "Fast storage of users list with active tasks",
					"data_structure":  "Thread-safe map for storing unique user_id",
					"sync_on_startup": "Automatic synchronization with DB on application startup",
					"periodic_sync":   "Periodic synchronization every 10 minutes (configurable via CACHE_SYNC_INTERVAL)",
					"operations": []string{
						"1. Add user when creating task in 'submitted' status",
						"2. Add user when parent task returns to 'submitted'",
						"3. Remove user when they have no active tasks left (on complete/cancel)",
						"4. Get list of all users with active tasks via GET /users-with-tasks",
						"5. Automatic full synchronization with DB on schedule",
					},
					"benefits": []string{
						"Instant access to users list",
						"No dependency on external services",
						"Automatic synchronization on startup",
						"Periodic synchronization for data freshness",
						"Thread-safe implementation",
					},
				},
			},
		}

		c.JSON(http.StatusOK, info)
	}
}
