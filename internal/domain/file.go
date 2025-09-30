package domain

import "time"

type File struct {
	URL        string    `json:"url"`
	Filename   string    `json:"filename"`
	Status     Status    `json:"status"`
	Size       int64     `json:"size"`
	Downloaded int64     `json:"downloaded"`
	CreatedAt  time.Time `json:"created_at"`
}
