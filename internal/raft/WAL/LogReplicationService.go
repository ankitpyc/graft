package wal

import (
	pb "cache/internal/election"
	"fmt"
	"golang.org/x/net/context"
	"log"
)

type LogReplicationService struct {
	pb.UnimplementedRaftLogReplicationServer
	WALManager *Manager
}

func NewLogReplicationService() *LogReplicationService {
	return &LogReplicationService{}
}

func (lr *LogReplicationService) AppendEntries(ctx context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	success := true
	err := lr.ReplicateLogEntry(req)
	if err != nil {
		success = false
		fmt.Printf("ReplicateLogEntry err: %v\n", err)
	}
	return &pb.AppendEntriesResponse{
		Success: success,
		Term:    req.Term,
	}, nil
}

func (lr *LogReplicationService) ReplicateLogEntry(request *pb.AppendEntriesRequest) error {
	_, err := lr.WALManager.AppendLog(request)
	if err != nil {
		return fmt.Errorf("error while replicating log entry: %v", err)
	}
	log.Println("Log Stream Ended Entry Replicated")
	return nil
}
