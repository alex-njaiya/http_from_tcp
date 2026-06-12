package request

import (
	"bytes"
	"fmt"
	"io"
	"tcp_to_http/internal/headers"
)

// The http parser should return something like this
// type Request struct {
//     RequestLine RequestLine
//     Headers     map[string]string
//     Body        []byte
// }

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	state       parseState
	Headers     headers.Headers
}

type parseState string

const (
	StateInit                  parseState = "init"
	StateDone                  parseState = "done"
	StateError                 parseState = "error"
	StateHeaders 				parseState = "headers"
)

var SEPARATOR = []byte("\r\n")
var MALFORMED_REQUEST_LINE = fmt.Errorf("malformed request-line")
var ErrorUnsupportedHttpVersion = fmt.Errorf("unsupported http version")
var ErrorRequestInErrorState = fmt.Errorf("request in error state")

func newRequest() *Request {
	return &Request{
		state: StateInit,
		Headers: *headers.NewHeaders(),
	}
}

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)

	if idx == -1 {
		return nil, 0, nil
	}

	startline := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startline, []byte(" "))

	if len(parts) != 3 {
		return nil, 0, MALFORMED_REQUEST_LINE
	}

	httpParts := bytes.Split(parts[2], []byte("/"))

	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, read, MALFORMED_REQUEST_LINE
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func (r *Request) parse(data []byte) (int, error) {

	read := 0

outer:
	for {
		currentData := data[read:]
		switch r.state {
		case StateError:
			return 0, ErrorRequestInErrorState
		case StateInit:
			rl, n, err := parseRequestLine(currentData)

			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n

			r.state = StateHeaders
		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)

			if err != nil {
				return 0, err
			}


			if n == 0 {
				break outer
			}

			read += n
			
			if done {
				r.state = StateDone
			}
		case StateDone:
			break outer
		}
	}

	return read, nil
}

func (r Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {

	request := newRequest()

	//create a buffer of 1kb
	buf := make([]byte, 1024)
	bufLen := 0

	for !request.done() {
		n, err := reader.Read(buf[bufLen:])

		if err != nil {
			return nil, err
		}

		bufLen += n
		readN, err := request.parse(buf[:bufLen])

		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}
