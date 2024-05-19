package raft

type ClusterPeer struct {
	NodeName   string
	NodeId     uint64
	NodeAddr   string
	NodePort   uint8
	NodeStatus string
}

func NewClusterPeer(nodeId uint64, nodeAddr string, nodePort uint8) *ClusterPeer {
	return &ClusterPeer{
		NodeId:   nodeId,
		NodeAddr: nodeAddr,
		NodePort: nodePort,
	}
}

type Client struct {
	ClusterName    string
	ClusterID      uint64
	ClusterMembers []*ClusterPeer
	NodeDetails    ClusterPeer
}

func initRaftClient(clusterName string) *Client {
	return &Client{
		ClusterName:    clusterName,
		ClusterID:      0,
		ClusterMembers: make([]*ClusterPeer, 0),
		NodeDetails:    ClusterPeer{},
	}
}
