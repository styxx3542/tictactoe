package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

func main() {
	// Create a UDP socket for listening
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: discoveryPort})
	if err != nil {
		fmt.Println("Error creating UDP socket:", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Listening for discovery messages...")

	// Map to store discovered peers
	discoveredPeers := make(map[string]time.Time)
	var mutex sync.Mutex

	// Goroutine for handling incoming messages
	go func() {
		buffer := make([]byte, bufferSize)
		for {
			n, addr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println("Error reading UDP message:", err)
				continue
			}

			message := string(buffer[:n])
			if message == "TicTacToeDiscovery" {
				// Add peer to discoveredPeers map
				peerAddr := addr.IP.String()
				mutex.Lock()
				discoveredPeers[peerAddr] = time.Now()
				mutex.Unlock()
				fmt.Println("Discovered peer:", peerAddr)
			}
		}
	}()

	//Goroutine for removing inactive participants
	go func() {
		mutex.Lock()
		for node, lastSeen := range discoveredPeers{
			if time.Since(lastSeen) > 30 * time.Second {
				fmt.Println("Peer disconnected:", node)
				delete(discoveredPeers,node)
			} 
		}
		mutex.Unlock()
		time.Sleep(30 * time.Second)
	}()

	// Wait indefinitely
	select {}
}
