package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tcp_to_http/internal/request"
	"tcp_to_http/internal/response"
	"tcp_to_http/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, func(w response.Writer, req *request.Request) {
		// get the path and do logic against that path
		target_path := req.RequestLine.RequestTarget
		headers := response.GetDefaultHeaders(0)
		body := response200()
		status := response.StatusCode(400)
		switch target_path {
		case "/yourproblem":
			w.WriteStatusLine(status)
			body := response400()

			headers.Set("Content-Length", fmt.Sprintf("%d\n", len(body)))
			w.WriteHeaders(headers)
			w.WriteBody(body)

		case "/myproblem":
			body := response500()
			status = response.StatusCode(500)

			w.WriteStatusLine(status)

			headers.Set("Content-Length", fmt.Sprintf("%d\n", len(body)))
			w.WriteHeaders(headers)
			w.WriteBody(body)

		default:
			headers.Set("Content-Length", fmt.Sprintf("%d\n", len(body)))
			status = response.StatusCode(200)
			w.WriteStatusLine(status)
			w.WriteHeaders(headers)
			w.WriteBody(response200())
		}
	})

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

func response200() []byte {
	return []byte(`
		<html>
			<head>
			<title>200 OK</title>
			</head>
			<body>
			<h1>Success!</h1>
			<p>Your request was an absolute banger.</p>
			</body>
		</html>
	`)
}

func response400() []byte {
	return []byte(`
		<html>
			<head>
			<title>400 Bad Request</title>
			</head>
			<body>
			<h1>Bad Request</h1>
			<p>Your request honestly kinda sucked.</p>
			</body>
		</html>
	`)
}

func response500() []byte {
	return []byte(`
		<html>
			<head>
			<title>500 Internal Server Error</title>
			</head>
			<body>
			<h1>Internal Server Error</h1>
			<p>Okay, you know what? This one is on me.</p>
			</body>
		</html>
	`)
}
