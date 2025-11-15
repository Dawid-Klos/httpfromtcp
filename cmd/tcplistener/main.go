package main

import (
	"fmt"
	"github.com/Dawid-Klos/httpfromtcp/internal/request"
	"log"
	"net"
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

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error", "err:", err)
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.Target)
		fmt.Printf("- Version: %s\n", req.RequestLine.HTTPVersion)
		fmt.Println("Headers:")
		for key, value := range req.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}
		fmt.Println("Body:")
		fmt.Printf("%s\n", string(req.Body))

		fmt.Println("-> Connection to", conn.RemoteAddr(), "closed")
	}
}
