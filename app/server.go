package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
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
	defer l.Close()
	fmt.Printf("✅ Server listening on address %s\n", address)

	store := NewInMemoryStore()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("⚠️ Error accepting connection: %v\n", err) // err.Error()
			continue                                               // Keep trying to accept new connections.
		}

		go handleConnection(conn, store)
	}
}

func handleConnection(conn net.Conn, store *InMemoryStore) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	commands := map[string]func(net.Conn, []string, *InMemoryStore){
		"PING": func(c net.Conn, args []string, store *InMemoryStore) { sendReply(c, "PONG") },
		"ECHO": runEcho,
		"LIST": runListKeys,
		"DEL":  runDelete,
		"SET":  runSet,
		"GET":  runGet,
	}

	for {
		fmt.Println("-------")
		line, err := reader.ReadString('\n')
		if err != nil {
			// fmt.Println("⚠️ Client disconnected.")
			return
		}

		args, err := parseRESP(line, reader)
		if err != nil {
			fmt.Println("❌ Error parsing request:", err)
			conn.Write([]byte("❌ Invalid request\r\n"))
			continue
		}

		fmt.Println("📥 Received args:", args)
		if len(args) == 0 {
			conn.Write([]byte("❌ Empty command\r\n"))
			continue
		}

		if len(args) > 0 {
			command := strings.ToUpper(args[0])
			if handler, exists := commands[command]; exists {
				handler(conn, args, store)
			} else {
				conn.Write([]byte("❌ Unknown command\r\n"))
				fmt.Println("❌ Unknown command:", args)
			}
		}
	}
}

func sendReply(conn net.Conn, message string, data ...interface{}) {
	var response string

	if len(data) == 0 {
		response = fmt.Sprintf("+%s\r\n", message)
	} else {
		switch v := data[0].(type) {
		case string:
			response = fmt.Sprintf("+%s %s\r\n", message, v)
		case int:
			response = fmt.Sprintf("+%s %d\r\n", message, v)
		case []string:
			response = fmt.Sprintf("+%s %v\r\n", message, strings.Join(v, ", "))
		case map[string]string:
			var items []string
			for key, value := range v {
				items = append(items, fmt.Sprintf("[%s] -> %s", key, value))
			}
			response = fmt.Sprintf("+%s %s\r\n", message, strings.Join(items, ", "))
		default:
			response = "❌ Error in sendReply - Unsupported data type\r\n"
		}
	}

	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("❌ Error sending response:", err)
		return
	}
	fmt.Println("✅ Answered with success. Sent: ", message)
}

func runEcho(conn net.Conn, args []string, store *InMemoryStore) {
	if len(args) > 1 {
		sendReply(conn, args[1])
	} else {
		sendReply(conn, "⚠️ ECHO requires a message")
	}
}
func runSet(conn net.Conn, args []string, store *InMemoryStore) {
	if len(args) <= 2 {
		sendReply(conn, "⚠️ Invalid key or value. Both must be non-empty.")
		return
	}

	msg, err := store.Set(args[1], args[2])

	if err != nil {
		sendReply(conn, err.Error())
		return
	}

	sendReply(conn, "OK", msg)
}
func runGet(conn net.Conn, args []string, store *InMemoryStore) {
	if len(args) <= 1 {
		sendReply(conn, "⚠️ Invalid key. Must be non-empty.")
		return
	}

	value, err := store.Get(args[1])
	if err != nil {
		sendReply(conn, err.Error())
		return
	}

	sendReply(conn, value)

}
func runListKeys(conn net.Conn, args []string, store *InMemoryStore) {
	keys, err := store.ListKeys()
	if err != nil {
		sendReply(conn, err.Error())
		return
	}

	sendReply(conn, "✅ Keys in Store:", keys)
}

func runDelete(conn net.Conn, args []string, store *InMemoryStore) {
	if len(args) <= 1 {
		sendReply(conn, "⚠️ Invalid key. Must be non-empty.")
		return
	}
	msg, err := store.Delete(args[1])
	if err != nil {
		sendReply(conn, err.Error())
		return
	}
	sendReply(conn, msg)

}
