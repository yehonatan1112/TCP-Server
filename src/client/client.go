package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// Connect to the TCP server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Create a reader for user input
	reader := bufio.NewReader(os.Stdin)

	for {
		// Read user input
		fmt.Print("Enter message (PUBLISH <msg> or CONSUME): ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if strings.HasPrefix(text, "PUBLISH ") {
			// Send PUBLISH message to the server
			_, err := conn.Write([]byte(text + "\n"))
			if err != nil {
				fmt.Println("Error writing to server:", err)
				break
			}

			// Close the writing end after sending the PUBLISH message
			err = conn.(*net.TCPConn).CloseWrite()
			if err != nil {
				fmt.Println("Error closing the write connection:", err)
				break
			}
		} else if text == "CONSUME" {
			// Send CONSUME message to the server
			_, err := conn.Write([]byte(text + "\n"))
			if err != nil {
				fmt.Println("Error writing to server:", err)
				break
			}
		} else {
			fmt.Println("Invalid input. Please enter PUBLISH <msg> or CONSUME.")
			continue
		}

		// Read response from server
		serverResponse, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from server:", err)
			break
		}

		fmt.Print("Server response: " + serverResponse)
	}
}