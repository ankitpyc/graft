package raft

import (
	"cache/config"
	"cache/factory"
	Cache "cache/internal/domain/interface"
	pb "cache/internal/election"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
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
	Apply()
	IsLeader() bool
}

type RaftClient struct {
	ClusterName     string
	ClusterID       uint64
	ClusterMembers  []*ClusterPeer
	NodeDetails     *ClusterPeer
	ServiceRegistry *ServiceRegistry
	MemberChannel   chan []*ClusterPeer
	Election        *ElectionService
	Store           Cache.Cache
	RMu             sync.RWMutex
}

func (client *RaftClient) JoinCluster(peer *ClusterPeer) {
	fmt.Println("Node :- ", peer.NodeAddr+":"+peer.NodePort)
	client.RMu.Lock()
	client.ClusterMembers = append(client.ClusterMembers, peer)
	client.RMu.Unlock()
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
	go listenForChannelEvents(client)
	go client.StartElectionServer()
	return client
}

func (client *RaftClient) BuildStore() (Cache.Cache, error) {
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

func (client *RaftClient) StartElectionServer() {
	sy := sync.WaitGroup{}
	client.Election = NewElectionService(client)
	sy.Add(1)
	port := strconv.Itoa(rangeIn(8100, 10000))
	go startElectionServer(port, client, &sy)
	sy.Wait()
}

func startElectionServer(port string, raft *RaftClient, wg *sync.WaitGroup) {
	defer wg.Done()
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", port, err)
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register the ElectionService with the gRPC server
	pb.RegisterLeaderElectionServer(s, raft.Election)

	// Log that the server is now listening on the specified port
	log.Printf("Election server listening on port %d", port)

	// Start serving incoming connections
	go s.Serve(lis)
	if err != nil {
		log.Fatalf("Failed to serve on port %d: %v", port, err)
	}

	_, _ = context.WithCancel(context.Background())
	// Start the election loop
	err = raft.Election.RunElectionLoop()
	if err != nil {
		fmt.Println("Election lost")
	}
}

func (client *RaftClient) Apply() {

}

func (client *RaftClient) IsLeader() bool {
	return client.NodeDetails.NodePort == client.Election.GetLeaderId()
}

func rangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}
