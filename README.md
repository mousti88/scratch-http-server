# scratch-http-server
Building my own HTTP server from scratch in Golang using the HTTP/1.1 protocol. Why would I do that? There is no better way to understand how something works than to implement it myself. 

# scratch-http-server

A minimal HTTP/1.1 server implemented from scratch in Go to learn how request parsing, buffering and streaming I/O work.

## What this is

This repository contains a small, educational HTTP/1.1 request parser and a tiny TCP listener that uses it. It's intended for learning and experimentation — not production use.

## Repository layout

- internal/request/ - incremental request parser and types
  - internal/request/request.go
  - internal/request/request_test.go
- cmd/tcplistener/ - simple TCP server that parses incoming request lines
- cmd/udpsender/ - small sender used for testing (sends test lines to the listener)
- messages.txt - example input lines used during development
- go.mod - module file

## Quick start

Build everything:

```sh
go build ./...
```
Run the TCP listener (in one terminal):

```sh
go run ./cmd/tcplistener
```

Run the UDP sender (in another terminal) to send test request lines:
go run ./cmd/udpsender

The sender transmits test HTTP request lines to the listener so you can see the parser handle incremental/chunked input.

## Tests
Run parser unit tests:
```sh
go test ./internal/request -v
```

The tests intentionally simulate chunked reads to verify the parser's incremental behavior.

## Design notes
The parser reads from an io.Reader into an internal buffer (see BUFFERSIZE in internal/request/request.go).
It finds the request-line terminated by CRLF (\r\n), validates method and HTTP version, and exposes a Request type.
Public entrypoint: request.RequestFromReader
The code focuses on the request-line and incremental parsing; header parsing and bodies are intentionally not implemented yet.

## Next steps / contributions
Ideas for improvement:

Add header parsing and body handling.
Harden error handling and add more tests for edge cases.
Add benchmarks for different chunking patterns.
Add examples demonstrating persistent connections and pipelining.
If you want to contribute, open issues or PRs with focused changes and tests.

## License
Educational code — no formal license included. Use and adapt for learning purposes.