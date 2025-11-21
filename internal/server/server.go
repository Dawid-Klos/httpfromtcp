// Package server handles requests and sends responses.
package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/Dawid-Klos/httpfromtcp/internal/request"
	"github.com/Dawid-Klos/httpfromtcp/internal/response"
)

type Server struct {
	closed   atomic.Bool
	listener net.Listener
	handler  Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%d", port))
	if err != nil {
		log.Fatal("error", "err", err)
	}

	s := &Server{
		listener: listener,
		handler:  handler,
	}

	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	if err := s.listener.Close(); err != nil {
		return err
	}
	s.closed.Store(false)
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if !s.closed.Load() {
				return
			}
			log.Fatal("error", "err", err)
			continue
		}

		fmt.Println("-> Accepted connection from", conn.RemoteAddr())
		go s.handle(conn)
		fmt.Println("-> Connection to", conn.RemoteAddr(), "closed")
	}
}

func (s *Server) handle(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println("error", "err", err)
			return
		}
	}()

	request, err := request.RequestFromReader(conn)
	if err != nil {
		handlerErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		if err := handlerErr.Write(conn); err != nil {
			log.Fatal("error", "err", err)
		}
		return
	}

	buf := new(bytes.Buffer)
	handlerErr := s.handler(buf, request)
	if handlerErr != nil {
		if err := handlerErr.Write(conn); err != nil {
			log.Println("error", "err", err)
		}
		return
	}

	if err := response.WriteStatusLine(conn, 200); err != nil {
		log.Println("error", "err", err)
	}

	headers := response.GetDefaultHeaders(buf.Len())
	if err := response.WriteHeaders(conn, headers); err != nil {
		log.Println("error writing headers:", err)
	}

	if _, err = buf.WriteTo(conn); err != nil {
		log.Println("error writing body", err)
	}
}

func (hErr *HandlerError) Write(w io.Writer) error {
	statusLine := "HTTP/1.1" + " " + strconv.Itoa(int(hErr.StatusCode)) + " " + hErr.Message + "\r\n"
	if _, err := w.Write([]byte(statusLine)); err != nil {
		return err
	}

	headers := response.GetDefaultHeaders(0)
	if err := response.WriteHeaders(w, headers); err != nil {
		return err
	}
	return nil
}
