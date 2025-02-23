package main

import (
	"fmt"
	"log"
	"syscall"
)

func main() {
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

		dataToSend, err := respFromServer()
		if err != nil {
			log.Printf("error getting response from server: %v", err)
			continue
		}

		if err := syscall.Sendto(conn, []byte("HTTP/1.1 200 ok\r\n\r\n"), 0, connAddr); err != nil {
			log.Printf("error sending message to socket: %v", err)
		}

		if err := syscall.Sendto(conn, dataToSend, 0, connAddr); err != nil {
			log.Printf("error sending message to socket: %v", err)
		}

		if err := syscall.Close(conn); err != nil {
			log.Printf("error closing connection: %v\n", err)
			continue
		}
	}
}

func respFromServer() (response []byte, err error) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return nil, fmt.Errorf("error opening socket: %v", err)
	}

	to := &syscall.SockaddrInet4{
		Port: 9000,
		Addr: [4]byte{127, 0, 0, 1},
	}

	if err := syscall.Connect(fd, to); err != nil {
		return nil, fmt.Errorf("could not connect: %v", err)
	}

	getREQ := []byte("GET / HTTP/1.1\r\nHost: localhost:9000\r\n")
	if err := syscall.Sendto(fd, getREQ, 0, to); err != nil {
		return nil, fmt.Errorf("error sending request to server: %v", err)
	}

	response = make([]byte, 4096)

	n, _, err := syscall.Recvfrom(fd, response, 0)
	if err != nil {
		return nil, fmt.Errorf("error receiving response %v", err)
	}

	if n == 0 {
		return nil, nil
	}

	return response[:n], nil
}
