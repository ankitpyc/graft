package raft

import (
	"cache/config"
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
		NodeType: 1,
	}
}

type RaftClient struct {
	ClusterName     string
	ClusterID       uint64
	ClusterMembers  []*ClusterPeer
	NodeDetails     *ClusterPeer
	ServiceRegistry *ServiceRegistry
	MemberChannel   chan []*ClusterPeer
	RMu             sync.RWMutex
}

func (client *RaftClient) AddClusterPeer(peer *ClusterPeer) {
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
	client.ServiceRegistry = &reg
	client.ServiceRegistry.client = client
	go listenForChannelEvents(client)
	return client
}

func listenForChannelEvents(client *RaftClient) {
	for {
		select {
		case event := <-client.MemberChannel:
			client.RMu.RLock()
			for _, peer := range event {
				fmt.Println("got peer as", peer)
				client.AddClusterPeer(peer)
			}
			client.RMu.RUnlock()
		}
	}
}
