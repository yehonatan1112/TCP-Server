package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func StartClient(serverAddress string) {
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter command (PUBLISH <message> or CONSUME): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.HasPrefix(input, "PUBLISH") || strings.HasPrefix(input, "CONSUME") {
			// Send input to server
			_, err := conn.Write([]byte(input + "\n")) // Add newline for server read consistency
			if err != nil {
				fmt.Println("Error sending data:", err)
				return
			}

			// Receive response from server
			response, err := bufio.NewReader(conn).ReadString('\n') // Read server response until newline
			if err != nil {
				fmt.Println("Error reading response:", err)
				return
			}

			fmt.Println("Server response:", strings.TrimSpace(response))
		} else {
			fmt.Println("Invalid command. Use PUBLISH <message> or CONSUME.")
		}
	}
}
