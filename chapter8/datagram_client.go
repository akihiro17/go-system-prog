package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func main() {
	clientPath := filepath.Join(os.TempDir(), "unixdomainsocket-client")
	os.Remove(clientPath)
	conn, err := net.ListenPacket("unixgram", clientPath)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	path := filepath.Join(os.TempDir(), "unixdomainsocket-server")
	unixServerAddr, err := net.ResolveUnixAddr("unixgram", path)
	if err != nil {
		panic(err)
	}
	var serverAddr net.Addr = unixServerAddr
	fmt.Println("Sending to server")
	_, err = conn.WriteTo([]byte("Hello from Client"), serverAddr)
	if err != nil {
		panic(err)
	}
	fmt.Println("Receiveing from server")
	buffer := make([]byte, 1500)
	length, _, err := conn.ReadFrom(buffer)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Received %s\n", string(buffer[:length]))
}
