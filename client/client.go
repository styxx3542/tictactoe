package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/styxx3542/tictactoe/gameService"
)

func getMoveFromTerminal() pb.MoveRequest {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter row and column (separated by space): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		coordinates := strings.Split(input, " ")
		if len(coordinates) != 2 {
			fmt.Println("Invalid input. Please enter row and column separated by space.")
			continue
		}
		row, err1 := strconv.Atoi(coordinates[0])
		column, err2 := strconv.Atoi(coordinates[1])
		if err1 != nil || err2 != nil || row < 0 || row > 2 || column < 0 || column > 2 {
			fmt.Println("Invalid input. Please enter valid row and column numbers (0, 1, or 2).")
			continue
		}
		fmt.Print("Enter player id (X or O): ")
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)
		return pb.MoveRequest{
			PlayerId: input,
			Row:      int32(row),
			Column:   int32(column),
		}
	}
}

func getMoveRequest() *pb.MoveRequest {
	// get move from terminal
	moveRequest := getMoveFromTerminal()
	return &moveRequest
}

func displayMoveResponse(response *pb.MoveResponse, err error) {
	if err != nil {
		log.Fatalf("could not make move: %v", err)
	}
	for row := 0; row < 3; row++ {
		for column := 0; column < 3; column++ {
			fmt.Printf("%s ", response.BoardState[row*3+column])
		}
		fmt.Println()
	}
}

func main() {
	// Create a gRPC client connection
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()

	// Create a gRPC client instance
	client := pb.NewTicTacToeServiceClient(conn)

	// Example: Make a move using the gRPC client
	for {
		response, err := client.MakeMove(context.Background(), getMoveRequest())
		displayMoveResponse(response, err)
		if response.GetMessage() == "Game over: Draw" || response.GetMessage() == "Game over: Player X wins" || response.GetMessage() == "Game over: Player O wins" {
			log.Println(response.GetMessage())
			break
		}
	}
}
