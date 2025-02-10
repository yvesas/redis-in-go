package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// 📌 Função para interpretar o protocolo RESP (simplificado para comandos básicos) feito com ajuda de IA.
// func parseRESP(firstLine string, reader *bufio.Reader) ([]string, error) {
// 	if strings.HasPrefix(firstLine, "*") {
// 		numArgs := 0
// 		if _, err := fmt.Sscanf(firstLine, "*%d", &numArgs); err != nil || numArgs <= 0 {
// 			return nil, fmt.Errorf("-ERR invalid RESP format: malformed array")
// 		}

// 		args := make([]string, 0, numArgs)
// 		for i := 0; i < numArgs; i++ {
// 			sizeLine, err := reader.ReadString('\n')
// 			if err != nil {
// 				return nil, err
// 			}

// 			var argSize int
// 			if _, err := fmt.Sscanf(sizeLine, "$%d", &argSize); err != nil || argSize < 0 {
// 				return nil, fmt.Errorf("-ERR invalid RESP format: malformed bulk string")
// 			}

// 			arg := make([]byte, argSize)
// 			_, err = reader.Read(arg)
// 			if err != nil {
// 				return nil, err
// 			}

// 			args = append(args, string(arg))

// 			// Lê o '\r\n' final
// 			if _, err := reader.ReadString('\n'); err != nil {
// 				return nil, fmt.Errorf("-ERR invalid RESP format: missing CRLF")
// 			}
// 		}
// 		return args, nil
// 	}

// 	return nil, fmt.Errorf("-ERR invalid RESP format")
// }

func parseRESP(firstLine string, reader *bufio.Reader) ([]string, error) {
	// Verifica se é um array RESP
	if !strings.HasPrefix(firstLine, "*") {
		return nil, fmt.Errorf("-ERR invalid RESP format")
	}

	// Extrai o número de argumentos
	numArgs, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(firstLine, "*")))
	if err != nil {
		return nil, fmt.Errorf("-ERR invalid RESP format: malformed array")
	}

	// Se o array estiver vazio, retorna uma lista vazia sem erro
	if numArgs == 0 {
		return []string{}, nil
	}

	args := make([]string, 0, numArgs)

	for i := 0; i < numArgs; i++ {
		// Lê a linha com o tamanho do argumento
		sizeLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("-ERR invalid RESP format: missing size line")
		}
		sizeLine = strings.TrimSpace(sizeLine)

		// Verifica se começa com "$"
		if !strings.HasPrefix(sizeLine, "$") {
			return nil, fmt.Errorf("-ERR invalid RESP format: malformed bulk string")
		}

		// Converte para inteiro
		argSize, err := strconv.Atoi(strings.TrimPrefix(sizeLine, "$"))
		if err != nil || argSize < 0 {
			return nil, fmt.Errorf("-ERR invalid RESP format: malformed bulk string")
		}

		// Lê o argumento com o tamanho especificado
		arg := make([]byte, argSize)
		_, err = reader.Read(arg)
		if err != nil {
			return nil, fmt.Errorf("-ERR invalid RESP format: failed to read bulk string")
		}

		// Lê o '\r\n' final para validar a estrutura RESP
		crlf, err := reader.ReadString('\n')
		if err != nil || crlf != "\r\n" {
			return nil, fmt.Errorf("-ERR invalid RESP format: missing CRLF")
		}

		args = append(args, string(arg))
	}

	return args, nil
}

func formatArray(arr []string) string {
	if len(arr) == 0 {
		return "*0\r\n" // Array vazio no formato Redis
	}

	var response strings.Builder
	response.WriteString(fmt.Sprintf("*%d\r\n", len(arr))) // Número de itens no array
	for _, item := range arr {
		response.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(item), item)) // Cada item é um Bulk String
	}

	return response.String()
}
