package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	listener, err := net.ResolveUDPAddr("udp", ":42069")
	if err != nil {
		log.Fatal("error", "err", err)
	}

	conn, err := net.DialUDP("udp", nil, listener)
	if err != nil {
		log.Fatal("error", "err", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println(">")
		input, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Printf("error: %s\n", err)
			return
		}

		n, err := conn.Write([]byte(input))
		if err != nil {
			fmt.Printf("error: %s\n", err)
		}

		fmt.Println("read", n, "characters")
	}

}
