package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter command (PUBLISH/CONSUME): ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)

	// Write the message to the server
	fmt.Fprintf(conn, text+"\n")

	// Close the write side of the connection after sending the message
	conn.(*net.TCPConn).CloseWrite()

	// Listen for the server's response
	response, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Print("Response from server: " + response)
}