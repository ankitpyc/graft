package wal

import (
	"encoding/json"
	"os"
)

type WALManager struct {
	Fd        *os.File
	Log       []WALLogEntry
	LogStream chan []WALLogEntry
}

func NewWALManager(filename string) *WALManager {
	fd, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	walManager := &WALManager{
		Fd:        fd,
		Log:       []WALLogEntry{},
		LogStream: make(chan []WALLogEntry),
	}
	go walManager.LogListener()
	return walManager

}

func (wal *WALManager) LogListener() {
	for {
		select {
		case lo :=
			<-wal.LogStream:
			for _, entry := range lo {
				wal.AppendLog(entry)
			}
		}
	}
}

func (walManager *WALManager) AppendLog(walLogEntry WALLogEntry) {
	err := json.NewEncoder(walManager.Fd).Encode(walLogEntry)
	if err != nil {

	}
}
