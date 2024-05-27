package raft

type NodeRegisterationResponse struct {
	Secret  string         `json:"Secret"`
	Members []*ClusterPeer `json:"Members"`
}
