package request

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"unicode"
)

type RequestLine struct {
	HTTPVersion string
	Target      string
	Method      string
}

type Request struct {
	RequestLine RequestLine
	state       parserState
}

type parserState string

const (
	Init  parserState = "init"
	Done  parserState = "done"
	Error parserState = "error"
)

var CRLF = []byte("\r\n")

var ErrRequestInErrorState = errors.New("request in error state")
var ErrUnsuppportedHTTPVersion = errors.New("unsupported HTTP version")
var ErrMalformedRequestLine = errors.New("malformed request line")
var ErrMalformedMethod = errors.New("malformed method in request line")
var ErrMalformedVersion = errors.New("malformed version in request line")
var ErrMalformedTarget = errors.New("malformed target in request line")

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		switch r.state {
		case Error:
			return 0, ErrRequestInErrorState
		case Init:
			rl, n, err := parseRequestLine(data[read:])
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.state = Done
		case Done:
			break outer
		}
	}

	return read, nil
}

func (r *Request) done() bool {
	return r.state == Done || r.state == Error
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{
		state: Init,
	}

	buf := make([]byte, 512)
	bufIdx := 0
	for !request.done() {
		n, err := reader.Read(buf[bufIdx:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				request.state = Done
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
		return nil, len(requestLineText), err
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
