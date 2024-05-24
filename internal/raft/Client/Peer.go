package raft

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
	GrpcPort              string
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
