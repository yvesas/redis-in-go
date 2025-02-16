package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type AppContext struct {
	Store  *InMemoryStore
	Config *ConfigStore
}

func main() {
	host := "0.0.0.0"
	port := 6379
	address := fmt.Sprintf("%s:%d", host, port)

	dir := flag.String("dir", "./data", "Directory to save files")
	dbfilename := flag.String("dbfilename", "dump.rdb", "Database file name")
	flag.Parse()

	config := NewConfigStore()
	success, errorConfig := config.Init(*dir, *dbfilename)
	if condition := success; condition {
		fmt.Printf("✅ Initialized config [dir]: %s - [dbfilename]: %s\n", *dir, *dbfilename)
	}
	if errorConfig != nil {
		fmt.Printf("❌ Failed to init config [dir]: %s - [dbfilename]: %s: %v\n", *dir, *dbfilename, errorConfig)
		os.Exit(1)
	}

	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("❌ Failed to bind to address %s: %v\n", address, err)
		os.Exit(1)
	}
	defer l.Close()
	fmt.Printf("✅ Server listening on address %s\n", address)

	store := NewInMemoryStore()

	appCtx := &AppContext{
		Store:  store,
		Config: config,
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("⚠️ Error accepting connection: %v\n", err)
			continue
		}

		go handleConnection(conn, appCtx)
	}
}

func handleConnection(conn net.Conn, appCtx *AppContext) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	commands := map[string]func(net.Conn, []string, *AppContext){
		"PING":   func(c net.Conn, args []string, ctx *AppContext) { sendReply(c, "PONG") },
		"ECHO":   runEcho,
		"SET":    runSet,
		"GET":    runGet,
		"DEL":    runDelete,
		"LIST":   runListKeys,
		"CONFIG": runConfig,
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
			handler(conn, args, appCtx)
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

func runEcho(conn net.Conn, args []string, ctx *AppContext) {
	if len(args) > 1 {
		sendReply(conn, args[1])
	} else {
		sendReply(conn, "-ERR ECHO requires a message")
	}
}

func runSet(conn net.Conn, args []string, ctx *AppContext) {
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

	_, err := ctx.Store.Set(key, value, ttlMs)

	if err != nil {
		sendReply(conn, err.Error())
		return
	}
	sendReply(conn, "OK")
}

func runGet(conn net.Conn, args []string, ctx *AppContext) {
	if len(args) <= 1 {
		sendReply(conn, "-ERR Invalid key")
		return
	}

	value, err := ctx.Store.Get(args[1])
	if err != nil {
		sendReply(conn, value)
		return
	}

	sendReply(conn, value)
}

func runDelete(conn net.Conn, args []string, ctx *AppContext) {
	if len(args) <= 1 {
		sendReply(conn, "-ERR Invalid key")
		return
	}

	count, err := ctx.Store.Delete(args[1:])
	if err != nil {
		sendReply(conn, err.Error())
		return
	}
	sendReply(conn, count)
}

func runListKeys(conn net.Conn, args []string, ctx *AppContext) {
	keys, err := ctx.Store.ListKeys()
	if err != nil {
		sendReply(conn, "$-1") // Empty array in RESP
		return
	}
	fmt.Println("✅ Keys in Store:", keys)
	sendReply(conn, keys)
}
func runConfig(conn net.Conn, args []string, ctx *AppContext) {
	if len(args) < 2 {
		sendReply(conn, "-ERR Invalid Config command")
		return
	}

	commands := map[string]func(net.Conn, []string, *AppContext){
		"SET": setConfig,
		"GET": getConfig,
		// "DEL":
		"LIST": ListConfig,
	}

	command := strings.ToUpper(args[1])
	data := args[2:]

	if handler, exists := commands[command]; exists {
		handler(conn, data, ctx)
	} else {
		fmt.Println("❌ Unknown command:", data)
		sendReply(conn, "-ERR Unknown command")
	}
}

func getConfig(conn net.Conn, args []string, ctx *AppContext) {
	if len(args) < 1 {
		sendReply(conn, "-ERR Invalid GET command. Usage: CONFIG GET key")
		return
	}
	value, err := ctx.Config.Get(args[0])
	if err != nil {
		sendReply(conn, err.Error())
		return
	}
	sendReply(conn, value)
}

func setConfig(conn net.Conn, args []string, ctx *AppContext) {
	if len(args) < 2 {
		sendReply(conn, "-ERR Invalid SET command. Usage: CONFIG SET key value")
		return
	}
	key, value := args[0], args[1]
	_, err := ctx.Config.Set(key, value)
	if err != nil {
		sendReply(conn, err.Error())
		return
	}
	sendReply(conn, "OK")
}

func ListConfig(conn net.Conn, args []string, ctx *AppContext) {
	keys, err := ctx.Config.ListConfig()
	if err != nil {
		sendReply(conn, "$-1") // Empty array in RESP
		return
	}
	fmt.Println("✅ Keys configured:", keys)
	sendReply(conn, keys)
}
