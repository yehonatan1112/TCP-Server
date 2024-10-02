package server

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

const (
	maxMessageSize = 10240 // 10KB
	maxMessages    = 100   // Maximum number of stored messages
)

type Server struct {
	mu       sync.Mutex
	messages []string          // Queue to store messages
	cond     *sync.Cond        // Condition variable to wait for messages
	waiting  map[net.Conn]bool // Track clients waiting for messages
}

func NewServer() *Server {
	s := &Server{
		messages: []string{},
		waiting:  make(map[net.Conn]bool),
	}
	s.cond = sync.NewCond(&s.mu) // Initialize the condition variable
	return s
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, maxMessageSize)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err)
			return
		}

		msg := strings.TrimSpace(string(buffer[:n]))
		fmt.Printf("Received message: %s\n", msg)

		if strings.HasPrefix(msg, "PUBLISH ") {
			s.handlePublish(conn, msg[8:]) // Extract the message after "PUBLISH "
		} else if strings.HasPrefix(msg, "CONSUME") {
			s.handleConsume(conn)
		} else {
			s.sendResponse(conn, "ERROR: Invalid message format")
		}
	}
}

func (s *Server) handlePublish(conn net.Conn, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(msg) > maxMessageSize {
		s.sendResponse(conn, "ERROR: Message too large")
		return
	}

	if len(s.messages) >= maxMessages {
		s.sendResponse(conn, "ERROR: Message queue full")
		return
	}

	// Add message to queue
	s.messages = append(s.messages, msg)
	fmt.Printf("Message published: %s\n", msg)

	// Notify any waiting clients
	if len(s.waiting) > 0 {
		for conn := range s.waiting {
			if len(s.messages) > 0 {
				s.sendResponse(conn, "SUCCESS: "+s.messages[0])
				s.messages = s.messages[1:] // Remove the consumed message from the queue
				delete(s.waiting, conn)     // Stop waiting for this client
				conn.Close()                // Close the connection after sending
				break
			}
		}
	} else {
		s.sendResponse(conn, "SUCCESS: Message published")
	}
}

func (s *Server) handleConsume(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.messages) > 0 {
		// Send the first message in the queue
		s.sendResponse(conn, "SUCCESS: "+s.messages[0])
		s.messages = s.messages[1:] // Remove the message from the queue
		conn.Close()                // Close connection after sending
	} else {
		// No message available, wait for a message to be published
		fmt.Println("No messages available, client waiting...")
		s.waiting[conn] = true
	}
}

func (s *Server) sendResponse(conn net.Conn, msg string) {
	_, err := conn.Write([]byte(msg + "\n"))
	if err != nil {
		fmt.Println("Error sending response:", err)
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
