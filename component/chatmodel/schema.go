package chatmodel

import "time"

type Status struct {
	Available bool
	Model     string
	Latency   time.Duration
	CheckedAt time.Time
}
