package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
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

		for line := range getLinesChannel(conn) {
			fmt.Printf("read: %s\n", line)
		}
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {

	out := make(chan string, 1)

	go func() {
		defer f.Close()
		defer close(out)

		// a string variable to persist the read data
		str := "" // holds the current line

		for {
			data := make([]byte, 8)
			n, err := f.Read(data)

			if err == io.EOF {
				break
			}

			data = data[:n]

			if i := bytes.IndexByte(data, '\n'); i != -1 {
				str += string(data[:i])

				data = data[i+1:]
				out <- str
				str = ""
			}

			str += string(data)
		}

		if len(str) != 0 {
			out <- str
		}
	}()

	return out
}
