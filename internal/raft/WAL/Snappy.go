package wal

import "time"

type Snappy struct {
	TimeInterval time.Duration
	Frequency    uint8
	Path         string
}
