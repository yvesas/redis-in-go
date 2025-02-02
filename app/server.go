package main

import (
	"fmt"
	"net"
	"os"
)

var _ = net.Listen
var _ = os.Exit

func main() {
	host := "0.0.0.0"
	port := 6379
	address := fmt.Sprintf("%s:%d", host, port) // joins string and number into one string

	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("❌ Failed to bind to address %s: %v\n", address, err)
		os.Exit(1)
	}
	defer l.Close() // ??? Listener date when exiting the function
	fmt.Printf("✅ Server listening on address %s\n", address)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("⚠️ Error accepting connection: %v\n", err) // err.Error()
			continue                                               // Keep trying to accept new connections.
		}

		fmt.Println("✅ Accepted connection.")

		_, writeErr := conn.Write([]byte("+PONG\r\n"))
		if writeErr != nil {
			fmt.Printf("⚠️ Failed to answered (write to connection): %v\n", writeErr)
		} else {
			fmt.Println("✅ Answered with success.")
		}

		conn.Close()
	}
}
