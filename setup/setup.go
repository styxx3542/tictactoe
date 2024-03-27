package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	directoryPort = "8888"
)
func main(){
	// Request the directory service for the game service address
	serverAddr, err := net.ResolveUDPAddr("udp", ":" + directoryPort)
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}
	conn, err := net.DialUDP("udp",nil,serverAddr)
	if err != nil {
		fmt.Println("Error connecting to the directory service:", err)
		return
	}
	
	_, err = conn.Write([]byte("TicTacToeDiscovery"))
	if err != nil {
		fmt.Println("Error sending data to the directory service:", err)
		return
	}
	conn.Close()
	currentPort := strings.Split(conn.LocalAddr().String(), ":")[1]
	portNum, err := strconv.Atoi(currentPort)
	if err != nil {
		fmt.Println("Error converting port number to integer:", err)
		return
	}
	conn,err = net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port:portNum})
	if err != nil {
		fmt.Println("Error creating UDP socket:", err)
		return
	}
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading data from the directory service:", err)
		return
	}
	if n == 0 {
		fmt.Println("No game service found")
		return
	}
	gameServiceAddr := string(buffer[:n])
	if gameServiceAddr == currentPort {
		fmt.Println("Launching game server on port", currentPort)
		go exec.Command("go", "run", "../server/server.go", currentPort).Run()
	}
	cmd :=  exec.Command("go", "run", "../client/client.go",gameServiceAddr)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}

