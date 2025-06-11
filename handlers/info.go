package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIInfo структура для документации API
type APIInfo struct {
	Title       string                `json:"title"`
	Version     string                `json:"version"`
	Description string                `json:"description"`
	BaseURL     string                `json:"base_url"`
	Auth        AuthInfo              `json:"authentication"`
	Endpoints   map[string][]Endpoint `json:"endpoints"`
}

// AuthInfo информация об аутентификации
type AuthInfo struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

// Endpoint описание эндпоинта
type Endpoint struct {
	Method      string      `json:"method"`
	Path        string      `json:"path"`
	Description string      `json:"description"`
	Auth        bool        `json:"requires_auth"`
	Request     interface{} `json:"request,omitempty"`
	Response    interface{} `json:"response,omitempty"`
	Errors      []ErrorInfo `json:"possible_errors,omitempty"`
}

// ErrorInfo информация об ошибке
type ErrorInfo struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

// InfoHandler обработчик для получения информации об API
func InfoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		info := APIInfo{
			Title:       "Agent Task Manager API",
			Version:     "1.0.0",
			Description: "API для управления задачами агентов с поддержкой иерархии задач и аутентификации через JWT",
			BaseURL:     "https://task.agent.lisacorp.com",
			Auth: AuthInfo{
				Type:        "Bearer Token (JWT)",
				Description: "Для защищенных эндпоинтов требуется JWT токен в заголовке Authorization",
				Example:     "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			},
			Endpoints: map[string][]Endpoint{
				"Health": {
					{
						Method:      "GET",
						Path:        "/health",
						Description: "Проверка работоспособности сервиса",
						Auth:        false,
						Response: map[string]string{
							"status": "healthy",
						},
					},
					{
						Method:      "GET",
						Path:        "/ready",
						Description: "Проверка готовности сервиса (включая БД)",
						Auth:        false,
						Response: map[string]string{
							"status": "ready",
						},
					},
				},
				"Authentication": {
					{
						Method:      "GET",
						Path:        "/generate-jwt",
						Description: "Генерация JWT токена для аутентификации",
						Auth:        false,
						Request: map[string]interface{}{
							"query_params": map[string]string{
								"secret":     "Секретный ключ сервиса (обязательный)",
								"user_id":    "ID пользователя (по умолчанию 'anonymous')",
								"expires_in": "Время жизни токена в часах (по умолчанию 8760)",
							},
						},
						Response: map[string]interface{}{
							"token":      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
							"expires_at": 1735689600,
							"user_id":    "user123",
						},
						Errors: []ErrorInfo{
							{Code: 400, Description: "Не указан secret или неверный формат параметров"},
							{Code: 401, Description: "Неверный secret"},
						},
					},
					{
						Method:      "GET",
						Path:        "/me",
						Description: "Получение информации о текущем пользователе",
						Auth:        true,
						Response: map[string]interface{}{
							"user_id":    "user123",
							"expires_at": 1735689600,
						},
						Errors: []ErrorInfo{
							{Code: 401, Description: "Отсутствует или невалидный токен"},
						},
					},
				},
				"Tasks": {
					{
						Method:      "POST",
						Path:        "/task",
						Description: "Создание новой задачи",
						Auth:        true,
						Request: map[string]interface{}{
							"description":    "Описание задачи (обязательный)",
							"assignee":       "ID исполнителя (опциональный)",
							"parent_task_id": "UUID родительской задачи (опциональный)",
							"delete_at":      "Дата удаления задачи ISO 8601 (опциональный, по умолчанию +3 месяца)",
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
							{Code: 400, Description: "Неверный формат данных или родительская задача в недопустимом статусе"},
							{Code: 401, Description: "Требуется авторизация"},
						},
					},
					{
						Method:      "GET",
						Path:        "/task",
						Description: "Получение задачи в работу (берется первая доступная задача где assignee = текущий пользователь и status = submitted)",
						Auth:        true,
						Response: map[string]interface{}{
							"id":          "123e4567-e89b-12d3-a456-426614174000",
							"status":      "working",
							"description": "Проанализировать данные",
							"_note":       "Статус автоматически меняется на 'working'",
							"completed_subtasks": []map[string]interface{}{
								{
									"id":          "456e7890-e89b-12d3-a456-426614174001",
									"description": "Подзадача 1",
									"status":      "completed",
									"result":      "Подзадача выполнена успешно",
								},
							},
						},
						Errors: []ErrorInfo{
							{Code: 404, Description: "Нет доступных задач для данного пользователя"},
							{Code: 401, Description: "Требуется авторизация"},
						},
					},
					{
						Method:      "POST",
						Path:        "/task/:id/complete",
						Description: "Завершение задачи (только assignee может завершить задачу)",
						Auth:        true,
						Request: map[string]interface{}{
							"description": "Результат выполнения задачи (обязательный)",
							"delete_at":   "Новая дата удаления ISO 8601 (опциональный)",
						},
						Response: map[string]interface{}{
							"id":     "123e4567-e89b-12d3-a456-426614174000",
							"status": "completed",
							"result": "Анализ завершен. Рост продаж составил 15%",
							"_note":  "При завершении задачи все активные подзадачи (submitted, working, waiting) автоматически отменяются. Если все подзадачи завершены (completed) или отменены (canceled), родительская задача переводится в status = submitted",
						},
						Errors: []ErrorInfo{
							{Code: 400, Description: "Задача не в статусе 'working' или неверный формат данных"},
							{Code: 403, Description: "Только assignee может завершить задачу"},
							{Code: 404, Description: "Задача не найдена"},
							{Code: 401, Description: "Требуется авторизация"},
						},
					},
					{
						Method:      "POST",
						Path:        "/task/:id/cancel",
						Description: "Отмена задачи и всех её подзадач (может выполнить assignee или создатель)",
						Auth:        true,
						Request: map[string]interface{}{
							"_note": "Тело запроса не требуется",
						},
						Response: map[string]interface{}{
							"id":     "123e4567-e89b-12d3-a456-426614174000",
							"status": "canceled",
							"_note":  "При отмене задачи все активные подзадачи (submitted, working, waiting) рекурсивно отменяются. Если все подзадачи родителя теперь completed или canceled, родитель переводится в status = submitted",
						},
						Errors: []ErrorInfo{
							{Code: 400, Description: "Нельзя отменить задачу со статусом completed, canceled или failed"},
							{Code: 403, Description: "Только assignee или создатель может отменить задачу"},
							{Code: 404, Description: "Задача не найдена"},
							{Code: 401, Description: "Требуется авторизация"},
						},
					},
					{
						Method:      "POST",
						Path:        "/tasks/:id/fail",
						Description: "Пометка задачи как неудачной (только assignee может провалить задачу)",
						Auth:        true,
						Request: map[string]interface{}{
							"reason": "Причина неудачи (обязательный)",
						},
						Response: map[string]interface{}{
							"id":     "123e4567-e89b-12d3-a456-426614174000",
							"status": "failed",
							"result": "FAILURE REASON: Не удалось подключиться к базе данных",
							"_note":  "Родительская задача остается в статусе waiting и ждет завершения остальных подзадач",
						},
						Errors: []ErrorInfo{
							{Code: 400, Description: "Задача не в статусе 'working' или неверный формат данных"},
							{Code: 403, Description: "Только assignee может провалить задачу"},
							{Code: 404, Description: "Задача не найдена"},
							{Code: 401, Description: "Требуется авторизация"},
						},
					},
					{
						Method:      "GET",
						Path:        "/root-task/:id/tasks",
						Description: "Получение всех задач по root_task_id (доступно только создателю root задачи)",
						Auth:        true,
						Response: []map[string]interface{}{
							{
								"id":             "123e4567-e89b-12d3-a456-426614174000",
								"created_at":     "2024-01-20T10:30:00Z",
								"created_by":     "user123",
								"assignee":       "agent1",
								"description":    "Основная задача",
								"root_task_id":   "123e4567-e89b-12d3-a456-426614174000",
								"parent_task_id": nil,
								"result":         "",
								"status":         "submitted",
								"_note":          "Поле credentials исключено из вывода",
							},
							{
								"id":             "456e7890-e89b-12d3-a456-426614174001",
								"created_at":     "2024-01-20T10:35:00Z",
								"created_by":     "user123",
								"assignee":       "agent2",
								"description":    "Подзадача",
								"root_task_id":   "123e4567-e89b-12d3-a456-426614174000",
								"parent_task_id": "123e4567-e89b-12d3-a456-426614174000",
								"result":         "",
								"status":         "working",
							},
						},
						Errors: []ErrorInfo{
							{Code: 400, Description: "Неверный формат ID"},
							{Code: 403, Description: "Доступ запрещен: вы не являетесь создателем root задачи"},
							{Code: 404, Description: "Root задача не найдена"},
							{Code: 401, Description: "Требуется авторизация"},
						},
					},
				},
			},
		}

		// Добавляем дополнительную информацию о статусах
		info.Endpoints["Task Statuses"] = []Endpoint{
			{
				Method:      "INFO",
				Path:        "",
				Description: "Возможные статусы задач",
				Auth:        false,
				Response: map[string]string{
					"submitted":      "Задача создана и ожидает взятия в работу",
					"working":        "Задача в процессе выполнения",
					"waiting":        "Задача ожидает завершения подзадач",
					"completed":      "Задача успешно выполнена",
					"failed":         "Задача завершилась с ошибкой",
					"canceled":       "Задача отменена",
					"rejected":       "Задача отклонена",
					"input-required": "Требуется дополнительный ввод",
				},
			},
		}

		// Добавляем информацию о бизнес-логике
		info.Endpoints["Business Logic"] = []Endpoint{
			{
				Method:      "INFO",
				Path:        "",
				Description: "Основные правила работы с задачами",
				Auth:        false,
				Response: map[string]interface{}{
					"rules": []string{
						"1. При создании подзадачи родительская задача автоматически переводится в статус 'waiting'",
						"2. Подзадачи можно создавать только для задач в статусах: waiting, working, submitted",
						"3. При завершении (completed) или отмене (canceled) всех подзадач родительская задача автоматически переводится в статус 'submitted'",
						"4. При завершении или отмене задачи все её активные подзадачи (submitted, working, waiting) рекурсивно отменяются",
						"5. Только assignee может взять задачу в работу, завершить или провалить её",
						"6. Assignee или создатель задачи может отменить задачу",
						"7. Задачи автоматически удаляются через 3 месяца (можно изменить при создании)",
						"8. Каждая задача имеет root_task_id для отслеживания иерархии",
						"9. При получении задачи (GET /task) в ответ включаются завершенные подзадачи первого уровня",
						"10. Только создатель root задачи может просматривать все задачи в её иерархии (GET /root-task/:id/tasks)",
					},
				},
			},
		}

		c.JSON(http.StatusOK, info)
	}
}
