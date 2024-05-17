package server

import (
	"cache/config"
	factory "cache/factory"
	"cache/internal/domain"
	Cache "cache/internal/domain/interface"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
)

type Server struct {
	Listener net.Listener
	Address  string
	Port     string
	store    Cache.Cache
}

func NewServerConfig(config config.Config) *Server {
	cache, _ := factory.CreateCache("TTL", 5)
	return &Server{
		Port:    config.Port,
		Address: config.Host,
		store:   cache,
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
		if v == -1 {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("key not found"))
			return
		}
		b, err := json.Marshal(map[domain.Key]domain.Key{k: v})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		io.WriteString(w, string(b))
	case http.MethodPost:
		// Read the value from the POST body.
		m := make(map[domain.Key]domain.Key)
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for k, v := range m {
			s.store.Put(k, v)
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
