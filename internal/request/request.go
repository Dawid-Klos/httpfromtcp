// Package request parses HTTP/1.1 requests from a stream.
//
// It reads and validates the request line, then incrementally parses header
// fields until the request is complete. The parser is stateful and supports
// partial reads from the provided io.Reader.
package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/Dawid-Klos/httpfromtcp/internal/headers"
)

type RequestLine struct {
	HTTPVersion string
	Target      string
	Method      string
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       parserState
}

type parserState string

const (
	requestStateInit           parserState = "init"
	requestStateParsingHeaders parserState = "parsing headers"
	requestStateDone           parserState = "done"
	requestStateError          parserState = "error"
)

var CRLF = []byte("\r\n")

var ErrRequestInErrorState = errors.New("request in error state")
var ErrUnsuppportedHTTPVersion = errors.New("unsupported HTTP version")
var ErrMalformedRequestLine = errors.New("malformed request line")
var ErrMalformedMethod = errors.New("malformed method in request line")
var ErrMalformedVersion = errors.New("malformed version in request line")
var ErrMalformedTarget = errors.New("malformed target in request line")
var ErrReadingDataInDoneState = errors.New("trying to read data in done state")
var ErrUnknownRequestState = errors.New("unknown request state")

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{
		state:   requestStateInit,
		Headers: headers.NewHeaders(),
	}

	buf := make([]byte, 512)
	bufIdx := 0
	for !request.done() {
		n, err := reader.Read(buf[bufIdx:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state: %s, read n bytes on EOF: %d", request.state, bufIdx)
				}
				//request.state = requestStateDone
				break
			}
			return nil, err
		}

		bufIdx += n
		readN, err := request.parse(buf[:bufIdx])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufIdx])
		bufIdx -= readN
	}

	return request, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, CRLF)
	if idx == -1 {
		return nil, 0, nil
	}

	requestLineText := string(data[:idx])
	read := idx + len(CRLF)
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, idx + 2, err
	}

	return requestLine, read, nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, ErrMalformedRequestLine
	}

	method := parts[0]
	if strings.ToUpper(method) != method {
		return nil, ErrMalformedMethod
	}

	for _, letter := range method {
		if !unicode.IsLetter(letter) {
			return nil, ErrMalformedMethod
		}
	}

	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, ErrMalformedRequestLine
	}
	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, ErrUnsuppportedHTTPVersion
	}
	version := versionParts[1]
	if version != "1.1" {
		return nil, ErrUnsuppportedHTTPVersion
	}

	target := parts[1]
	if !strings.Contains(target[:1], "/") {
		return nil, ErrMalformedTarget
	}
	if strings.Contains(target, " ") {
		return nil, ErrMalformedTarget
	}

	return &RequestLine{
		HTTPVersion: version,
		Target:      target,
		Method:      method,
	}, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalParsedBytes := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalParsedBytes:])
		if err != nil {
			return 0, err
		}
		totalParsedBytes += n
		if n == 0 {
			//r.state = requestStateDone
			break
		}
	}

	return totalParsedBytes, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInit:
		rl, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *rl
		r.state = requestStateParsingHeaders
		return n, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = requestStateDone
		}
		return n, nil
	case requestStateDone:
		return 0, ErrRequestInErrorState
	case requestStateError:
		return 0, ErrRequestInErrorState
	default:
		return 0, ErrUnknownRequestState
	}
}

func (r *Request) done() bool {
	return r.state == requestStateDone || r.state == requestStateError
}
