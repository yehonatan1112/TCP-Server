package main

import (
	"TCP-Server/src/client"
	"TCP-Server/src/server"
	"flag"
	"fmt"
)

//go run main.go -mode=server -port=8080
//go run main.go -mode=client -port=8080 -address=localhost

func main() {
	mode := flag.String("mode", "server", "Mode to run: server or client")
	port := flag.String("port", "8080", "Port to listen on or connect to")
	address := flag.String("address", "localhost", "Server address to connect to (client mode)")
	flag.Parse()

	if *mode == "server" {
		fmt.Println("Starting server...")
		srv := server.NewServer()
		srv.Start(*port)
	} else if *mode == "client" {
		fmt.Printf("Starting client to connect to %s:%s...\n", *address, *port)
		client.StartClient(fmt.Sprintf("%s:%s", *address, *port))
	} else {
		fmt.Println("Invalid mode. Use -mode=server or -mode=client")
	}
}
