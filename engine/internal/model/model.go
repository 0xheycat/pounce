// Package model holds the core data types shared across the Pounce engine.
package model

import "time"

// Status represents the lifecycle state of a download.
type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusPaused    Status = "paused"
	StatusCompleted Status = "completed"
	StatusError     Status = "error"
	StatusCanceled  Status = "canceled"
)

// Segment is one byte-range of a file, downloaded over its own connection.
// Downloaded tracks how many bytes of this segment are already on disk, which
// is what makes resume-after-restart possible.
type Segment struct {
	Index      int   `json:"index"`
	Start      int64 `json:"start"`
	End        int64 `json:"end"` // -1 when the total size is unknown
	Downloaded int64 `json:"downloaded"`
}

// Download is a single managed download. It is persisted to disk so that the
// engine can resume it across restarts.
type Download struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Filename    string    `json:"filename"`
	Dir         string    `json:"dir"`
	TotalSize   int64     `json:"totalSize"`
	Downloaded  int64     `json:"downloaded"`
	Status      Status    `json:"status"`
	Connections int       `json:"connections"`
	Resumable   bool      `json:"resumable"`
	Segments    []Segment `json:"segments"`
	SpeedLimit  int64     `json:"speedLimit"` // bytes/sec, 0 = unlimited
	Error       string    `json:"error,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Speed is a runtime-only estimate of the current throughput in bytes/sec.
	Speed int64 `json:"speed"`
}
