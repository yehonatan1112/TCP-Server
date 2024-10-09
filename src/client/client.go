package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	for {
		// Read input from the user (PUBLISH or CONSUME)
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter PUBLISH <message> or CONSUME: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input) // Remove trailing newline and spaces

		// Create a connection to the server
		conn, err := net.Dial("tcp", "localhost:8080")
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}

		// Send the message to the server
		_, err = conn.Write([]byte(input + "\n"))
		if err != nil {
			fmt.Println("Error writing to server:", err)
			conn.Close()
			continue
		}

		// Close the write side of the connection to indicate end of message
		err = conn.CloseWrite()
		if err != nil {
			fmt.Println("Error closing write connection:", err)
			conn.Close()
			continue
		}

		// Wait for the server's response
		response, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from server:", err)
		} else {
			fmt.Println("Server response:", strings.TrimSpace(response))
		}

		// Close the connection
		conn.Close()
	}
}