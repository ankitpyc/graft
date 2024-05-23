package raft

import (
	"cache/config"
	"cache/factory"
	Cache "cache/internal/domain/interface"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type NODESTATE int

const (
	FOLLOWER NODESTATE = iota
	LEADER
	CANDIDATE
)

type ClusterPeer struct {
	NodeType              NODESTATE
	NodeID                string
	NodeAddr              string
	NodePort              string
	ClusterID             string
	HasReceivedLeaderPing bool
}

func NewClusterPeer(nodeId string, nodeAddr string, nodePort string) *ClusterPeer {
	return &ClusterPeer{
		NodeID:                nodeId,
		NodeAddr:              nodeAddr,
		NodePort:              nodePort,
		NodeType:              0,
		HasReceivedLeaderPing: false,
	}
}

type RaftClientInf interface {
	InitRaftClient(config *config.Config) *RaftClient
	JoinCluster(peer *ClusterPeer)
	LeaveCluster(peer *ClusterPeer)
	GetLeader(peer *ClusterPeer)
	ListenForLeader(*config.Config)
	StartLeaderElection(*config.Config)
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
	go client.ListenForLeader()
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

func (client *RaftClient) IsLeader() bool {
	return client.NodeDetails.HasReceivedLeaderPing
}

func (client *RaftClient) ListenForLeader() bool {
	timer := time.NewTimer(time.Second * 5)
	for {
		select {
		case <-timer.C:
			if !client.NodeDetails.HasReceivedLeaderPing {
				client.StartLeaderElection()
			}
		}
	}
}

func (client *RaftClient) StartLeaderElection() {
	log.Print("Starting leader election")
}
