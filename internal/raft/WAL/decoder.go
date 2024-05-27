package wal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
)

func decodeWALEntry(walManager *Manager) (*WALLogEntry, error) {
	buf := make([]byte, 1)
	bytesRead := []byte{}
	walManager.Fd.Seek(0, 0)
	for {
		_, err := walManager.Fd.Read(buf)
		if err != nil {

			log.Fatal("recieved error reading from file: ", err)
		}
		bytesRead = append(bytesRead, buf...)
		if buf[0] == '\n' {
			decodeBytesToWal(bytesRead)
			bytesRead = []byte{}

		}
	}
	return nil, nil
}

func decodeBytesToWal(read []byte) *WALLogEntry {
	var timeStamp, Logindex uint64
	var encodingType uint8
	var klen, vlen uint8

	binary.Read(bytes.NewReader(read[:8]), binary.BigEndian, &timeStamp)
	binary.Read(bytes.NewReader(read[8:16]), binary.BigEndian, &Logindex)
	binary.Read(bytes.NewReader(read[16:17]), binary.BigEndian, &klen)
	keyBytes := make([]byte, klen)
	binary.Read(bytes.NewReader(read[17:17+klen]), binary.BigEndian, &keyBytes)
	fmt.Println(string(keyBytes))
	binary.Read(bytes.NewReader(read[17+klen:17+klen+1]), binary.BigEndian, &encodingType)
	binary.Read(bytes.NewReader(read[17+klen+1:17+klen+2]), binary.BigEndian, &vlen)
	vbytes := make([]byte, vlen)
	binary.Read(bytes.NewReader(read[17+klen+2:17+klen+2+vlen]), binary.BigEndian, &vbytes)
	fmt.Println(string(vbytes))

	log := &WALLogEntry{
		Key:       string(keyBytes),
		Value:     string(vbytes),
		Comm:      Command(encodingType),
		Timestamp: timeStamp,
		LogIndex:  Logindex,
	}
	fmt.Println(log)
	return log
}
