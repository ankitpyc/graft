package raft

import (
	"cache/config"
	"math/rand"
	"sync"
)

type ClusterPeer struct {
	NodeName   string
	NodeId     string
	NodeAddr   string
	NodePort   string
	NodeStatus string
}

func NewClusterPeer(nodeId string, nodeAddr string, nodePort string) *ClusterPeer {
	return &ClusterPeer{
		NodeId:     nodeId,
		NodeAddr:   nodeAddr,
		NodePort:   nodePort,
		NodeStatus: "FOLLOWER",
	}
}

type RaftClient struct {
	ClusterName     string
	ClusterID       uint64
	ClusterMembers  []*ClusterPeer
	NodeDetails     *ClusterPeer
	ServiceRegistry *ServiceRegistry
	RMu             sync.RWMutex
}

func (client *RaftClient) AddClusterPeer(peer *ClusterPeer) {
	client.RMu.Lock()
	client.ClusterMembers = append(client.ClusterMembers, peer)
	client.RMu.Unlock()
}

func InitRaftClient(config *config.Config) *RaftClient {
	return &RaftClient{
		ClusterName:     config.ClusterName,
		ClusterID:       rand.Uint64(),
		ClusterMembers:  make([]*ClusterPeer, 0, 5),
		NodeDetails:     NewClusterPeer("0", config.Host, config.Port),
		ServiceRegistry: &ServiceRegistry{},
		RMu:             sync.RWMutex{},
	}
}
