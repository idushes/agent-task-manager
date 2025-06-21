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
						Method:      "POST",
						Path:        "/generate-jwt",
						Description: "Генерация JWT токена для аутентификации (Rate limit: 5 запросов в минуту с одного IP)",
						Auth:        false,
						Request: map[string]interface{}{
							"body": map[string]interface{}{
								"secret":     "Секретный ключ сервиса (обязательный)",
								"user_id":    "ID пользователя (опциональный, по умолчанию 'anonymous')",
								"expires_in": "Время жизни токена в часах (опциональный, по умолчанию 8760)",
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
							{Code: 400, Description: "Неверный формат JSON или отсутствует обязательное поле secret"},
							{Code: 401, Description: "Неверный secret"},
							{Code: 429, Description: "Превышен лимит запросов (rate limit)"},
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
					{
						Method:      "GET",
						Path:        "/root-task",
						Description: "Получение списка корневых задач текущего пользователя (где id == root_task_id и created_by == текущий пользователь)",
						Auth:        true,
						Response: []map[string]interface{}{
							{
								"root_task_id": "123e4567-e89b-12d3-a456-426614174000",
								"created_at":   "2024-01-20T10:30:00Z",
								"delete_at":    "2024-04-20T10:30:00Z",
								"assignee":     "agent1",
								"description":  "Основная задача 1",
								"status":       "submitted",
							},
							{
								"root_task_id": "789a0123-e89b-12d3-a456-426614174002",
								"created_at":   "2024-01-21T14:00:00Z",
								"delete_at":    nil,
								"assignee":     "agent2",
								"description":  "Основная задача 2",
								"status":       "completed",
							},
						},
						Errors: []ErrorInfo{
							{Code: 401, Description: "Требуется авторизация"},
							{Code: 500, Description: "Ошибка при получении задач из базы данных"},
						},
					},
				},
				"Statistics": {
					{
						Method:      "GET",
						Path:        "/stat",
						Description: "Получение статистики по задачам пользователя за указанный период",
						Auth:        true,
						Request: map[string]interface{}{
							"query_params": map[string]string{
								"period": "Период для статистики (опциональный, по умолчанию 'all-time'). Возможные значения: today, yesterday, week, month, year, all-time",
							},
						},
						Response: map[string]interface{}{
							"period":        "week",
							"pending_tasks": 15,
							"in_progress":   3,
							"new_tasks":     25,
							"failed_tasks":  2,
							"_note":         "pending_tasks и in_progress показывают текущее состояние, new_tasks и failed_tasks считаются за указанный период",
						},
						Errors: []ErrorInfo{
							{Code: 400, Description: "Неверный параметр period"},
							{Code: 401, Description: "Требуется авторизация"},
							{Code: 500, Description: "Ошибка при подсчете статистики"},
						},
					},
				},
				"Users": {
					{
						Method:      "GET",
						Path:        "/users-with-tasks",
						Description: "Получение списка пользователей с активными задачами (из in-memory кэша)",
						Auth:        true,
						Request: map[string]interface{}{
							"query_params": map[string]string{
								"filter": "Список пользователей через запятую для фильтрации (опциональный). Пример: user1,user2,user3",
							},
						},
						Response: map[string]interface{}{
							"users": []string{"user1", "user2", "user3"},
							"count": 3,
							"_note": "Если параметр filter указан, возвращаются только пользователи из списка, которые имеют активные задачи",
						},
						Errors: []ErrorInfo{
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
						"11. In-memory кэш используется для хранения списка пользователей с активными задачами",
						"12. При старте приложения происходит синхронизация кэша с базой данных",
						"13. Кэш автоматически синхронизируется с БД каждые 10 минут (настраивается через CACHE_SYNC_INTERVAL)",
						"14. Автоматическая очистка задач с истекшим DeleteAt запускается каждый час (настраивается через CLEANUP_INTERVAL)",
					},
				},
			},
		}

		// Добавляем информацию о безопасности
		info.Endpoints["Security"] = []Endpoint{
			{
				Method:      "INFO",
				Path:        "",
				Description: "Механизмы безопасности",
				Auth:        false,
				Response: map[string]interface{}{
					"features": []string{
						"1. JWT токены для аутентификации с настраиваемым временем жизни",
						"2. Rate limiting на эндпоинте /generate-jwt (5 запросов в минуту с одного IP)",
						"3. Blacklist пользователей через переменную окружения BLACKLISTED_USERS (разделенные запятыми)",
						"4. Секретный ключ передается через тело POST запроса, а не через URL",
						"5. Проверка алгоритма подписи JWT для защиты от algorithm confusion атак",
						"6. Блокировка всех токенов заблокированного пользователя автоматически",
					},
					"environment_variables": map[string]string{
						"SECRET_KEY":          "Секретный ключ для подписи JWT токенов (обязательный)",
						"BLACKLISTED_USERS":   "Список заблокированных пользователей через запятую (опциональный)",
						"CACHE_SYNC_INTERVAL": "Интервал синхронизации кэша с БД (опциональный, по умолчанию 10m)",
					},
				},
			},
		}

		// Добавляем информацию о кэшировании
		info.Endpoints["In-Memory Cache"] = []Endpoint{
			{
				Method:      "INFO",
				Path:        "",
				Description: "In-memory кэширование пользователей с активными задачами",
				Auth:        false,
				Response: map[string]interface{}{
					"purpose":         "Быстрое хранение списка пользователей с активными задачами",
					"data_structure":  "Thread-safe map для хранения уникальных user_id",
					"sync_on_startup": "Автоматическая синхронизация с БД при старте приложения",
					"periodic_sync":   "Периодическая синхронизация каждые 10 минут (настраивается через CACHE_SYNC_INTERVAL)",
					"operations": []string{
						"1. Добавление пользователя при создании задачи в статусе 'submitted'",
						"2. Добавление пользователя когда родительская задача возвращается в 'submitted'",
						"3. Удаление пользователя когда у него не остается активных задач (при complete/cancel)",
						"4. Получение списка всех пользователей с активными задачами через GET /users-with-tasks",
						"5. Автоматическая полная синхронизация с БД по расписанию",
					},
					"benefits": []string{
						"Мгновенный доступ к списку пользователей",
						"Нет зависимости от внешних сервисов",
						"Автоматическая синхронизация при старте",
						"Периодическая синхронизация для актуальности данных",
						"Потокобезопасная реализация",
					},
				},
			},
		}

		c.JSON(http.StatusOK, info)
	}
}
