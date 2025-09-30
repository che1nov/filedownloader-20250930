package domain

type CreateTaskRequest struct {
	URLs []string `json:"urls"`
}

type CreateTaskResponse struct {
	TaskID string `json:"task_id"`
}

type TaskStatusResponse struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
	Files    []File `json:"files"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
