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
	defer l.Close() // ??? Listener date when exiting the function
	fmt.Printf("✅ Server listening on address %s\n", address)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("⚠️ Error accepting connection: %v\n", err) // err.Error()
			continue                                               // Keep trying to accept new connections.
		}

		go handleConnection(conn)
		// _, writeErr := conn.Write([]byte("+PONG\r\n"))
		// if writeErr != nil {
		// 	fmt.Printf("⚠️ Failed to answered (write to connection): %v\n", writeErr)
		// } else {
		// 	fmt.Println("✅ Answered with success.")
		// }
		// conn.Close()
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("⚠️ Client disconnected.")
			return
		}

		args, err := parseRESP(line, reader)
		if err != nil {
			fmt.Println("⚠️ Error parsing request:", err)
			conn.Write([]byte("-ERR invalid request\r\n"))
			continue
		}

		if len(args) > 0 && strings.ToUpper(args[0]) == "PING" {
			sendMessage(conn, "PONG")
		} else {
			conn.Write([]byte("-ERR unknown command\r\n"))
			fmt.Println("⚠️ Unknown command:", args)
		}
	}
}

func sendMessage(conn net.Conn, message string) {
	response := fmt.Sprintf("+%s\r\n", message)
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("⚠️ Error sending response:", err)
		return
	}
	fmt.Println("✅ Answered with success. Sent: ", message)
}

// 📌 Função para interpretar o protocolo RESP (simplificado para comandos básicos) feito com ajuda de IA.
func parseRESP(firstLine string, reader *bufio.Reader) ([]string, error) {
	// Se o primeiro caractere for "*", significa que é um array
	if strings.HasPrefix(firstLine, "*") {
		numArgs := 0
		fmt.Sscanf(firstLine, "*%d", &numArgs)

		args := make([]string, 0, numArgs)
		for i := 0; i < numArgs; i++ {
			// Lê a linha que contém o tamanho do argumento
			sizeLine, err := reader.ReadString('\n')
			if err != nil {
				return nil, err
			}

			// Pega apenas o valor após o "$", que é o tamanho
			var argSize int
			fmt.Sscanf(sizeLine, "$%d", &argSize)

			// Lê a linha do argumento com o tamanho certo
			arg := make([]byte, argSize)
			_, err = reader.Read(arg)
			if err != nil {
				return nil, err
			}

			// Adiciona o argumento ao array
			args = append(args, string(arg))

			// Lê o '\r\n' final
			reader.ReadString('\n')
		}
		return args, nil
	}

	// Caso contrário, retorna erro
	return nil, fmt.Errorf("invalid RESP format")
}
