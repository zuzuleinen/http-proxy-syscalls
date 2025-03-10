package main

import (
	"bytes"
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
	proxyAddr := &syscall.SockaddrInet4{
		Port: portNumber,
		Addr: [4]byte{0, 0, 0, 0},
	}

	// Enable SO_REUSEADDR to allow rebinding to the same address immediately
	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		log.Fatalf("setsockopt SO_REUSEADDR failed: %v", err)
	}

	if err := syscall.Bind(fd, proxyAddr); err != nil {
		log.Fatalf("error binding socket to proxyAddr: %v", err)
	}

	// listen for incoming connections
	if err := syscall.Listen(fd, 10); err != nil {
		log.Fatalf("error on listen: %v", err)
	}

	log.Printf("proxy started listening on port: %d\n", portNumber)

	for {
		// accept a connection on the socket
		clientConn, connAddr, err := syscall.Accept(fd)
		if err != nil {
			log.Fatalf("error on accepting connection on the socket")
		}

		log.Printf("new connection from port %#v\n", connAddr.(*syscall.SockaddrInet4).Port)

		request := make([]byte, 4096)
		n, _, err := syscall.Recvfrom(clientConn, request, 0)
		if err != nil {
			log.Printf("error on recvfrom: %v", err)
			continue
		}
		if n == 0 {
			log.Println("n == 0 breaking")
			break
		}

		respFromUpstream, err := sendToUpstream(request[:n])
		if err != nil {
			syscall.Close(clientConn)
			log.Printf("error getting response from server: %v", err)
			continue
		}

		if err := syscall.Sendto(clientConn, respFromUpstream.Bytes(), 0, connAddr); err != nil {
			log.Printf("error sending message to socket: %v", err)
		}

		if err := syscall.Close(clientConn); err != nil {
			log.Printf("error closing connection: %v\n", err)
			continue
		}
	}
}

func sendToUpstream(data []byte) (resp *bytes.Buffer, err error) {
	upstreamSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return nil, fmt.Errorf("error opening socket: %v", err)
	}
	defer func() {
		closeErr := syscall.Close(upstreamSocket)
		if closeErr != nil {
			err = closeErr
		}
	}()

	upstreamAddr := &syscall.SockaddrInet4{
		Port: 9000,
		Addr: [4]byte{127, 0, 0, 1},
	}

	if err := syscall.Connect(upstreamSocket, upstreamAddr); err != nil {
		return nil, fmt.Errorf("could not connect: %v", err)
	}
	log.Println("connected to upstream")

	err = syscall.Sendto(upstreamSocket, data, 0, upstreamAddr)
	if err != nil {
		return nil, fmt.Errorf("error sending to upstream %v", err)
	}

	bufSize := 4096
	upstreamResponse := make([]byte, bufSize)
	var fullResp bytes.Buffer

	for {
		n, _, err := syscall.Recvfrom(upstreamSocket, upstreamResponse, 0)
		log.Println("received from upstream", n, err)
		if err != nil {
			log.Printf("error receving from upstream from: %v", err)
			continue
		}

		fullResp.Write(upstreamResponse[:n])

		if n < bufSize {
			break
		}
	}

	return &fullResp, nil
}
