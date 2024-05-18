package server

import (
	"cache/config"
	factory "cache/factory"
	"cache/internal/domain"
	Cache "cache/internal/domain/interface"
	raft "cache/raft/WAL"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

type Server struct {
	Listener net.Listener
	Address  string
	log      []*raft.WALLogEntry
	fd       *os.File
	Port     string
	store    Cache.Cache
}

func NewServerConfig(config config.Config) *Server {
	cache, _ := factory.CreateCache("LRU", 5)
	file, err := os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	return &Server{
		Port:    config.Port,
		Address: config.Host,
		store:   cache,
		fd:      file,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/key") {
		s.handleKeyRequest(w, r)
		s.store.GetAllCacheData()
	} else if r.URL.Path == "/join" {
		s.handleJoin(w, r)
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

	switch r.Method {
	case http.MethodGet:
		k := getKey()
		if k == "" {
			w.WriteHeader(http.StatusBadRequest)
		}
		v := s.store.Get(k)
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
			s.store.Put(k, v)
			walLog := raft.NewWALLogEntry(raft.Command(2), k, v)
			encoder := json.NewEncoder(s.fd)
			if err := encoder.Encode(&walLog); err != nil {
				log.Print("Failed to encode log entry to JSON: %s", err)
			}
		}

	case "DELETE":
		k := getKey()
		if k == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		s.store.Delete(k)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	return
}

func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {

}
