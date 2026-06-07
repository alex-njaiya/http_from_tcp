package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	address, err := net.ResolveUDPAddr("udp", "127.0.0.1:42069")

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// prepare udp connection and defer closing of the connection
	conn, err := net.DialUDP("udp", nil, address)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	defer conn.Close()

	go func() {
		buf := make([]byte, 8)

		for {
			n, err := conn.Read(buf)

			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Received: %s\n", buf[:n])
		}

	}()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(">")
		line, err := reader.ReadString('\n')

		if err != nil {
			log.Fatalf("error: %v", err)
		}

		_, err = conn.Write([]byte(line))

		if err != nil {
			log.Fatalf("error: %v", err)
		}
	}


}
