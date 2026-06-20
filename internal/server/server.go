package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"tcp_to_http/internal/request"
	"tcp_to_http/internal/response"
)

type Server struct {
	closed   atomic.Bool
	listener net.Listener
	handler  Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w response.Writer, req *request.Request)


func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return nil, err
	}

	server := &Server{listener: listener, handler: handler}
	fmt.Println("Serving on port: ", port)

	go server.listen()

	return server, nil

}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()

		if err != nil {
			if s.closed.Load() {
				return
			}

			log.Printf("Error accepting connection: %v", err)
			continue
		}

		//handle the client in the background
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	fmt.Println("New client connected: ", conn.RemoteAddr())

	responseWriter := response.NewWriter(conn)
	r, err := request.RequestFromReader(conn)

	if err != nil {
		responseWriter.WriteStatusLine(response.StatusCode(500))
		responseWriter.WriteHeaders(response.GetDefaultHeaders(0))
		return
	}
	// Wrap the connection in a buffered reader
	s.handler(*responseWriter, r)

}
