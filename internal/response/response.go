// Package response helps to create HTTP/1.1 responses.
//
// It includes helpers for writing status line, getting default headers and
// writing headers into the HTTP response.
package response

import (
	"io"
	"strconv"

	"github.com/Dawid-Klos/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var reason string
	switch statusCode {
	case StatusOK:
		reason = "StatusOK"
	case StatusBadRequest:
		reason = "Bad Request"
	case StatusInternalServerError:
		reason = "Internal Server Error"
	}

	statusLine := "HTTP/1.1" + " " + strconv.Itoa(int(statusCode)) + " " + reason + "\r\n"
	if _, err := w.Write([]byte(statusLine)); err != nil {
		return err
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers["Content-Length"] = strconv.Itoa(contentLen)
	headers["Connection"] = "close"
	headers["Content-Type"] = "text/plain"

	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		header := key + ": " + value + "\r\n"
		if _, err := w.Write([]byte(header)); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte("\r\n")); err != nil {
		return err
	}
	return nil
}
