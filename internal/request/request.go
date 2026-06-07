package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
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
}

var SEPARATOR = "\r\n"
var MALFORMED_REQUEST_LINE = fmt.Errorf("malformed request-line")


func parseRequestLine(b string) (*RequestLine, string, error) {
	idx := strings.Index(b, SEPARATOR)

	if idx == -1 {
		return nil, b, nil
	}

	startline := b[:idx]
	restOfMsg := b[idx + len(SEPARATOR):]

	parts := strings.Split(startline,  " ")

	if len(parts) != 3 {
		return nil, restOfMsg, MALFORMED_REQUEST_LINE
	}

	httpParts := strings.Split(parts[2], "/")

	if len(httpParts) != 2 || httpParts[0] != "HTTP" || httpParts[1] != "1.1" {
		return nil, restOfMsg, MALFORMED_REQUEST_LINE
	}

	rl := &RequestLine{
		Method: parts[0],
		RequestTarget: parts[1],
		HttpVersion: httpParts[1],
	}

	return rl, restOfMsg, nil
}


func RequestFromReader(reader io.Reader) (*Request, error) {
	//read the entire file parse later
	data, err := io.ReadAll(reader)

	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("unable to read io.readAll"), 
			err,
		)
	}

	str := string(data)

	rl, _, err := parseRequestLine(str)

	if err != nil {
		return nil, err
	}

	if rl == nil {
		return nil, fmt.Errorf("parsed request line is nill")
	}

	return &Request{
		RequestLine: *rl,
	}, err
}