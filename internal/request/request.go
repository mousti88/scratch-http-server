package request

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"boot.moustafa.tutorial/internal/headers"
)

type Status int

const (
	initialized    Status = 0
	done           Status = 1
	ParsingHeaders Status = 2
	ParsingBody    Status = 3
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}
type Request struct {
	RequestLine  RequestLine
	HeaderLines  headers.Headers
	Body         []byte
	ParserStatus Status
}

var SEPARATOR = "\r\n"
var BUFFERSIZE = 8

// Chapter4 L3
// creates a new request with default values, which can be used to incrementally parse a request from a stream of data
// returns a pointer to the new request, which can be modified by the caller
func newRequest() *Request {
	v := Request{} //creates a new request with default values
	v.HeaderLines = headers.NewHeaders()
	return &v //returns a pointer to the new request, which can be modified by the caller
	// or return &Request{}
}

func GetIntKey(h *headers.Headers, key string) (int, error) {
	value := h.Get(key)
	if value == "" {
		return 0, nil // no header with that key
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("error converting header value to integer: %w", err)
	}
	return intValue, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {

	buf := make([]byte, BUFFERSIZE)
	readToIndex := 0
	r := newRequest()
	r.ParserStatus = initialized

	//data := make([]byte, 8)
	for r.ParserStatus != done {
		numberBytesRead, readerr := reader.Read(buf[readToIndex:])
		if readerr != nil && readerr != io.EOF {
			return nil, fmt.Errorf("error reading from reader: %w", readerr)
		}
		readToIndex = readToIndex + numberBytesRead

		if readToIndex == len(buf) {
			fmt.Printf("Buffer full, doubling buffer size. Current buffer: %s\n", string(buf))
			OldBuf := buf
			buf = make([]byte, len(OldBuf)*2)
			copy(buf, OldBuf)
		}

		numberBytesParsed, parserr := r.parse(buf[:readToIndex])

		copy(buf[:readToIndex-numberBytesParsed], buf[numberBytesParsed:readToIndex])
		readToIndex = readToIndex - numberBytesParsed

		if parserr != nil {
			return nil, fmt.Errorf("error parsing request: %w", parserr)
		}
		if readerr == io.EOF {
			fmt.Printf("Reached EOF while reading from reader. Current buffer: %s\n", string(buf[:readToIndex]))
			break
		}
	}
	if r.ParserStatus != done {
		return nil, fmt.Errorf("incomplete request: EOF before CRLF")
	}
	return r, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	//The for loop inside parse() handles one type of parsing (request line, headers, or body) for what's currently in the buffer. The outer main loop handles reading more data when you need it.
outer:
	for {
		currentData := data[totalBytesParsed:]
		switch r.ParserStatus {
		case done:
			return 0, fmt.Errorf("error: trying to read from invalid parser state: %d", r.ParserStatus)

		case initialized:
			Parsedline, numBytes, err := parseRequestLine(data)
			totalBytesParsed = totalBytesParsed + numBytes
			if err == nil && numBytes == 0 {
				return 0, nil
			}
			if err != nil {
				return 0, err
			}
			fmt.Printf("Parsed Request Line: %+v\n", Parsedline)
			r.RequestLine = Parsedline
			r.ParserStatus = ParsingHeaders

		case ParsingHeaders:
			numBytesHeaders, doneParsing, err := r.HeaderLines.Parse(currentData)
			totalBytesParsed = totalBytesParsed + numBytesHeaders

			if err != nil {
				fmt.Printf("Error parsing headers: %v\n", err)
				return 0, err
			}
			if doneParsing {
				totalBytesParsed = totalBytesParsed + len(headers.LINE_SEPARATOR) // account for the \r\n that separates headers from body
				r.ParserStatus = ParsingBody
			}
			// Always break after Headers.Parse() to let main loop manage buffer
			break outer

		case ParsingBody:
			ContentLength, err := GetIntKey(&r.HeaderLines, "Content-Length")
			if err != nil {
				return 0, fmt.Errorf("error getting Content-Length header: %w", err)
			}
			if ContentLength == 0 { //  for our implementation we're going to assume that if there is no Content-Length header, there is no body present. This is a safe assumption for our purposes, though it might not be true in all cases in the wild.
				fmt.Printf("No Content-Length header found in request, assuming no body is present. Current buffer: %s\n", string(currentData))
				r.ParserStatus = done
			}
			numBytesParsed, err := parseBody(r, currentData)
			totalBytesParsed = totalBytesParsed + numBytesParsed
			fmt.Printf("Current body content: %s\n", string(r.Body))
			if err != nil {
				return 0, fmt.Errorf("Error while parsing body")
			}
			if len(r.Body) > ContentLength {
				fmt.Printf("Body length exceeds Content-Length header. Body length: %d, Content-Length: %d\n", len(r.Body), ContentLength)
				return 0, fmt.Errorf("error: Body greater than specified content-length")
			}
			if len(r.Body) == ContentLength {
				r.ParserStatus = done
				fmt.Printf("I have consumed all the data you have given me!\n")
			}
			// Always break after parsing body to let main loop manage buffer and EOF
			break outer

		}

	}
	return totalBytesParsed, nil
}

func parseRequestLine(request []byte) (RequestLine, int, error) {
	ParsedRequestLine := RequestLine{}
	requestStr := string(request)
	idx := strings.Index(requestStr, SEPARATOR)
	numBytes := idx + len(SEPARATOR)

	if idx == -1 {
		fmt.Printf("Request line not complete, waiting for more data. Current buffer: %s\n", requestStr)
		return RequestLine{}, 0, nil // no CRLF found, but we need more bytes to determine if it's a malformed request or just incomplete
	}

	fullrequestLine := requestStr[:idx]
	if len(strings.Split(fullrequestLine, " ")) != 3 {
		return RequestLine{}, numBytes, fmt.Errorf("invalid request line: expected 3 parts, got %d", len(strings.Split(fullrequestLine, " ")))
	}
	ParsedRequestLine.Method = strings.Split(fullrequestLine, " ")[0]
	ParsedRequestLine.RequestTarget = strings.Split(fullrequestLine, " ")[1]

	if len(strings.Split(strings.Split(fullrequestLine, " ")[2], "/")) != 2 {
		return RequestLine{}, numBytes, fmt.Errorf("invalid HTTP version format: expected HTTP/x.x, got %s", strings.Split(fullrequestLine, " ")[2])
	}
	ParsedRequestLine.HttpVersion = strings.Split(strings.Split(fullrequestLine, " ")[2], "/")[1]

	if strings.ToUpper(ParsedRequestLine.Method) != ParsedRequestLine.Method {
		return RequestLine{}, numBytes, fmt.Errorf("malformed method: %s", ParsedRequestLine.Method)
	}
	if ParsedRequestLine.HttpVersion != "1.1" {
		return RequestLine{}, numBytes, fmt.Errorf("unsupported HTTP version: %s", ParsedRequestLine.HttpVersion)
	}

	return ParsedRequestLine, numBytes, nil
}

func parseBody(request *Request, body []byte) (int, error) {
	fmt.Printf("Parsing body. Current buffer: %s\n", string(body))
	request.Body = append(request.Body, body...)
	return len(body), nil
}
