// Package headers parses HTTP/1.1 headers from a byte slice.
//
// It incrementally reads field-name and field-value pairs until the blank
// line is found that terminates the headers section. Field names are validated
// according to HTTP token rules and normalized to lowercase. Duplicate headers
// are combined into a single, comma-separated value.
package headers

import (
	"bytes"
	"errors"
	"fmt"
)

type Headers map[string]string

func NewHeaders() Headers {
	headers := make(map[string]string, 16)
	return headers
}

var validHeaderChar [256]bool

var CRLF = []byte("\r\n")
var COLON = []byte(":")
var WHITESPACE = []byte(" ")

var ErrMalformedFieldName = errors.New("malformed field-name")
var ErrMalformedFieldValue = errors.New("malformed field-value")

func init() {
	for c := 'a'; c <= 'z'; c++ {
		validHeaderChar[c] = true
	}
	for c := 'A'; c <= 'Z'; c++ {
		validHeaderChar[c] = true
	}
	for c := '0'; c <= '9'; c++ {
		validHeaderChar[c] = true
	}
	for _, c := range "!#$%&'*+-.^_`|~" {
		validHeaderChar[c] = true
	}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	crlfIdx := bytes.Index(data, CRLF)
	if crlfIdx == -1 {
		return 0, false, nil
	}
	if crlfIdx == 0 {
		return 0, true, nil
	}

	header := data[:crlfIdx]
	colonIdx := bytes.Index(header, COLON)

	fieldName, err := h.validateFieldName(header[:colonIdx])
	if err != nil {
		return 0, false, err
	}

	fieldValue, err := h.validateFieldValue(header[colonIdx+2:])
	if err != nil {
		return 0, false, err
	}

	field := string(bytes.ToLower(fieldName))
	value, ok := h[field]
	if !ok {
		h[field] = string(fieldValue)
	} else {
		h[field] = fmt.Sprintf("%s,%s", value, string(fieldValue))
	}

	return crlfIdx + 2, false, nil
}

func (h Headers) validateFieldValue(b []byte) ([]byte, error) {
	fieldValue := bytes.TrimRight(b, string(WHITESPACE))
	if bytes.Contains(fieldValue, WHITESPACE) {
		return nil, ErrMalformedFieldValue
	}

	return fieldValue, nil
}
func (h Headers) validateFieldName(b []byte) ([]byte, error) {
	fieldName := bytes.TrimLeft(b, string(WHITESPACE))

	if bytes.Contains(fieldName, WHITESPACE) {
		return nil, ErrMalformedFieldName
	}

	isValid := isToken(fieldName)
	if !isValid {
		return nil, ErrMalformedFieldName
	}

	return fieldName, nil
}

func isToken(b []byte) bool {
	isValid := true
	for _, char := range b {
		if !validHeaderChar[char] {
			isValid = false
			break
		}
	}

	return isValid
}
