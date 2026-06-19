package response

import (
	"fmt"
	"io"
	"strconv"
	"tcp_to_http/internal/headers"
)

type Writer struct {
	writer io.Writer
}

type StatusCode uint

const (
	success      StatusCode = 200
	bad_request  StatusCode = 400
	server_error StatusCode = 500
)

func NewWriter(writer io.Writer) *Writer {
	return &Writer{
		writer: writer,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	statusLine := []byte{}
	switch statusCode {
	case success:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case bad_request:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case server_error:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	default:
		return fmt.Errorf("unrecognized error code")
	}

	_, err := w.writer.Write(statusLine)
	return err
}

func (w *Writer)WriteHeaders(headers headers.Headers) error {
	for k, v := range headers.All() {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}

	_, err := w.writer.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteBody(b []byte) (int, error) {
	n, err := w.writer.Write(b)

	return n, err
} 

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	h.Set("Content-Length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return *h
}
