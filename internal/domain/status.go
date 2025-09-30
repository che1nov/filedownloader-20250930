package domain

type Status string

const (
	StatusPending     Status = "pending"
	StatusDownloading Status = "downloading"
	StatusCompleted   Status = "completed"
	StatusFailed      Status = "failed"
)
