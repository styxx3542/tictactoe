package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	directoryPort = 8888
)
func main(){
	// Create a UDP socket for listening
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: directoryPort})
	if err != nil {
		fmt.Println("Error creating UDP socket:", err)
		os.Exit(1)
	}
	defer conn.Close()
	waitingPeers := []int{}
	fmt.Println("Listening for discovery messages...")
	//goroutine for handling incoming messages
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, addr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println("Error reading UDP message:", err)
				continue
			}
			message := string(buffer[:n])
			if message == "TicTacToeDiscovery" {
				waitingPeers = append(waitingPeers, addr.Port)
				fmt.Println("Discovered peer:", addr.IP.String())
			}
		}
	}()

	//goroutine for pairing peers 
	go func() {
		for {
			if len(waitingPeers) >= 2 {
				fmt.Println("Pairing peers:", waitingPeers[0], waitingPeers[1])
				handleElection(waitingPeers[0], waitingPeers[1])
				waitingPeers = waitingPeers[2:]
			}
			time.Sleep(5 * time.Second)
		}
	}()
	select {}
}

	func handleElection(peer1 int, peer2 int) {
		fmt.Println("Election started between:", peer1, peer2)
		leaderIdx := rand.Intn(2)
		var leader int
		if leaderIdx == 0 {
			leader = peer1
		} else {
			leader = peer2
		}
		fmt.Println("Leader elected:", leader)
		for _, peer := range []int{peer1, peer2} {
			conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
				Port: peer,
			})
			if err != nil {
				fmt.Println("Error creating UDP socket:", err)
				os.Exit(1)
			}
			defer conn.Close()
			_, err = conn.Write([]byte(strconv.Itoa(leader)))
			if err != nil {
				fmt.Println("Error broadcasting leader:", err)
			}
		}
}
