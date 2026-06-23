package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
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
		fmt.Printf("DEBUG: Received path = %q\n", target_path)

		// ============ yourproblem path ===================
		if target_path == "/yourproblem" {
			body := response400()

			headers := response.GetDefaultHeaders(0)
			headers.Set("Content-Length", strconv.Itoa(len(body)))
			headers.Set("Content-Type", "text/html")

			w.WriteStatusLine(response.StatusCode(400))
			w.WriteHeaders(headers)
			w.WriteBody(body)
			return
		} else if target_path == "/myproblem" {
			body := response500()
			headers := response.GetDefaultHeaders(0)
			headers.Set("Content-Length", strconv.Itoa(len(body)))
			headers.Set("Content-Type", "text/html")

			w.WriteStatusLine(response.StatusCode(500))
			w.WriteHeaders(headers)
			w.WriteBody(body)
			return
		} else if req.RequestLine.RequestTarget == "/video" {
			//we are going to read the file chunk by chunk
			file, err := os.Open("assets/vim.mp4")

			if err != nil {
				log.Printf("Error opening the video file: %v", err)
				w.WriteStatusLine(response.StatusCode(500))
				return
			}

			defer file.Close()

			// get the real file size
			stat, err := file.Stat()
			if err != nil {
				log.Printf("Error stating file: %v", err)
				w.WriteStatusLine(response.StatusCode(500))
				return
			}

			// write the statusline
			headers := response.GetDefaultHeaders(0)
			headers.Set("Content-Type", "video/mp4")
			headers.Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
			headers.Set("Accept-Ranges", "bytes")
			headers.Set("Cache-Control", "no-cache")

			// write the statusline and the headers
			w.WriteStatusLine(response.StatusCode(200))
			w.WriteHeaders(headers)

			// define the chunk size and read the chunks in a loop
			const chunkSize = 1024 * 1024
			buf := make([]byte, chunkSize)

			hasher := sha256.New()
			totalBytes := int64(0)

			for {
				n, err := file.Read(buf)

				if n > 0 {
					chunk := buf[:n]
					hasher.Write(chunk)
					totalBytes += int64(n)

					if _, err := w.WriteChunkedBody(chunk); err != nil {
						return
					}
				}
				
				if err == io.EOF {
					break
				}

				if err != nil {
					log.Printf("Error reading chunk from file: %v", err)
				}

			}

			// add the trailers
			trailerHeaders := response.GetDefaultHeaders(0)
			hashValue := hasher.Sum(nil)
			trailerHeaders.Delete("Content-Length")
			trailerHeaders.Set("X-Content-SHA256", hex.EncodeToString(hashValue))
			trailerHeaders.Set("X-Content-Length", strconv.FormatInt(totalBytes, 10))

			if err := w.WriteTrailers(trailerHeaders); err != nil {
				fmt.Printf("DEBUG: WriteTrailers error: %v\n", err)
			}
			return

		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbingo/") {
			fmt.Println("DEBUG: Entering /httpbingo/ proxy handler")
			// read the http request
			proxyPath := strings.TrimPrefix(target_path, "/httpbingo/")
			res, err := http.Get("https://httpbingo.org/" + proxyPath)

			if err != nil {
				fmt.Printf("DEBUG: http.Get error: %v\n", err)
				return
			}

			defer res.Body.Close()

			// Initial response headers -- fetch freah object
			initHeaders := response.GetDefaultHeaders(0)
			// delete the content length headers
			initHeaders.Delete("Content-Length")

			// add a Tranfer-Encoding header to the request headers
			initHeaders.Set("Transfer-Encoding", "chunked")
			initHeaders.Set("Content-Type", res.Header.Get("Content-Type"))
			initHeaders.Set("Trailers", "X-Content-SHA256, X-Content-Length") // announce trailers

			// Write the headers and startline
			w.WriteStatusLine(response.StatusCode(res.StatusCode))
			w.WriteHeaders(initHeaders)

			// hashing
			hasher := sha256.New()
			totalBytes := 0
			buf := make([]byte, 32*1024)

			for {
				read, err := res.Body.Read(buf)

				if read > 0 {
					chunk := buf[:read]
					hasher.Write(chunk)
					totalBytes += read

					if _, err := w.WriteChunkedBody(chunk); err != nil {
						return
					}
				}

				// handle eof error
				if err == io.EOF {
					break
				}

				if err != nil {
					break
				}
			}

			// Build the trailers header and send them
			trailerHeaders := response.GetDefaultHeaders(0)
			hashValue := hasher.Sum(nil)
			trailerHeaders.Delete("Content-Length")
			trailerHeaders.Set("X-Content-SHA256", hex.EncodeToString(hashValue))
			trailerHeaders.Set("X-Content-Length", strconv.Itoa(totalBytes))

			if err := w.WriteTrailers(trailerHeaders); err != nil {
				fmt.Printf("DEBUG: WriteTrailers error: %v\n", err)
			}
			return

		}

		// ================== DEFAULT ==================
		fmt.Println("DEBUG: Falling to default handler")
		body := response200()

		headers := response.GetDefaultHeaders(0)
		headers.Set("Content-Length", strconv.Itoa(len(body)))
		headers.Set("Content-Type", "text/html")

		w.WriteStatusLine(response.StatusCode(200))
		w.WriteHeaders(headers)
		w.WriteBody(body)
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
