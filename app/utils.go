package main

import (
	"bufio"
	"fmt"
	"strings"
)

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
