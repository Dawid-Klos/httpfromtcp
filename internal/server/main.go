// Package server handles requests and sends responses.
package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/Dawid-Klos/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	isOpen   atomic.Bool
}

func Serve(port int) (*Server, error) {
	var srv Server
	srv.isOpen.Store(true)

	go func() {
		listener, err := net.Listen("tcp", ":42069")
		if err != nil {
			log.Fatal("error", "err", err)
		}
		defer func() {
			err := listener.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()

		srv.listener = listener
		srv.listen()
	}()

	return &srv, nil
}

func (s *Server) Close() error {
	if err := s.listener.Close(); err != nil {
		return err
	}
	s.isOpen.Store(false)
	return nil
}

func (s *Server) listen() {
	if !s.isOpen.Load() {
		return
	}

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Fatal("error", "err", err)
		}
		go func() {
			fmt.Println("-> Accepted connection from", conn.RemoteAddr())
			s.handle(conn)
			fmt.Println("-> Connection to", conn.RemoteAddr(), "closed")
		}()
	}
}

func (s *Server) handle(conn net.Conn) {
	//message := "Hello World!"
	headers := response.GetDefaultHeaders(0)

	if err := response.WriteStatusLine(conn, 200); err != nil {
		log.Fatal("error", "err", err)
	}
	if err := response.WriteHeaders(conn, headers); err != nil {
		log.Fatal("error", "err", err)
	}

	if err := conn.Close(); err != nil {
		log.Fatal("error", "err", err)
	}
}
