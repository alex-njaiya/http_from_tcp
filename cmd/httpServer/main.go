package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"tcp_to_http/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port)

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	defer server.Close()

	log.Println("Server started on port:", port)

	sigChan := make(chan os.Signal, 1) // buffered channel that listens for 1 os signal

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Server gracefully stopped")
}