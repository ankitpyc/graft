package main

import (
	"cache/config"
	"cache/server"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
)

func main() {

	c := config.Config{
		Port: "9091",
		Host: "127.0.0.1",
	}

	servHandler := server.NewServerConfig(c)

	httpServer := http.Server{
		Handler: servHandler,
	}

	listener, err := net.Listen("tcp", servHandler.Address+":"+servHandler.Port)

	if err != nil {
		log.Fatal(err)
	}
	//Attach the listener to handler
	servHandler.Listener = listener

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

func handleIncomingTCPConnection(conn net.Conn) {
	defer conn.Close()
}
