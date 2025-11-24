package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"boot.thehemantgupta.tv/httpfromtcp/internal/headers"
)

type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
	StateDone    parserState = "done"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        string
	state       parserState
}

func getInt(headers *headers.Headers, name string, defaultValue int) int {
	valueStr, exists := headers.Get(name)
	if !exists {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
		Body:    "",
	}
}

var MALFORMED_CONTENT = fmt.Errorf("malformed request-line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("Unsupported HTTP version")
var BAD_START_LINE = fmt.Errorf("bad start line")
var SEPARATOR = []byte("\r\n")

// func parseRequestLine(b string) (*RequestLine, error) {
// 	idx := strings.Index(b, SEPARATOR)
// 	if idx == -1 {
// 		return nil, nil
// 	}
// 	startLine := b[:idx]
// 	// restMsg := b[idx+len(SEPARATOR):];

// 	parts := strings.Split(startLine, " ")
// 	if len(parts) != 3 {
// 		return nil, MALFORMED_CONTENT
// 	}

// 	httpParts := strings.Split(parts[2], "/")
// 	if len(httpParts) != 2 || httpParts[0] != "HTTP" || httpParts[1] != "1.1" {
// 		return nil, MALFORMED_CONTENT
// 	}

// 	rl := &RequestLine{
// 		Method:        parts[0],
// 		RequestTarget: parts[1],
// 		HttpVersion:   httpParts[1],
// 	}

// 	return rl, nil
// }

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}
	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, MALFORMED_CONTENT
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, MALFORMED_CONTENT
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func (r *Request) hasBody() bool {
	length := getInt(r.Headers, "content-length", 0)
	return length > 0
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}
		switch r.state {
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.state = StateHeaders
		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateDone
				return 0, err
			}
			if n == 0 {
				break outer
			}
			read += n
			if done {
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}
		case StateBody:
			length := getInt(r.Headers, "content-length", 0)
			if length == 0 {
				r.state = StateDone
				break
			}
			remaining := min(length-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining
			if len(r.Body) == length {
				r.state = StateDone
			}
		case StateDone:
			break outer
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone
}
func RequestFromReader(reader io.Reader) (*Request, error) {

	request := newRequest()

	buf := make([]byte, 1024)
	bufLen := 0

	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}
	return request, nil

	// data, err := io.ReadAll(reader)
	// if err != nil {
	// 	return nil, err
	// }

	// str := string(data)
	// rl, err := parseRequestLine(str)

	// if err != nil {
	// 	return nil, err
	// }

	// return &Request{
	// 	RequestLine: *rl,
	// }, nil
}
