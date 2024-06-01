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
)

type ManagerInf interface {
	LogListener()
	AppendLog(entry *pb.AppendEntriesRequest) (bool, error)
	UpdateReplicaDataStore(entry *pb.AppendEntriesRequest) (bool error)
}

type Manager struct {
	sync.RWMutex
	Fd                *os.File
	Log               []*pb.AppendEntriesRequest
	LogStream         chan *pb.AppendEntriesRequest
	LatestCommitIndex int32
	client            *raft.Client
	LogReplicator     *LogReplicationService
	Snappy            *Snappy
	MaxLogSize        uint64
}

func NewWALManager(filename string, client *raft.Client, option ...Options) *Manager {
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
	for _, opt := range option {
		opt(walManager)
	}
	return walManager
}

func (walManager *Manager) AppendLog(entry *pb.AppendEntriesRequest) (bool, error) {
	walManager.Lock()
	defer walManager.Unlock()
	for _, en := range entry.Entries {
		commitedEntry, err := encodeWALEntry(walManager, en)
		if err != nil || commitedEntry == nil {
			return false, fmt.Errorf("encode log entry failed: %v", err)
		}
		walManager.LatestCommitIndex = commitedEntry.Term
		if !walManager.client.IsLeader() {
			log.Printf("The node is a Replica %v \n", walManager.client.NodeDetails.NodePort)
			er := walManager.UpdateReplicaDataStore(entry)
			if er != nil {
				return false, fmt.Errorf("updating replica failed: %v", err)
			}
			log.Printf("Entry replicated at follower node %v \n ", walManager.client.NodeDetails.GrpcPort)
		}
		walManager.Log = append(walManager.Log, entry)
	}
	return true, nil
}

func (walManager *Manager) UpdateReplicaDataStore(entry *pb.AppendEntriesRequest) error {
	store := walManager.client.Store
	for _, en := range entry.Entries {
		switch en.Operation {
		case 0:
			store.Set(en.Key, en.Value)
			break
		case 1:
			store.Delete(en.Key)
		default:
			return fmt.Errorf("invalid operation in the store %d", en.Operation)
		}
	}
	return nil
}

func (walManager *Manager) ReplicateEntries(request *pb.AppendEntriesRequest, wg *sync.WaitGroup) error {
	walManager.Lock()
	defer wg.Done()
	defer walManager.Unlock()
	log.Print("Replicating entry", request)
	maxQuorum := len(walManager.client.ClusterMembers)
	replicatedNodes := 0
	for _, member := range walManager.client.ClusterMembers {
		if member.GrpcPort == walManager.client.NodeDetails.GrpcPort {
			continue
		}
		dial, err := grpc.NewClient(member.NodeAddr+":"+member.GrpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Println(err)
		}
		client := pb.NewRaftLogReplicationClient(dial)
		entries, err := client.AppendEntries(context.Context(context.Background()), request)
		if err != nil || entries == nil || !entries.Success {
			continue
		}
		if entries.Success {
			replicatedNodes += 1
		}
	}
	if (maxQuorum / 2) > replicatedNodes {
		return fmt.Errorf("replication has failed ! Quorum between nodes cant be acheived")
	}
	return nil
}
