package main

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRESP_ValidInput(t *testing.T) {
	input := "*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	args, err := parseRESP("*2\r\n", reader)
	assert.Nil(t, err, "Expected no error for valid RESP input")
	assert.Equal(t, []string{"ECHO", "hey"}, args, "Expected parsed args to match input")
}

func TestParseRESP_InvalidFormat(t *testing.T) {
	input := "*2\r\n$4ECHO\r\n$3hey\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	args, err := parseRESP("*2\r\n", reader)

	assert.Error(t, err)
	assert.Nil(t, args)
	assert.Contains(t, err.Error(), "invalid RESP format")
}

func TestParseRESP_EmptyArray(t *testing.T) {
	input := "*0\r\n"
	reader := bufio.NewReader(strings.NewReader(input))
	args, err := parseRESP(input, reader)
	assert.Nil(t, err, "Expected no error for empty array")
	assert.Equal(t, []string{}, args, "Expected empty array")
}

func TestParseRESP_MissingCRLF(t *testing.T) {
	input := "*2\r\n$4\r\nECHO\r\n$3\r\nhey"
	reader := bufio.NewReader(strings.NewReader(input))

	_, err := parseRESP("*2\r\n", reader)
	assert.NotNil(t, err, "Expected error for missing CRLF")
	assert.Contains(t, err.Error(), "missing CRLF", "Expected missing CRLF error")
}

func TestFormatArray_ValidInput(t *testing.T) {
	arr := []string{"SET", "foo", "bar"}
	expected := "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"

	result := formatArray(arr)

	assert.Equal(t, expected, result)
}

func TestFormatArray_EmptyArray(t *testing.T) {
	expected := "*0\r\n"

	result := formatArray([]string{})

	assert.Equal(t, expected, result)
}
