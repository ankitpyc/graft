package server

import (
	"cache/config"
	"cache/internal/domain"
	"cache/internal/raft/Client"
	wal2 "cache/internal/raft/WAL"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	Listener   net.Listener
	Address    string
	Client     *raft.RaftClient
	WAlManager *wal2.WALManager
	Port       string
}

func NewServerConfig(config config.Config, registry *raft.RaftClient) *Server {
	wlManager := wal2.NewWALManager(config.WALFilePath)
	server := &Server{
		Port:       config.Port,
		Address:    config.Host,
		WAlManager: wlManager,
		Client:     registry,
	}
	return server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/key") {
		s.handleKeyRequest(w, r)
	} else if r.URL.Path == "/health" {
		s.HealthStatus(w, r)
	} else if r.URL.Path == "/Join" {
		s.HandlePeerCon(r, w)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Server) handleKeyRequest(w http.ResponseWriter, r *http.Request) {

	getKey := func() domain.Key {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 3 {
			return ""
		}
		return parts[2]
	}
	var walLog []wal2.WALLogEntry = []wal2.WALLogEntry{}
	switch r.Method {
	case http.MethodGet:
		k := getKey()
		if k == "" {
			w.WriteHeader(http.StatusBadRequest)
		}
		v := s.Client.Store.Get(k)
		if v == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("key not found"))
			return
		}
		b, err := json.Marshal(map[string]domain.Key{k.(string): v})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		io.WriteString(w, string(b))
	case http.MethodPost:
		// Read the value from the POST body.
		m := map[string]domain.Key{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for k, v := range m {
			s.Client.Store.Put(k, v)
			walLog = append(walLog, wal2.WALLogEntry{Comm: 2, Key: k, Value: v})
		}

	case "DELETE":
		k := getKey()
		if k == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		s.Client.Store.Delete(k)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	s.WAlManager.LogStream <- walLog
	return
}

func (s *Server) HealthStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Health Check at :", time.Now())
	w.Write([]byte("SUCCESS"))
}

func (s *Server) HandlePeerCon(r *http.Request, w http.ResponseWriter) {
	var peer *raft.ClusterPeer = &raft.ClusterPeer{}
	if err := json.NewDecoder(r.Body).Decode(&peer); err != nil {
		fmt.Println("err", err)
		return
	}
	fmt.Println("New Peer Added ", peer.NodeAddr+":"+peer.NodePort)
	s.Client.JoinCluster(peer)
}
