package server

import (
	"cache/config"
	"cache/internal/raft/Client"
	wal2 "cache/internal/raft/WAL"
	"net"
)

// Server struct holds the server configuration and components
type Server struct {
	Listener   net.Listener
	Address    string
	Client     *raft.Client
	WAlManager *wal2.Manager
	Port       string
}

// NewServerConfig initializes a new Server instance with provided configuration and raft client
func NewServerConfig(config config.Config, registry *raft.Client) *Server {
	wlManager := wal2.NewWALManager(config.WALFilePath, registry, wal2.WithLogReplication())
	server := &Server{
		Port:       config.Port,
		Address:    config.Host,
		WAlManager: wlManager,
		Client:     registry,
	}
	return server
}
