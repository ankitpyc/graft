package wal

import "cache/internal/domain"

type Command int

// Declare constants using iota
const (
	SET Command = iota
	GET
	DELETE
)

type WALLogEntry struct {
	Comm  Command    `json:"Command"`
	Key   string     `json:"Key"`
	Value domain.Key `json:"Value"`
}

func NewWALLogEntry(comm Command, key string, value domain.Key) WALLogEntry {
	return WALLogEntry{
		Comm:  comm,
		Key:   key,
		Value: value,
	}
}
