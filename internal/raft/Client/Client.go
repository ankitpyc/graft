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
)

type ClientInf interface {
	InitRaftClient(config *config.Config) *Client
	JoinCluster(peer *ClusterPeer)
	RunElectionLoop(ctx context.Context, ClusterMembers []*ClusterPeer) error
	LeaveCluster(peer *ClusterPeer)
	GetLeader(peer *ClusterPeer)
	StartGRPCServer()
	Apply()
	IsLeader() bool
	AppendEntries(peer *ClusterPeer) error
}

type Client struct {
	sync.Mutex
	ClusterName     string
	ClusterID       string
	ClusterMembers  []*ClusterPeer
	NodeDetails     *ClusterPeer
	ServiceRegistry *ServiceRegistry
	MemberChannel   chan []*ClusterPeer
	Election        *ElectionService
	Store           store.StoreInf
	GrpcServer      *grpc.Server
}

func (client *Client) JoinCluster(peer *ClusterPeer) {
	fmt.Println("Node :- ", peer.NodeAddr+":"+peer.NodePort, " || grpc port ", peer.NodeAddr+":"+peer.GrpcPort)
	client.Lock()
	client.ClusterMembers = append(client.ClusterMembers, peer)
	client.Unlock()
}

func (client *Client) LeaveCluster(peer *ClusterPeer) {
	fmt.Println("Node :- ", peer.NodeAddr+":"+peer.NodePort, " || grpc port ", peer.NodeAddr+":"+peer.GrpcPort)
	client.leaveCluster(peer)
}

func (client *Client) leaveCluster(peer *ClusterPeer) {
	for i, mem := range client.ClusterMembers {
		if mem.NodePort == peer.NodePort {
			client.Lock()
			defer client.Unlock()
			// Remove the member from the slice
			client.ClusterMembers = append(client.ClusterMembers, client.ClusterMembers[i+1:]...)
		}
	}
}

func InitRaftClient(config *config.Config, option ...Option) *Client {
	client := &Client{
		ClusterName:    config.ClusterName,
		ClusterID:      config.ClusterUUID,
		ClusterMembers: make([]*ClusterPeer, 0, 5),
		NodeDetails:    NewClusterPeer("0", config.Host, config.Port),
		MemberChannel:  make(chan []*ClusterPeer),
	}
	for _, opt := range option {
		opt(client)
	}
	go listenForChannelEvents(client)
	client.StartElectionServer(config)
	return client
}

func (client *Client) BuildStore() (store.StoreInf, error) {
	cache, err := factory.CreateCache("LFU", 4)
	if err != nil {
		return nil, fmt.Errorf("create store err: %v", err)
	}
	return cache, nil
}

func listenForChannelEvents(client *Client) {
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

func (client *Client) StartElectionServer(config *config.Config) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	client.Election = NewElectionService(client)
	if config.DebugMode {
		config.GRPCPort = strconv.Itoa(rangeIn(8100, 10000))
	}
	client.NodeDetails.GrpcPort = config.GRPCPort
	startElectionServer(client, wg)
	wg.Wait()
}

func startElectionServer(raft *Client, wg *sync.WaitGroup) {
	// Start the election loop
	defer wg.Done()
	go func() {
		err := raft.Election.RunElectionLoop()
		if err != nil {
			log.Fatalf("Election lost: %v", err)
		}
	}()
}

func (client *Client) IsLeader() bool {
	return client.NodeDetails.NodePort == client.Election.GetLeaderId()
}

func rangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}
