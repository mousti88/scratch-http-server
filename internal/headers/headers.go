package headers

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

var LINE_SEPARATOR = []byte("\r\n")
var KEY_VALUE_SEPARATOR = []byte(":")
var EMPTY_FIELD_LINE = []byte("\r\n\r\n")

func NewHeaders() Headers {
	return make(map[string]string)
	//return map[string]string{}
}

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}
func CheckInvalidChar(key string) bool {
	return regexp.MustCompile(`^[A-Za-z0-9!#$%&'*+\-.^_{|}` + "`" + `~]+$`).MatchString(key)
}
func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	if bytes.HasPrefix(data, EMPTY_FIELD_LINE) {
		fmt.Printf("Reached end of headers section. Current buffer: %s\n", string(data))
		return len(EMPTY_FIELD_LINE), true, nil
	}
	var numBytesConsumed int

	for {

		// reached the end of headers section, return the number of bytes consumed, and that we're done parsing headers
		if bytes.HasPrefix(data, LINE_SEPARATOR) {
			fmt.Printf("Reached end of  . Current buffer: %s\n", string(data))
			return numBytesConsumed, false, nil // why false? because we want to allow for the possibility of more headers to be parsed in the future if we want to support multiple header sections (e.g. trailers in chunked encoding), but for now we can just return false since we're only parsing one header section
		}
		crlfIdx := bytes.Index(data, LINE_SEPARATOR) // search for the crlf that separates header lines, if it doesn't exist, we need to wait for more data to determine if it's a malformed header or just incomplete
		if crlfIdx == -1 {
			fmt.Printf("Header line not complete, waiting for more data. Current buffer: %s\n", string(data))
			return 0, false, nil
		}

		idx := bytes.Index(data[:crlfIdx], KEY_VALUE_SEPARATOR) // search for the key value separator in the current header line, if it doesn't exist, it's a malformed header and we can return an error immediately without waiting for more data
		if idx == -1 {
			fmt.Printf("Invalid header format: missing valid separator in header line: %s\n", string(data))
			return 0, false, fmt.Errorf("Invalid header format: missing valid separator")
		}

		key := string(data[:idx])
		if !CheckInvalidChar(key) {
			fmt.Printf("Invalid header format: invalid characters in header key: %s\n", string(data))
			return 0, false, fmt.Errorf("Invalid header format: invalid characters in header key")
		}
		value := strings.TrimSpace(string(data[idx+1 : crlfIdx]))

		emptySpaceIdx := strings.Index(key, " ") // check for empty spaces in the header key, if it exists, it's a malformed header and we can return an error immediately without waiting for more data
		if emptySpaceIdx != -1 {
			fmt.Printf("Invalid header format: empty spaces in header key: %s\n", string(data))
			return 0, false, fmt.Errorf("Invalid header format: empty spaces in header key")
		}
		key = strings.ToLower(key)
		h[key] = value // store the parsed header key value pair in the Headers map
		fmt.Printf("Parsed header: %s: %s\n", key, value)
		//numBytesConsumed = crlfIdx + len(LINE_SEPARATOR)
		data = data[crlfIdx+len(LINE_SEPARATOR):]
		numBytesConsumed += crlfIdx + len(LINE_SEPARATOR)
	}
	//return numBytesConsumed, false, nil
	//return 0, false, nil
}
