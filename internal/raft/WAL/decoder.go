package wal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
)

func decodeWALEntry(walManager *Manager) ([]*WALLogEntry, error) {
	buf := make([]byte, 1)
	var bytesRead []byte
	logs := make([]*WALLogEntry, 0)
	walManager.Fd.Seek(0, 0)
	for {
		_, err := walManager.Fd.Read(buf)
		if err != nil {
			log.Fatal("received error reading from file: ", err)
		}
		bytesRead = append(bytesRead, buf...)
		if buf[0] == '\n' {
			wal, er_ := decodeBytesToWal(bytesRead)
			logs = append(logs, wal)
			if err != nil {
				return nil, fmt.Errorf("error decoding WAL entry: %v", er_)
			}
			bytesRead = []byte{}
		}
	}
	return logs, nil
}

// decodeBytesToWal decodes a byte slice into a WALLogEntry
// Parameters:
// - read: The byte slice to decode
// Returns:
// - *WALLogEntry: Pointer to the decoded WALLogEntry
// - error: Error if any occurs during decoding
func decodeBytesToWal(read []byte) (*WALLogEntry, error) {
	var timeStamp, logIndex uint64
	var encodingType uint8
	var keyLen, valLen uint8
	bi := uint8(0)

	// Read timestamp (8 bytes)
	err := binary.Read(bytes.NewReader(read[bi:bi+TimestampBSize]), binary.BigEndian, &timeStamp)
	if err != nil {
		return nil, fmt.Errorf("error reading timestamp: %v", err)
	}
	bi += TimestampBSize

	// Read log index (8 bytes)
	err = binary.Read(bytes.NewReader(read[bi:bi+LogIndexBSize]), binary.BigEndian, &logIndex)
	if err != nil {
		return nil, fmt.Errorf("error reading log index: %v", err)
	}
	bi += LogIndexBSize

	// Read key length (1 byte)
	err = binary.Read(bytes.NewReader(read[bi:bi+KlenBSize]), binary.BigEndian, &keyLen)
	if err != nil {
		return nil, fmt.Errorf("error reading key length: %v", err)
	}
	bi += KlenBSize

	// Read key (keyLen bytes)
	keyBytes := make([]byte, keyLen)
	err = binary.Read(bytes.NewReader(read[bi:bi+keyLen]), binary.BigEndian, &keyBytes)
	if err != nil {
		return nil, fmt.Errorf("error reading key: %v", err)
	}
	bi += keyLen

	// Read encoding type (1 byte)
	err = binary.Read(bytes.NewReader(read[bi:bi+1]), binary.BigEndian, &encodingType)
	if err != nil {
		return nil, fmt.Errorf("error reading encoding type: %v", err)
	}
	bi += 1

	// Read value length (1 byte)
	err = binary.Read(bytes.NewReader(read[bi:bi+1]), binary.BigEndian, &valLen)
	if err != nil {
		return nil, fmt.Errorf("error reading value length: %v", err)
	}
	bi += 1

	// Read value (valLen bytes)
	valBytes := make([]byte, valLen)
	err = binary.Read(bytes.NewReader(read[bi:bi+valLen]), binary.BigEndian, &valBytes)
	if err != nil {
		return nil, fmt.Errorf("error reading value bytes: %v", err)
	}

	// Create and return WALLogEntry
	walLogEntry := &WALLogEntry{
		Key:       string(keyBytes),
		Value:     string(valBytes),
		Comm:      Command(encodingType),
		Timestamp: timeStamp,
		LogIndex:  logIndex,
	}

	// Print the value for debugging
	fmt.Println(string(valBytes))
	fmt.Println(walLogEntry)

	return walLogEntry, nil
}
