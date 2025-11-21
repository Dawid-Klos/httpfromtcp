package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Dawid-Klos/httpfromtcp/internal/request"
	"github.com/Dawid-Klos/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(42069, myHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer func() {
		err := server.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func myHandler(w io.Writer, req *request.Request) *server.HandlerError {
	var handlerErr *server.HandlerError
	var message []byte

	switch req.RequestLine.Target {
	case "/yourproblem":
		handlerErr = &server.HandlerError{
			StatusCode: 400,
			Message:    "Bad Request",
		}
		message = []byte("Your problem is not my problem\n")
	case "/myproblem":
		handlerErr = &server.HandlerError{
			StatusCode: 500,
			Message:    "Internal Server Error",
		}
		message = []byte("Woopsie, my bad!\n")
	default:
		message = []byte("All good, frfr\n")
	}

	if _, err := w.Write(message); err != nil {
		handlerErr = &server.HandlerError{
			StatusCode: 500,
			Message:    "Internal Server Error",
		}
	}

	return handlerErr
}
