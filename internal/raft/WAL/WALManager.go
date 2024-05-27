package wal

import (
	pb "cache/internal/election"
	raft "cache/internal/raft/Client"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"sync"
	"time"
)

type Manager struct {
	sync.RWMutex
	Fd                *os.File
	Log               []*pb.AppendEntriesRequest
	LogStream         chan *pb.AppendEntriesRequest
	LatestCommitIndex int32
	client            *raft.RaftClient
	LogReplicator     *LogReplicationService
	MaxLogSize        uint64
}

func NewWALManager(filename string, client *raft.RaftClient) *Manager {
	fd, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	walManager := &Manager{
		Fd:                fd,
		Log:               []*pb.AppendEntriesRequest{},
		LogStream:         make(chan *pb.AppendEntriesRequest),
		LatestCommitIndex: 0,
		LogReplicator:     NewLogReplicationService(),
		client:            client,
		MaxLogSize:        100,
	}
	walManager.LogReplicator.WALManager = walManager
	go walManager.LogListener()
	return walManager
}

func (walManager *Manager) LogListener() {
	for {
		select {
		case lo :=
			<-walManager.LogStream:
			walManager.AppendLog(lo)
		}
	}
}

func (walManager *Manager) AppendLog(entry *pb.AppendEntriesRequest) bool {
	walManager.Lock()
	defer walManager.Unlock()
	for _, en := range entry.Entries {
		commitedEntry, err := encodeWALEntry(walManager, en)
		walManager.LatestCommitIndex = commitedEntry.Term
		walManager.Log = append(walManager.Log, entry)
		for _, log := range entry.Entries {
			fmt.Println("Adding key ", log.Key)
		}
		log.Println("Key Entry replicated", entry.Entries[0].Key)
		if err != nil {
			fmt.Println("Error commiting entry")
		}
	}
	time.Sleep(10 * time.Second)
	return true
}

func (walManager *Manager) ReplicateEntries(request *pb.AppendEntriesRequest, wg *sync.WaitGroup) {
	walManager.Lock()
	defer wg.Done()
	defer walManager.Unlock()
	log.Print("Replicating entry", request)
	for _, member := range walManager.client.ClusterMembers {
		if member.GrpcPort == walManager.client.NodeDetails.GrpcPort {
			continue
		}
		dial, err := grpc.NewClient(member.NodeAddr+":"+member.GrpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Println(err)
		}
		client := pb.NewRaftLogReplicationClient(dial)
		_, err = client.AppendEntries(context.Context(context.Background()), request)
		if err != nil {
			fmt.Println("error appending entries", err)
		}
	}
}
