package server

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

const (
	maxMessageSize = 10240 // 10KB
	maxMessages    = 100   // Max number of messages the server can store
)

type Server struct {
	mu        sync.Mutex
	messages  []string      // Slice to store messages
	consumeCh chan net.Conn // Channel to queue consumers waiting for messages
}

func NewServer() *Server {
	return &Server{
		messages:  make([]string, 0, maxMessages),
		consumeCh: make(chan net.Conn, 100), // Channel to queue consumers waiting for messages
	}
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

	if len(s.messages) >= maxMessages {
		s.sendError(conn, "ERROR: occupied")
		return
	}

	// Append the message to the queue
	s.messages = append(s.messages, msg)
	s.sendResponse(conn, "SUCCESS: Message published")

	// Notify any waiting consumer that a new message is available
	select {
	case consumerConn := <-s.consumeCh:
		s.sendNextMessage(consumerConn)
	default:
		// No consumer waiting, nothing to do
	}
}

func (s *Server) handleConsume(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if there's a message to consume
	if len(s.messages) > 0 {
		s.sendNextMessage(conn)
	} else {
		// No message, add the connection to the waiting queue
		go func() {
			s.consumeCh <- conn
		}()
	}
}

func (s *Server) sendNextMessage(conn net.Conn) {
	if len(s.messages) == 0 {
		s.sendError(conn, "ERROR: no message")
		return
	}

	// Pop the first message from the queue and send it
	msg := s.messages[0]
	s.messages = s.messages[1:]
	response := fmt.Sprintf("SUCCESS: %s", msg)
	s.sendResponse(conn, response)
}

func (s *Server) sendResponse(conn net.Conn, msg string) {
	_, err := conn.Write([]byte(msg + "\n")) // Send newline-terminated message
	if err != nil {
		fmt.Println("Error sending response:", err)
		conn.Close() // Ensure connection is closed in case of write error
		return
	}

	go func() {
		conn.Close() // Close the connection only after ensuring write is successful
	}()
}

func (s *Server) sendError(conn net.Conn, errMsg string) {
	_, err := conn.Write([]byte(errMsg + "\n")) // Send newline-terminated error message
	if err != nil {
		fmt.Println("Error sending error message:", err)
		conn.Close() // Ensure connection is closed in case of error
		return
	}

	go func() {
		conn.Close() // Ensure connection is closed after sending error
	}()
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