package main

import (
	"cache/config"
	"cache/internal/raft/Client"
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
	c, _ := config.LoadConfig(ConfigFilename)
	if c.DebugMode {
		c.Port = strconv.Itoa(rangeIn(5000, 6000))
	}
	// Initializes Raft Client
	client := raft.InitRaftClient(c, raft.WithServiceReg(), raft.WithDataStore())
	servHandler := server.NewServerConfig(*c, client)
	servHandler.Client = client
	client.ServiceRegistry.RegisterRaftClient(c)
	servHandler.StartGRPCServer()
	httpServer := http.Server{
		Handler: servHandler,
	}

	listener, err := net.Listen("tcp", servHandler.Address+":"+servHandler.Port)

	if err != nil {
		log.Fatal("err")
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
	client.ServiceRegistry.DeregisterService(c)
	log.Println("http server stopped")
}

func rangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}

func handleIncomingTCPConnection(conn net.Conn) {
	defer conn.Close()
}
