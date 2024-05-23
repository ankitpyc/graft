package raft

import (
	"bytes"
	"cache/config"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type ServiceRegistryInf interface {
	InitServiceRegister(*config.Config)
}

type ServiceRegistry struct {
	client *RaftClient
}

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
	conn, err := net.Dial("tcp", config.ServiceDiscoveryAddr)
	if err != nil {
		log.Fatal("connect service register fail", err)
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
	write, err := conn.Write(msg.Bytes())
	if err != nil {
		fmt.Println("err", err)
	}
	go readResponse(sr, conn)
	fmt.Println("written bytes :", write)
	//defer conn.Close()
}

func readResponse(service *ServiceRegistry, conn net.Conn) {
	var buffer bytes.Buffer
	var ClusterMember []*ClusterPeer = make([]*ClusterPeer, 0)
	for {
		data := make([]byte, 1024)
		n, err := conn.Read(data)
		if err != nil {
			// Handle error or potential connection close
		}
		buffer.Write(data[:n])
		if bytes.IndexByte(buffer.Bytes(), '\n') > -1 { // Check for newline delimiter (replace if needed)
			break
		}
		json.Unmarshal(data[:n], &ClusterMember)
		fmt.Printf("Total Cluster Members %v\n", ClusterMember)
		service.client.MemberChannel <- ClusterMember
	}
}

func (sr *ServiceRegistry) RegisterRaftClient(config *config.Config) {
	sr.InitServiceRegister(config)
}
