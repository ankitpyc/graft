package server

import (
	"cache/internal/domain"
	pb "cache/internal/election"
	raft "cache/internal/raft/Client"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Handler interface {
	handleKeyRequest(w http.ResponseWriter, r *http.Request)
	handleGetKey(w http.ResponseWriter, k string) bool
	handleSetKey(w http.ResponseWriter, r *http.Request, walLog []*pb.LogEntry) ([]*pb.LogEntry, bool)
	HealthStatus(w http.ResponseWriter, r *http.Request)
	HandlePeerCon(r *http.Request, w http.ResponseWriter)
	HandleLeaveCon(r *http.Request, w http.ResponseWriter)
	StartGRPCServer()
}

// ServeHTTP handles incoming HTTP requests and routes them based on the URL path
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/key") {
		s.handleKeyRequest(w, r)
	} else if r.URL.Path == "/Health" {
		s.HealthStatus(w, r)
	} else if r.URL.Path == "/Join" {
		s.HandlePeerCon(r, w)
	} else if r.URL.Path == "/Leave" {
		s.HandleLeaveCon(r, w)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// handleKeyRequest processes key-related HTTP requests (GET, POST, DELETE)
func (s *Server) handleKeyRequest(w http.ResponseWriter, r *http.Request) {
	getKey := func() string {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 3 {
			return ""
		}
		return parts[2]
	}
	var walLog = []*pb.LogEntry{}

	switch r.Method {
	case http.MethodGet:
		k := getKey()
		if s.handleGetKey(w, k) {
			return
		}
		return
	case http.MethodPost:
		// Handle POST request to set a key-value pair
		_, done := s.handleSetKey(w, r, walLog)
		if done {
			return
		}
	case http.MethodDelete:
		k := getKey()
		if k == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		s.Client.Store.Delete(k)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	return
}

func (s *Server) handleGetKey(w http.ResponseWriter, k string) bool {
	if k == "" {
		w.WriteHeader(http.StatusBadRequest)
		return true
	}
	v, _ := s.Client.Store.Get(k)
	if v == nil || v == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("key not found"))
		return true
	}
	b, err := json.Marshal(map[string]domain.Key{k: v})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return true
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	return false
}

// handleSetKey handles setting a key in the store, logging the operation, and ensuring replication
func (s *Server) handleSetKey(w http.ResponseWriter, r *http.Request, walLog []*pb.LogEntry) ([]*pb.LogEntry, bool) {
	m := map[string]string{}
	logentry := &pb.AppendEntriesRequest{LeaderId: s.Client.Election.GetLeaderId()}
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, true
	}
	for k, v := range m {
		walLog = append(walLog, &pb.LogEntry{Term: s.WAlManager.LatestCommitIndex, Key: k, Value: v, Operation: 0, TimeName: timestamppb.Now()})
	}
	logentry.Entries = walLog
	wg := &sync.WaitGroup{}
	wg.Add(1)
	err = s.WAlManager.ReplicateEntries(logentry, wg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return nil, false
	}
	wg.Wait()
	// Since the entry has been successfully replicated, put it into leader store
	for _, entry := range walLog {
		s.Client.Store.Set(entry.Key, entry.Value)
	}
	appendLog, err := s.WAlManager.AppendLog(logentry)
	if err != nil || !appendLog {
		return nil, false
	}
	return walLog, false
}

// HealthStatus handles the health check endpoint
func (s *Server) HealthStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Health Check at :", time.Now())
	w.Write([]byte("SUCCESS"))
}

// HandlePeerCon handles the request to join a new peer to the cluster
func (s *Server) HandlePeerCon(r *http.Request, w http.ResponseWriter) {
	var peer *raft.ClusterPeer = &raft.ClusterPeer{}
	if err := json.NewDecoder(r.Body).Decode(&peer); err != nil {
		fmt.Println("err", err)
		return
	}
	fmt.Println("New Peer Added ", peer.NodeAddr+":"+peer.NodePort)
	s.Client.JoinCluster(peer)
}

// HandleLeaveCon handles the request to remove a peer from the cluster
func (s *Server) HandleLeaveCon(r *http.Request, w http.ResponseWriter) {
	var peer *raft.ClusterPeer = &raft.ClusterPeer{}
	if err := json.NewDecoder(r.Body).Decode(&peer); err != nil {
		fmt.Println("err", err)
		return
	}
	fmt.Println("Peer Removed ", peer.NodeAddr+":"+peer.NodePort)
	s.Client.LeaveCluster(peer)
}

// StartGRPCServer starts the gRPC server for handling RPC calls
func (s *Server) StartGRPCServer() {
	port := s.Client.NodeDetails.GrpcPort
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", port, err)
	}

	// Create a new gRPC server instance
	gserv := grpc.NewServer()
	s.Client.GrpcServer = gserv
	// Register the ElectionService and LogReplication Services with the gRPC server
	pb.RegisterLeaderElectionServer(gserv, s.Client.Election)
	pb.RegisterRaftLogReplicationServer(gserv, s.WAlManager.LogReplicator)
	go func() {
		err := gserv.Serve(lis)
		if err != nil {
			log.Fatalf("Failed GRPC serve on port %d: %v", port, err)
		}
	}()
}
