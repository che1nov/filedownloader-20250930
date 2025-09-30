package domain

import "time"

type Task struct {
	ID        string    `json:"id"`
	URLs      []string  `json:"urls"`
	Status    Status    `json:"status"`
	Files     []File    `json:"files"`
	CreatedAt time.Time `json:"created_at"`
	Progress  int       `json:"progress"`
}
