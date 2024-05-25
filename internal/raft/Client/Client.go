package raft

import (
	"cache/config"
	"cache/factory"
	"cache/internal/store"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type RaftClientInf interface {
	InitRaftClient(config *config.Config) *RaftClient
	JoinCluster(peer *ClusterPeer)
	RunElectionLoop(ctx context.Context, ClusterMembers []*ClusterPeer) error
	LeaveCluster(peer *ClusterPeer)
	GetLeader(peer *ClusterPeer)
	ListenForLeader() bool
	StartGRPCServer()
	Apply()
	IsLeader() bool
	AppendEntries(peer *ClusterPeer) error
}

type RaftClient struct {
	ClusterName     string
	ClusterID       uint64
	ClusterMembers  []*ClusterPeer
	NodeDetails     *ClusterPeer
	ServiceRegistry *ServiceRegistry
	MemberChannel   chan []*ClusterPeer
	Election        *ElectionService
	Store           store.StoreInf
	GrpcServer      *grpc.Server
	RMu             sync.RWMutex
}

func (client *RaftClient) JoinCluster(peer *ClusterPeer) {
	fmt.Println("Node :- ", peer.NodeAddr+":"+peer.NodePort, " || grpc port ", peer.NodeAddr+":"+peer.GrpcPort)
	client.RMu.Lock()
	client.ClusterMembers = append(client.ClusterMembers, peer)
	client.RMu.Unlock()
}
func (client *RaftClient) LeaveCluster(peer *ClusterPeer) {
	fmt.Println("Node :- ", peer.NodeAddr+":"+peer.NodePort, " || grpc port ", peer.NodeAddr+":"+peer.GrpcPort)
	client.RMu.Lock()
	client.leaveCluster(peer)
	client.RMu.Unlock()
}

func (client *RaftClient) leaveCluster(peer *ClusterPeer) {
	for i, mem := range client.ClusterMembers {
		if mem.NodePort == peer.NodePort {
			// Remove the member from the slice
			client.ClusterMembers = append(client.ClusterMembers, client.ClusterMembers[i+1:]...)
		}
	}
}

func InitRaftClient(config *config.Config) *RaftClient {
	reg := ServiceRegistry{}
	client := &RaftClient{
		ClusterName:    config.ClusterName,
		ClusterID:      rand.Uint64(),
		ClusterMembers: make([]*ClusterPeer, 0, 5),
		NodeDetails:    NewClusterPeer("0", config.Host, config.Port),
		RMu:            sync.RWMutex{},
		MemberChannel:  make(chan []*ClusterPeer),
	}
	client.Store, _ = client.BuildStore()
	client.ServiceRegistry = &reg
	client.ServiceRegistry.client = client
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go listenForChannelEvents(client)
	client.StartElectionServer(wg)
	wg.Wait()
	return client
}

func (client *RaftClient) BuildStore() (store.StoreInf, error) {
	cache, err := factory.CreateCache("LFU", 4)
	if err != nil {
		return nil, err
	}
	return cache, nil
}

func listenForChannelEvents(client *RaftClient) {
	for {
		select {
		case event := <-client.MemberChannel:
			fmt.Println("Cluster Info : Total Members : ", len(event))
			for _, peer := range event {
				client.JoinCluster(peer)
			}
		}
	}
}

func (client *RaftClient) ListenForLeader() bool {
	timer := time.NewTimer(time.Second * 5)
	for {
		select {
		case <-timer.C:
			if !client.NodeDetails.HasReceivedLeaderPing {
			}
		}
	}
}

func (client *RaftClient) StartElectionServer(wg *sync.WaitGroup) {
	defer wg.Done()
	sy := &sync.WaitGroup{}
	client.Election = NewElectionService(client)
	sy.Add(1)
	port := strconv.Itoa(rangeIn(8100, 10000))
	client.NodeDetails.GrpcPort = port
	startElectionServer(client, sy)
	sy.Wait()
}

func startElectionServer(raft *RaftClient, wg *sync.WaitGroup) {
	defer wg.Done()
	// Start the election loop
	go func() {
		err := raft.Election.RunElectionLoop()
		if err != nil {
			log.Fatalf("Election lost: %v", err)
		}
	}()
}

func (client *RaftClient) Apply() {

}

func (client *RaftClient) IsLeader() bool {
	return client.NodeDetails.NodePort == client.Election.GetLeaderId()
}

func (client *RaftClient) AppendEntries() bool {
	return client.NodeDetails.NodePort == client.Election.GetLeaderId()
}

func rangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}
