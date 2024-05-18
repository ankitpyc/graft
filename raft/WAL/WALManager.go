package wal

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"os"
	"sync"
	"sync/atomic"
)

const (
	CommandBsize   = 3
	TimestampBsize = 8
	MaxWlogSize    = 4096
)
const (
	TypeJSON byte = iota + 1
	TypeString
	TypeInt
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

func (walManager *WALManager) IncrementLatestCommitedIndex() uint64 {
	return atomic.AddUint64(&walManager.LatestCommitIndex, 1)
}

func (walManager *WALManager) AppendLog(entry WALLogEntry) {

	_, err := encodeWALEntry(walManager, entry)
	if err != nil {
	}
	return
}

func encodeWALEntry(walManager *WALManager, entry WALLogEntry) ([]byte, error) {
	// Initialize buffer with the maximum size

	// Encode timestamp
	// Create a buffered writer
	writer := bufio.NewWriter(walManager.Fd)

	// Write data type byte
	_, err := writer.Write([]byte{0})
	if err != nil {
		return nil, err
	}

	// Convert and write timestamp field to bytes
	timestampBytes := make([]byte, 8) // 8 bytes for uint64
	binary.BigEndian.PutUint64(timestampBytes, entry.Timestamp)
	_, err = writer.Write(timestampBytes)
	if err != nil {
		return nil, err
	}

	// Convert and write log index field to bytes
	logIndexBytes := make([]byte, 8) // 8 bytes for uint64
	binary.BigEndian.PutUint64(logIndexBytes, entry.LogIndex)
	_, err = writer.Write(logIndexBytes)
	if err != nil {
		return nil, err
	}

	// Write key field as bytes
	_, err = writer.WriteString(entry.Key)
	if err != nil {
		return nil, err
	}

	// Convert and write value field to bytes based on data type
	switch entry.Value.(type) {
	case string:
		_, err = writer.Write([]byte{0})
		_, err := writer.Write([]byte(entry.Value.(string)))
		if err != nil {
			return nil, err
		}
	default:
		_, err = writer.Write([]byte{2})
	}

	// Write a newline to separate entries if needed
	_, err = writer.WriteString("\n")

	// Flush the buffer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (walManager *WALManager) encodeTimeStamp(entry WALLogEntry) ([]byte, error) {
	buffer := new(bytes.Buffer)
	if err := binary.Write(buffer, binary.BigEndian, entry.Timestamp); err != nil {
		return nil, err
	}
	fullBytes := buffer.Bytes()
	if len(fullBytes) > 3 {
		fullBytes = fullBytes[len(fullBytes)-3:]
	}
	return fullBytes, nil
}
