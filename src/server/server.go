package server

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

const (
	maxMessageSize = 10240 // 10KB
)

type Server struct {
	mu      sync.Mutex
	message string
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) handleConnection(conn net.Conn) {
	buffer := make([]byte, maxMessageSize)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err)
			return // Exit the loop if there's an error
		}

		msg := string(buffer[:n])
		fmt.Printf("Received message: %s\n", msg)

		if strings.HasPrefix(msg, "PUBLISH ") {
			s.handlePublish(conn, msg[8:]) // Extract the message after "PUBLISH "
		} else if strings.HasPrefix(msg, "CONSUME") {
			s.handleConsume(conn)
		} else {
			s.sendError(conn, "ERROR: Invalid message format")
		}
	}
}

func (s *Server) handlePublish(conn net.Conn, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(msg) > maxMessageSize {
		s.sendError(conn, "ERROR: Message too large")
		return
	}

	if s.message != "" {
		s.sendError(conn, "ERROR: occupied")
		return
	}

	s.message = msg
	s.sendResponse(conn, "SUCCESS: Message published")
}

func (s *Server) handleConsume(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.message == "" {
		s.sendError(conn, "ERROR: no message")
		return
	}

	response := fmt.Sprintf("SUCCESS: %s", s.message)
	s.sendResponse(conn, response)
	s.message = "" // Clear the message after consumption
}

func (s *Server) sendResponse(conn net.Conn, msg string) {
	_, err := conn.Write([]byte(msg + "\n")) // Send newline-terminated message
	if err != nil {
		fmt.Println("Error sending response:", err)
	}
}

func (s *Server) sendError(conn net.Conn, errMsg string) {
	_, err := conn.Write([]byte(errMsg + "\n")) // Send newline-terminated error message
	if err != nil {
		fmt.Println("Error sending error message:", err)
	}
}

func (s *Server) Start(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on port", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go s.handleConnection(conn)
	}
}
