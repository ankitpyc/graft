package raft

import (
	"bytes"
	"cache/config"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
)

type RegisterInf interface {
	InitServiceRegister(*config.Config)
}

type ServiceRegistry struct{}

type NODETYPE int

const (
	NEWJOIN NODETYPE = iota
	LEAVE
)

type ClusterMember struct {
	NodeType  NODETYPE
	NodeID    string
	NodeAddr  string
	NodePort  string
	ClusterID string
}

func (sr *ServiceRegistry) InitServiceRegister(config *config.Config) {
	dial, err := net.Dial("tcp", config.ServiceDiscoveryAddr)
	if err != nil {
		return
	}
	msg := bytes.NewBuffer(nil)
	member := &ClusterMember{
		NodeType:  NEWJOIN,
		NodeID:    "2",
		NodeAddr:  config.Host,
		NodePort:  config.Port,
		ClusterID: config.ClusterUUID,
	}
	encodedJson, _ := json.Marshal(member)
	binary.Write(msg, binary.BigEndian, uint8(0))
	msg.Write(encodedJson)
	write, err := dial.Write(msg.Bytes())
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Println("written bytes :", write)
	defer dial.Close()
}

func (sr *ServiceRegistry) RegisterRaftClient(config *config.Config) *ServiceRegistry {
	sr.InitServiceRegister(config)
	return sr
}
