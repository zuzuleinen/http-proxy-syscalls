package main

import (
	"log"
	"syscall"
)

func main() {
	// socket that listens to HTTP GET request on :8000
	// once the GET request is received, we open a socket to connect with the server
	// we do a GET request to the server
	// we write the response from the server to the initial socket

	// create the TCP socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		log.Fatalf("error opening socket: %v", err)
	}

	// bind the socket to localhost:8000
	portNumber := 8000
	addr := &syscall.SockaddrInet4{
		Port: portNumber,
		Addr: [4]byte{0, 0, 0, 0},
	}

	// Enable SO_REUSEADDR to allow rebinding to the same address immediately
	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		log.Fatalf("setsockopt SO_REUSEADDR failed: %v", err)
	}

	if err := syscall.Bind(fd, addr); err != nil {
		log.Fatalf("error binding socket to addr: %v", err)
	}

	// listen for incoming connections
	if err := syscall.Listen(fd, 10); err != nil {
		log.Fatalf("error on listen: %v", err)
	}

	log.Printf("proxy started listening on port: %d\n", portNumber)

	for {
		// accept a connection on the socket
		conn, connAddr, err := syscall.Accept(fd)
		if err != nil {
			log.Fatalf("error on accepting connection on the socket")
		}

		log.Printf("new connection from port %#v\n", connAddr.(*syscall.SockaddrInet4).Port)

		request := make([]byte, 4096)
		n, _, err := syscall.Recvfrom(conn, request, 0)
		if err != nil {
			log.Printf("error on recvfrom: %v", err)
			continue
		}
		if n == 0 {
			log.Println("n == 0 breaking")
			break
		}

		if err := syscall.Sendto(conn, []byte("HTTP/1.1 200 ok\r\n\r\n"), 0, connAddr); err != nil {
			log.Printf("error sending message to socket: %v", err)
		}

		dataToSend := []byte("hello world!\n") // todo fetch this from server
		if err := syscall.Sendto(conn, dataToSend, 0, connAddr); err != nil {
			log.Printf("error sending message to socket: %v", err)
		}

		if err := syscall.Close(conn); err != nil {
			log.Printf("error closing connection: %v\n", err)
			continue
		}
	}
}
