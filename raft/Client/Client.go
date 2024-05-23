package raft

import (
	"cache/config"
	"cache/factory"
	Cache "cache/internal/domain/interface"
	"fmt"
	"math/rand"
	"sync"
)

type NODESTATE int

const (
	FOLLOWER NODESTATE = iota
	LEADER
	CANDIDATE
)

type ClusterPeer struct {
	NodeType  NODESTATE
	NodeID    string
	NodeAddr  string
	NodePort  string
	ClusterID string
}

func NewClusterPeer(nodeId string, nodeAddr string, nodePort string) *ClusterPeer {
	return &ClusterPeer{
		NodeID:   nodeId,
		NodeAddr: nodeAddr,
		NodePort: nodePort,
		NodeType: 0,
	}
}

type RaftClientInf interface {
	InitRaftClient(config *config.Config) *RaftClient
	JoinCluster(peer *ClusterPeer)
	LeaveCluster(peer *ClusterPeer)
	GetLeader(peer *ClusterPeer)
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
