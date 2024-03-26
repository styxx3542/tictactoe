package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

const (
	discoveryPort = 8888
	discoveryMessage = "TicTacToeDiscovery"
	discoveryInterval = 5 * time.Second
	bufferSize    = 1024
)

func discover() {
	// Create a UDP socket for broadcasting
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(255, 255, 255, 255),
		Port: discoveryPort,
	})
	if err != nil {
		fmt.Println("Error creating UDP socket:", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Start broadcasting presence
	fmt.Println("Broadcasting presence...")
	for {
		_, err := conn.Write([]byte(discoveryMessage))
		if err != nil {
			fmt.Println("Error broadcasting presence:", err)
		}

		time.Sleep(discoveryInterval)
	}
}
