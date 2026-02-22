package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

// func TestRequestLineParse(t *testing.T) {
// 	//assert.Equal(t, "TheTestagen", "TheTestagen")
// 	// Test: Good GET Request line
// 	r, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
// 	require.NoError(t, err)
// 	require.NotNil(t, r)
// 	assert.Equal(t, "GET", r.RequestLine.Method)
// 	assert.Equal(t, "/", r.RequestLine.RequestTarget)
// 	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

// 	// Test: Good GET Request line with path
// 	r, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
// 	require.NoError(t, err)
// 	require.NotNil(t, r)
// 	assert.Equal(t, "GET", r.RequestLine.Method)
// 	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
// 	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

//		// Test: Invalid number of parts in request line
//		_, err = RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
//		require.Error(t, err)
//	}
func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	t.Logf("Test 1 start: data=%q numBytesPerRead=%d", reader.data, reader.numBytesPerRead)
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	t.Logf("Test 1 parsed: method=%q target=%q version=%q", r.RequestLine.Method, r.RequestLine.RequestTarget, r.RequestLine.HttpVersion)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	t.Logf("Test 2 start: data=%q numBytesPerRead=%d", reader.data, reader.numBytesPerRead)
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	t.Logf("Test 2 parsed: method=%q target=%q version=%q", r.RequestLine.Method, r.RequestLine.RequestTarget, r.RequestLine.HttpVersion)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}
