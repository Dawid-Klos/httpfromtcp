package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("error", "err", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error", "err", err)
		}
		fmt.Println("-> Accepted connection from", conn.RemoteAddr())

		for line := range getLinesChannel(conn) {
			fmt.Println(line)
		}

		fmt.Println("-> Connection to", conn.RemoteAddr(), "closed")
	}
}

func getLinesChannel(conn io.ReadCloser) <-chan string {
	lines := make(chan string)

	go func() {
		defer func() {
			defer close(lines)
			err := conn.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()

		chunk := make([]byte, 8)
		var line string
		for {
			n, err := conn.Read(chunk)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %s\n", err)
				return
			}

			parts := strings.Split(string(chunk[:n]), "\n")
			line += parts[0]
			if len(parts) > 1 {
				lines <- line
				line = parts[1]
			}
		}
	}()

	return lines
}
