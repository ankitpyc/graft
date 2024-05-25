package wal

import (
	pb "cache/internal/election"
	"fmt"
	"golang.org/x/net/context"
	"log"
	"time"
)

type LogReplicationService struct {
	pb.UnimplementedRaftLogReplicationServer
	WALManager *WALManager
}

func NewLogReplicationService() *LogReplicationService {
	return &LogReplicationService{}
}

func (lr *LogReplicationService) AppendEntries(ctx context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	fmt.Println("Received AppendEntriesRequest")
	lr.ReplicateLogEntry(req)
	time.Sleep(5 * time.Second)
	return &pb.AppendEntriesResponse{
		Success: false,
		Term:    0,
	}, nil
}

func (lr *LogReplicationService) ReplicateLogEntry(request *pb.AppendEntriesRequest) {
	lr.WALManager.LogStream <- request
	log.Println("Log Stream Ended Entry Replicated")
}
