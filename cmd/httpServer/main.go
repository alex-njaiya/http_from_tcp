package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
		if target_path == "/yourproblem" {
			w.WriteStatusLine(status)
			body := response400()

			headers.Set("Content-Length", strconv.Itoa(len(body)))
			headers.Set("Content-Type", "text/html")

			w.WriteHeaders(headers)
			w.WriteBody(body)
		}

		if target_path == "/myproblem" {
			body := response500()
			status = response.StatusCode(500)

			w.WriteStatusLine(status)

			headers.Set("Content-Length", strconv.Itoa(len(body)))
			headers.Set("Content-Type", "text/html")

			w.WriteHeaders(headers)
			w.WriteBody(body)
		}
		if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/stream") {
			// read the http request
			res, err := http.Get("https://httpbin.org/" + target_path[len("/httpbin/"):])

			if err != nil {
				body = response500()
				status = http.StatusInternalServerError
			} else {
				// write the statusline
				w.WriteStatusLine(response.StatusCode(200))

				// delete the content length headers
				headers.Delete("Content-Length")

				// add a Tranfer-Encoding header to the request headers
				headers.Set("Transfer-Encoding", "chunked")
				headers.Set("Content-Type", "text/plain")
				// Write the headers
				w.WriteHeaders(headers)
				// Write the body

				for {
					buf := make([]byte, 32)

					read, err := res.Body.Read(buf)

					if err != nil {
						break
					}

					//write the hex rep for the chunk read
					w.WriteBody([]byte(fmt.Sprintf("%x", read)))
					w.WriteBody([]byte("\r\n"))

					w.WriteBody(buf[:read])
					w.WriteBody([]byte("\r\n"))
				}
				w.WriteBody([]byte("0\r\n"))
				return

			}
		}

		headers.Set("Content-Length", strconv.Itoa(len(body)))
		headers.Set("Content-Type", "text/html")
		status = response.StatusCode(200)
		w.WriteStatusLine(status)
		w.WriteHeaders(headers)
		w.WriteBody(response200())
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
