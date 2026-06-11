package main

import (
	"fmt"
	"log"
	"net"
	"tcp_to_http/internal/request"
)

func main() {
	// create and accept a tcp connection
	listener, err := net.Listen("tcp", ":42069")

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Fatalf("error: %v", err)
		}

		fmt.Println("A connection has been accepted")

		r, err := request.RequestFromReader(conn)

		if err != nil {
			log.Fatalf("Error: %v", err)
		}

			
		fmt.Printf("Request line:\n - Method: %s\n - Target: %s\n - Version: %s", r.RequestLine.Method, r.RequestLine.RequestTarget, r.RequestLine.HttpVersion)
	}
}
