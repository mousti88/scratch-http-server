package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")

	if err != nil {
		log.Fatal("error", "error", err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error dialing UDP: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf(">")
		text, err := reader.ReadString('\n')

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}
		byteText := []byte(text)
		nBytes, err := conn.Write(byteText)
		fmt.Printf("sent %d bytes\n", nBytes)
		if err != nil {
			log.Fatal("error", "error", err)
		}
		fmt.Printf("Message sent: %s", text)
	}

}
