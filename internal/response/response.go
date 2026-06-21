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

func (w *Writer) WriteHeaders(headers headers.Headers) error {
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
	h.Set("Transfer-Encoding", "chunked")

	return *h
}




// FIX THIS 2 FUNCTIONS
// func (w *Writer) WriteChunkedBody(b []byte) (int, error) {
// 	// Convert that byte count to hexadecimal text
// 	size := strconv.FormatUint(uint64(len(b)), 16)

// 	// write the chunk-size line
// 	_, err := w.writer.Write([]byte(size + "\r\n"))

// 	if err != nil {
// 		return 0, err
// 	}

// 	// Write body and handle partial writes
// 	offset := 0
// 	for offset < len(b) {
// 		n, err := w.writer.Write(b[offset:])

// 		if err != nil {
// 			return offset, err
// 		}
// 		offset += n
// 	}

// 	// Write the trailing crlf
// 	_, err = w.writer.Write([]byte("\r\n"))

// 	if err != nil {
// 		return 0, err
// 	}
// 	// Return how many payload bytes were processed and an error
// 	return len(b), nil
// }

// func (w *Writer) WriteChunkedBodyDone() (int, error) {
// 	n1, err := w.writer.Write([]byte("0\r\n"))

// 	if err != nil {
// 		return 0, err
// 	}

// 	n2, err := w.writer.Write([]byte("\r\n"))
// 	if err != nil {
// 		return 0, err
// 	}

// 	return n1 + n2, nil
// }
