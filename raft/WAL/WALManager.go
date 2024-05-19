package wal

import (
	"os"
	"sync"
)

type WALManager struct {
	Fd                *os.File
	Log               []WALLogEntry
	LogStream         chan []WALLogEntry
	LatestCommitIndex uint64
	mu                sync.Mutex
}

func NewWALManager(filename string) *WALManager {
	fd, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	walManager := &WALManager{
		Fd:                fd,
		Log:               []WALLogEntry{},
		LogStream:         make(chan []WALLogEntry),
		mu:                sync.Mutex{},
		LatestCommitIndex: 0,
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
				wal.mu.Lock()
				wal.AppendLog(entry)
				wal.mu.Unlock()
			}
		}
	}
}

func (walManager *WALManager) AppendLog(entry WALLogEntry) *WALLogEntry {

	commitedEntry, err := encodeWALEntry(walManager, &entry)
	walManager.LatestCommitIndex = commitedEntry.LogIndex
	if err != nil {
		return &entry
	}
	return commitedEntry
}
