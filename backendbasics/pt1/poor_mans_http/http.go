// artisanal http library
package main

import (
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Header represents an HTTP header. An HTTP header is a key-value pair, separated by a colon (:);
// the key should be formatted in Title-Case.
// Use Request.AddHeader() or Response.AddHeader() to add headers to a request or response and
// guarantee title-casing of the key
type Header struct{ Key, Value string }

// Request represents an HTTP 1.1 request
type Request struct {
	Method  string   // GET, POST, PUT, DELETE
	Path    string   // e.g. /index.html
	Headers []Header // e.g. Host: www.google.com
	Body    string   // e.g. {"some": "json"}
}

type Response struct {
	StatusCode int      // e.g. 200
	Headers    []Header // e.g. Content-Type: text/html
	Body       string   // e.g. <html><body><h1>hello, world!</h1></body></html>
}

func NewRequest(method, path, host, body string) (*Request, error) {
	switch {
	case method == "":
		return nil, errors.New("missing required argument: method")
	case path == "":
		return nil, errors.New("missing required argument: path")
	case !strings.HasPrefix(path, "/"):
		return nil, errors.New("path must start with /")
	case host == "":
		return nil, errors.New("missing required argument: host")
	default:
		headers := make([]Header, 2)
		headers[0] = Header{"Host", host}
		if body != "" {
			headers = append(headers, Header{"Content-Length", fmt.Sprintf("%d", len(body))})
		}
		return &Request{Method: method, Path: path, Headers: headers, Body: body}, nil
	}
}

func NewResponse(status int, body string) (*Response, error) {
	switch {
	case status < 100 || status > 599:
		return nil, errors.New("invalid status code")
	default:
		if body == "" {
			body = http.StatusText(status)
		}
		//// notice the syntax for initializing slices of structs. Slice (or array) initialization follows [n<optional>]T{el1, el2, ..., eln}
		//// and since the header struct has two string fields, Key and Value, those are also wrapped in {}, resulting in
		//// []T{{field1, ..., fieldN}} or []T{{Field1: value1, ..., FieldN: valueN}}
		headers := []Header{
			{"Content-Length", fmt.Sprintf("%d", len(body))},
		}
		return &Response{StatusCode: status, Headers: headers, Body: body}, nil
	}
}

func (resp *Response) WithHeader(key, value string) *Response {
	resp.Headers = append(resp.Headers, Header{AsTitle(key), value})
	return resp
}

func (r *Request) WithHeader(key, value string) *Request {
	r.Headers = append(r.Headers, Header{AsTitle(key), value})
	return r
}

// returns a given header key in Title-Case
func AsTitle(key string) string {
	// empty keys are probably programmer errors, so we panic
	if key == "" {
		panic("empty header key")
	}
	if isTitleCase(key) {
		return key
	}
	// allocation in expensive but iterating through strings is cheap
	// so it's better to check twice than allocate once
	return newTitleCase(key)
}

func newTitleCase(key string) string {
	var b strings.Builder
	b.Grow(len(key))
	for i := range key {
		if i == 0 || key[i-1] == '-' {
			b.WriteByte(upper(key[i]))
		} else {
			b.WriteByte(lower(key[i]))
		}
	}
	return b.String()
}

// K&R C, 2nd edition, page 43
// // and as in K&R C, we're dealing with ASCII (as header keys should be)
func lower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 'a' - 'A'
	}
	return c
}

func upper(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c + 'A' - 'a'
	}
	return c
}

// isTitleCase returns true if the given header key is already title case;
// i.e. it is of the form "Content-Type" or "Content-Length", "Some-Odd-Header", etc
func isTitleCase(key string) bool {
	for i := range key {
		//// nice condition. We won't go out of bounds because the OR will be true on i == 0
		//// so the second condition won't be checked in the first iteration.
		//// Not only that, we also get to check hyphenated-keys-like-this with the same condition
		//// if the last character was a hyphen, this current one must be a capital letter
		if i == 0 || key[i-1] == '-' {
			if key[i] >= 'a' && key[i] <= 'z' {
				return false
			}
		} else if key[i] >= 'A' && key[i] <= 'Z' {
			return false
		}
	}
	return true
}

// Write writes the Request to the given io.Writer
func (r *Request) WriteTo(w io.Writer) (n int64, err error) {
	// write & count bytes written
	// using small closures like this to cut down on repetition can be nice
	// but you sometimes pay a performance penalty
	printf := func(format string, args ...any) error {
		//// err is being shadowed here but we return it in the end and will assign to the err returned to WriteTo
		m, err := fmt.Fprintf(w, format, args...)
		//// and same as err, n was declared when it was used as a return value, notice we never
		//// initialize it in this function
		n += int64(m)
		return err
	}
	// remember, an HTTP request looks like this:
	// <METHOD>  <PATH>   <PROTOCOL/VERSION>
	// <HEADER>: <VALUE>
	// <HEADER>: <VALUE>
	//
	// <REQUEST BODY>

	// write the request line: like "GET /index.html HTTP/1.1"
	if err := printf("%s %s HTTP/1.1\r\n", r.Method, r.Path); err != nil {
		return n, err
	}

	// write the headers. We don't do anything to order them or combine/merge duplicate headers
	// this is just an exmaple
	for _, h := range r.Headers {
		if err := printf("%s: %s\r\n", h.Key, h.Value); err != nil {
			return n, err
		}
	}
	printf("\r\n")
	err = printf("%s\r\n", r.Body)
	return n, err
}

func (resp *Response) WriteTo(w io.Writer) (n int64, err error) {
	printf := func(format string, args ...any) error {
		m, err := fmt.Fprintf(w, format, args...)
		n += int64(m)
		return err
	}

	if err := printf("HTTP/1.1 %d %s\r\n", resp.StatusCode, http.StatusText(resp.StatusCode)); err != nil {
		return n, err
	}
	for _, h := range resp.Headers {
		if err := printf("%s: %s\r\n", h.Key, h.Value); err != nil {
			return n, err
		}
	}
	if err := printf("\r\n%s\r\n", resp.Body); err != nil {
		return n, err
	}
	return n, nil
}

// compile-time check that both Request and Response implement fmt.Stringer
var _, _ fmt.Stringer = (*Request)(nil), (*Response)(nil)

// same for encoding.TextMarshaler
var _, _ encoding.TextMarshaler = (*Request)(nil), (*Response)(nil)

func (r *Request) String() string     { b := new(strings.Builder); r.WriteTo(b); return b.String() }
func (resp *Response) String() string { b := new(strings.Builder); resp.WriteTo(b); return b.String() }

func (r *Request) MarshalText() ([]byte, error) {
	b := new(bytes.Buffer)
	r.WriteTo(b)
	return b.Bytes(), nil
}
func (resp *Response) MarshalText() ([]byte, error) {
	b := new(bytes.Buffer)
	resp.WriteTo(b)
	return b.Bytes(), nil
}

// ParseRequest parses an HTTP request from the given text.
func ParseRequest(raw string) (r Request, err error) {
	// request has three parts:
	// 1. Request line
	// 2. Headers
	// 3. Body (optional)
	lines := splitLines(raw)

	if len(lines) < 3 {
		return Request{}, fmt.Errorf("malformed request: should have at least 3 lines")
	}

	first := strings.Fields(lines[0])
	r.Method, r.Path = first[0], first[1]

	if !strings.HasPrefix(r.Path, "/") {
		return Request{}, fmt.Errorf("malformed request: path should start with /")
	}
	if !strings.Contains(first[2], "HTTP") {
		return Request{}, fmt.Errorf("malformed request: first line should contain HTTP version")
	}

	var foundhost bool
	var bodyStart int

	// first line is evaluated and validated, checking headers up until an empty line
	for i := 1; i < len(lines); i++ {
		if lines[i] == "" {
			bodyStart = i + 1
			break
		}
		key, val, ok := strings.Cut(lines[i], ": ")
		if !ok {
			return Request{}, fmt.Errorf("malformed request: header %q should be of form 'key: value'", lines[i])
		}
		// special case, Host is required
		if key == "Host" {
			foundhost = true
		}
		key = AsTitle(key)

		r.Headers = append(r.Headers, Header{key, val})
	}

	if !foundhost {
		return Request{}, fmt.Errorf("malformed request: missing Host header")
	}

	// skip last empty line
	end := len(lines) - 1
	r.Body = strings.Join(lines[bodyStart:end], "\r\n")

	return r, nil
}

// ParseResponse parses the given HTTP/1.1 response string into the Response.
// It returns an error if the Response is invalid, meaning:
// - not a valid integer
// - invalid status code
// - missing status code
// - invalid headers
// it doesn't properly handle multi-line headers, headers with multiple values, or html-encoding etc.
func ParseResponse(raw string) (resp *Response, err error) {
	// Three parts:
	// 1. Response line
	// 2. Headers
	// 3. Body (optional)
	lines := splitLines(raw)

	first := strings.SplitN(lines[0], " ", 3)
	if !strings.Contains(first[0], "HTTP") {
		return nil, fmt.Errorf("malformed response: first line should contain HTTP version")
	}
	resp = new(Response)
	resp.StatusCode, err = strconv.Atoi(first[1])
	if err != nil {
		return nil, fmt.Errorf("malformed response: expected status code to be an integer, got %q", first[1])
	}
	if first[2] == "" || http.StatusText(resp.StatusCode) != first[2] {
		log.Printf(
			"missing or incorrect status text for status code %d: expected %q, but got %q",
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			first[2],
		)
	}

	var bodyStart int
	for i := 1; i < len(lines); i++ {
		if lines[i] == "" {
			bodyStart = i + 1
			break
		}
		key, val, ok := strings.Cut(lines[i], ": ")
		if !ok {
			return nil, fmt.Errorf("malformed response: header %q should be of form 'key: value", lines[i])
		}
		key = AsTitle(key)
		resp.Headers = append(resp.Headers, Header{key, val})
	}
	resp.Body = strings.TrimSpace(strings.Join(lines[bodyStart:], "\r\n"))
	return resp, nil
}

// splitLines on the "\r\n" sequence; multiple separators in a row are NOT collapsed
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	i := 0
	for {
		j := strings.Index(s[i:], "\r\n")
		if j == -1 {
			lines = append(lines, s[i:])
			return lines
		}
		// up to but not including the \r\n
		lines = append(lines, s[i:i+j])
		// skip the \r\n
		i += j + 2
	}
}
