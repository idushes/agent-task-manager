package tasks

import (
	"agent-task-manager/models"
	"encoding/json"
)

// validateCredentials проверяет структуру credentials
func validateCredentials(credentials json.RawMessage) (map[string]map[string]string, error) {
	if credentials == nil || len(credentials) == 0 {
		return nil, nil
	}

	var credsMap map[string]map[string]string
	if err := json.Unmarshal(credentials, &credsMap); err != nil {
		return nil, err
	}

	return credsMap, nil
}

// isParentStatusAllowed проверяет, разрешен ли статус родительской задачи
func isParentStatusAllowed(status models.TaskStatus) bool {
	allowedStatuses := []models.TaskStatus{
		models.StatusWaiting,
		models.StatusWorking,
		models.StatusSubmitted,
	}

	for _, allowed := range allowedStatuses {
		if status == allowed {
			return true
		}
	}

	return false
}
