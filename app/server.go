package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	host := "0.0.0.0"
	port := 6379
	address := fmt.Sprintf("%s:%d", host, port)

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
			fmt.Printf("⚠️ Error accepting connection: %v\n", err)
			continue
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
		"SET":  runSet,
		"GET":  runGet,
		"DEL":  runDelete,
		"LIST": runListKeys,
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
			sendReply(conn, err.Error())
			continue
		}

		fmt.Println("📥 Received args:", args)
		if len(args) == 0 {
			sendReply(conn, "-ERR Empty command")
			continue
		}

		command := strings.ToUpper(args[0])
		if handler, exists := commands[command]; exists {
			handler(conn, args, store)
		} else {
			fmt.Println("❌ Unknown command:", args)
			sendReply(conn, "-ERR Unknown command")
		}
	}
}

func sendReply(conn net.Conn, message interface{}) {
	var response string

	switch v := message.(type) {
	case string:
		if strings.HasPrefix(v, "-ERR") {
			response = fmt.Sprintf("%s\r\n", v)
		} else {
			response = fmt.Sprintf("+%s\r\n", v)
		}
	case error:
		response = fmt.Sprintf("-ERR %s\r\n", v.Error())
	case int:
		response = fmt.Sprintf(":%d\r\n", v)
	case []string:
		response = formatArray(v)
	default:
		response = "-ERR Unsupported data type\r\n"
	}

	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("❌ Error sending response:", err)
		fmt.Println("-ERR Error sending response:", err)
	}
	fmt.Println("✅ Sent:", response)
}

func runEcho(conn net.Conn, args []string, store *InMemoryStore) {
	if len(args) > 1 {
		sendReply(conn, args[1])
	} else {
		sendReply(conn, "-ERR ECHO requires a message")
	}
}

func runSet(conn net.Conn, args []string, store *InMemoryStore) {
	if len(args) < 3 {
		sendReply(conn, "-ERR Invalid SET command. Usage: SET key value [PX milliseconds]")
		return
	}

	key, value := args[1], args[2]
	var ttlMs int64 = 0

	if len(args) == 5 && strings.ToUpper(args[3]) == "PX" {
		parsedTTL, err := strconv.ParseInt(args[4], 10, 64)
		if err != nil || parsedTTL <= 0 {
			sendReply(conn, "-ERR PX argument must be a positive integer")
			return
		}
		ttlMs = parsedTTL
	}

	_, err := store.Set(key, value, ttlMs)

	if err != nil {
		sendReply(conn, err.Error())
		return
	}
	sendReply(conn, "OK")
}

func runGet(conn net.Conn, args []string, store *InMemoryStore) {
	if len(args) <= 1 {
		sendReply(conn, "-ERR Invalid key")
		return
	}

	value, err := store.Get(args[1])
	if err != nil {
		sendReply(conn, value)
		return
	}

	sendReply(conn, value)
}

func runDelete(conn net.Conn, args []string, store *InMemoryStore) {
	if len(args) <= 1 {
		sendReply(conn, "-ERR Invalid key")
		return
	}

	fmt.Println(">> ✅ Keys: ", args[1:])

	count, err := store.Delete(args[1:])
	if err != nil {
		sendReply(conn, err.Error())
		return
	}
	sendReply(conn, count)
}

func runListKeys(conn net.Conn, args []string, store *InMemoryStore) {
	keys, err := store.ListKeys()
	if err != nil {
		sendReply(conn, "$-1") // Empty array in RESP
		return
	}
	fmt.Println("✅ Keys in Store:", keys)
	sendReply(conn, keys)
}
