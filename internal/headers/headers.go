package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers struct {
	headers map[string]string
}
type ParsedState string

var rn = []byte("\r\n")

func isToken(str []byte) bool {
	for _, ch := range str {
		found := false
		if ch >= 'A' && ch <= 'Z' ||
			ch >= 'a' && ch <= 'z' ||
			ch >= '0' && ch <= '9' {
				found = true
			}

			switch ch {
			case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
				found = true
			}

			if !found {
				return false
			}
	}
	return true
}

func (h *Headers) Get(name string) string {
	return h.headers[strings.ToLower(name)]
}

func (h *Headers) Set(name, value string) {
	h.headers[strings.ToLower(name)] = value
}

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed field line")
	}

	name := parts[0]

	if len(name) == 0 {
		return "", "", fmt.Errorf("empty field name")
	}

	if name[0] == ' ' || name[0] == '\t' {
		return "", "", fmt.Errorf("Invalid whitespace in the field name")
	}
	value := bytes.TrimSpace(parts[1])

	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", fmt.Errorf("malformed field name")
	}

	return string(name), string(value), nil
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	read := 0

	done = false

	for {
		idx := bytes.Index(data, rn)

		if idx == -1 {
			break
		}

		//empty header --
		if idx == 0 {
			done = true
			break
		}

		line := data[:idx]
		name, value, err := parseHeader(line)

		if err != nil {
			return 0, false, err
		}

		if !isToken([]byte(name)) {
			return 0, false, fmt.Errorf("malformed field name")
		}

		h.Set(name, value)
		consumed := idx + len(rn)
		read += consumed
		data = data[consumed:]

	}

	return read, done, nil
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}
