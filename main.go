package main

import (
	"cache/config"
	raft "cache/raft/Client"
	"cache/server"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

const (
	ConfigFilename = "config.json"
)

func main() {
	//Loads Configuration
	c, _ := config.NewConfig().LoadConfig(ConfigFilename)
	c.Port = strconv.Itoa(rangeIn(2000, 8000))
	// Initializes Raft Client
	sr := raft.InitRaftClient(c)
	sr.ServiceRegistry.RegisterRaftClient(c)
	servHandler := server.NewServerConfig(*c, sr)
	servHandler.Client = sr
	httpServer := http.Server{
		Handler: servHandler,
	}

	listener, err := net.Listen("tcp", servHandler.Address+":"+servHandler.Port)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Serving on " + servHandler.Address + ":" + servHandler.Port)
	//Attach the listener to handler
	servHandler.Listener = listener
	defer servHandler.Listener.Close()
	//Setting the handler for the route
	http.Handle("/", servHandler)

	go func() {
		err := httpServer.Serve(listener)
		if err != nil {
			log.Fatalf("HTTP serve: %s", err)
		}
	}()

	// this keeps the server running infinitely until Interrupt occurs
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Println("http server stopped")
}

func rangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}

func handleIncomingTCPConnection(conn net.Conn) {
	defer conn.Close()
}
