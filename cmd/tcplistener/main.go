package main

import (
	"bytes"
	"fmt"
	"io"
	"log"

	// "os"
	"net"

	"boot.moustafa.tutorial/internal/request"
)

// Q4&5
// returns a channel that the for loop in main will immediately start reading from as lines are available
// the channel is immediately returned, because/and we start a goroutine to read from the file and send lines to the channel
func getLinesChannel(f io.ReadCloser) <-chan string {

	channelOutput := make(chan string)

	go func() {
		defer f.Close()            // only runs when goroutine returns eof or error
		defer close(channelOutput) // only runs when goroutine returns eof or error

		stringOutput := ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)

			dataMeaningful := data[:n]

			if i := bytes.IndexByte(dataMeaningful, '\n'); i != -1 {
				stringOutput += string(dataMeaningful[:i])
				dataMeaningful = dataMeaningful[i+1:]
				//fmt.Printf("read: %s\n", stringOutput)
				channelOutput <- stringOutput
				stringOutput = ""
			}
			stringOutput += string(dataMeaningful)

			if err == io.EOF {
				break
			}
		}
		if len(stringOutput) != 0 {
			channelOutput <- stringOutput
		}
	}()

	return channelOutput
}

func main() {
	listener, err := net.Listen("tcp", "localhost:42069")
	if err != nil {
		log.Fatal("error", "error listening for TCP traffic: %s\n", err.Error())
	}
	defer listener.Close() //Now no matter what happens:✅ error✅ return✅ panic✅ success it closes.
	for {
		conn, err := listener.Accept()
		fmt.Printf("connection has been accepted.\n")

		requestParsed, err := request.RequestFromReader(conn)
		fmt.Printf("read line: \n")
		fmt.Printf("- Method: %s\n", string(requestParsed.RequestLine.Method))
		fmt.Printf("- Target: %s\n", string(requestParsed.RequestLine.RequestTarget))
		fmt.Printf("- Version: %s\n", string(requestParsed.RequestLine.HttpVersion))
		fmt.Printf("- Headers:\n")
		for key, value := range requestParsed.HeaderLine {
			fmt.Printf("  - %s: %s\n", key, value)
		}
		if err != nil {
			break
		}
		fmt.Printf("Listener has been closed!")
	}
}
