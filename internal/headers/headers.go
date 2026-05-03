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
func setHeader(h Headers, key, value string) {
	_, ok := h[strings.ToLower(key)]
	if ok {
		fmt.Printf("Header key %s already exists, appending value to existing header value. Current header value: %s, new header value: %s\n", key, h[strings.ToLower(key)], value)
		h[strings.ToLower(key)] = h[strings.ToLower(key)] + ", " + value
	} else {
		h[strings.ToLower(key)] = value
		fmt.Printf("Set header set: %s: %s\n", key, value)
	}
}
func (h Headers) Get(key string) string {
	// return key if it exists, otherwise return error
	value, exists := h[strings.ToLower(key)]
	if !exists {
		return ""
	}
	return value
}
func CheckInvalidChar(key string) bool {
	return regexp.MustCompile(`^[A-Za-z0-9!#$%&'*+\-.^_{|}` + "`" + `~]+$`).MatchString(key)
}

func parseHeader(fieldline []byte) (string, string, error) {

	keyValueIdx := bytes.Index(fieldline, KEY_VALUE_SEPARATOR) // search for the key value separator in the current header line, if it doesn't exist, it's a malformed header and we can return an error immediately without waiting for more data
	if keyValueIdx == -1 {
		fmt.Printf("Invalid header format: missing valid separator in header line: %s\n", string(fieldline))
		return "", "", fmt.Errorf("Invalid header format: missing valid separator")
	}

	key := string(fieldline[:keyValueIdx])
	if !CheckInvalidChar(key) {
		fmt.Printf("Malformed field line: invalid characters in header key: %s\n", string(fieldline))
		return "", "", fmt.Errorf("Malformed field line: invalid characters in header key")
	}
	value := strings.TrimSpace(string(fieldline[keyValueIdx+1:]))

	emptySpaceIdx := strings.Index(key, " ") // check for empty spaces in the header key, if it exists, it's a malformed header and we can return an error immediately without waiting for more data
	if emptySpaceIdx != -1 {
		fmt.Printf("Malformed field name: empty spaces in header value: %s\n", string(fieldline))
		return "", "", fmt.Errorf("Malformed field name: empty spaces in header value")
	}
	key = strings.ToLower(key)
	//h[key] = value // store the parsed header key value pair in the Headers map
	fmt.Printf("Parsed header: %s: %s\n", key, value)
	return key, value, nil
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	if bytes.HasPrefix(data, EMPTY_FIELD_LINE) {
		fmt.Printf("Reached end of headers section. Current buffer: %s\n", string(data))
		return len(EMPTY_FIELD_LINE), true, nil
	}
	var numBytesConsumed int
	var doneParsing bool

	for {

		crlfIdx := bytes.Index(data, LINE_SEPARATOR) // search for the crlf that separates header lines, if it doesn't exist, we need to wait for more data to determine if it's a malformed header or just incomplete
		if crlfIdx == -1 {
			fmt.Printf("Header line not complete, waiting for more data. Current buffer: %s\n", string(data))
			return numBytesConsumed, false, nil // return bytes already consumed, not 0
		}

		// EMPTY HEADER FOUND at the start of the current data, which means we have reached the end of the headers section, return the number of bytes consumed, and that we're done parsing headers
		// This is the same as doing this: if bytes.HasPrefix(data, LINE_SEPARATOR) { // why LINE_SEPARATOR and not EMPTY_FIELD_LINE? because we already removed one registered nurse \r\n LINE_SEPARATOR from the previous header line (in this line of code: data = data[crlfIdx+len(LINE_SEPARATOR):], as you can see it it goes to the crlf index PLUS 2!! which is the first '\r\n' that we encounter), so if we see a LINE_SEPARATOR at the start of the current data, it means we have reached the end of the headers section (i.e. we have seen \r\n\r\n)
		if crlfIdx == 0 {
			doneParsing = true
			break
		}

		key, value, err := parseHeader(data[:crlfIdx])
		if err != nil {
			return 0, false, err
		}
		setHeader(h, key, value)
		data = data[crlfIdx+len(LINE_SEPARATOR):]
		numBytesConsumed += crlfIdx + len(LINE_SEPARATOR)
	}
	return numBytesConsumed, doneParsing, nil
}
