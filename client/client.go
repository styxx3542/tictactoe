package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	pb "github.com/styxx3542/tictactoe/gameService"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getMoveRequest(reader *bufio.Reader, id string, state *pb.BoardState) *pb.MoveRequest {
	// get move from terminal
	for {
		fmt.Print("\nEnter row and column (separated by space): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		coordinates := strings.Split(input, " ")

		if len(coordinates) != 2 {
			fmt.Println("Invalid input. Please enter row and column separated by space.")
			continue
		}

		row, err1 := strconv.Atoi(coordinates[0])
		column, err2 := strconv.Atoi(coordinates[1])

		if err1 != nil || err2 != nil {
			fmt.Println("Invalid input. Please enter row and column as integers.")
			continue
		}
		if !validateMove(row, column,state) {
			fmt.Println("Invalid move. Please try again.")
			continue
		}

		return &pb.MoveRequest{PlayerId: id, Row: int32(row), Column: int32(column)}
	}
}

func validateMove(row int, column int, boardState *pb.BoardState) bool {
	if row < 0 || row > 2 || column < 0 || column > 2 {
		return false
	}

	if boardState.Board[row*3+column] != "" {
		return false
	}

	return true
}

func displayBoardState(response *pb.BoardState) {
	if response.GetMessage() == "Invalid move" {
		fmt.Println("Invalid move. Please try again.")
		return
	}

	for row := 0; row < 3; row++ {
		for column := 0; column < 3; column++ {
			fmt.Printf(" %s ", response.Board[row*3+column])
			if column < 2 {
				fmt.Print("|")
			}
		}
		if row < 2 {
			fmt.Println("\n---------")
		}
	}
}

func main() {
	// Create a gRPC client connection
	serverAddr := "localhost:" + os.Args[1] 
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()

	// Create a gRPC client instance
	client := pb.NewTicTacToeServiceClient(conn)

	// Start the PlayStream RPC
	stream, err := client.PlayStream(context.Background())
	if err != nil {
		log.Fatalf("could not start game: %v", err)
	}
	fmt.Println("Client stream started")

	// Receive the client's ID from the server
	clientStream, err := stream.Recv()
	if err != nil {
		log.Fatalf("could not receive client ID: %v", err)
	}
	id := clientStream.GetClientId()
	fmt.Printf("Your client ID is: %s\n", id)

	reader := bufio.NewReader(os.Stdin)

	// Game loop
	for {
		// Get the state from the server
		resp, err := stream.Recv()
		if err == io.EOF {
			// Server has closed the stream
			break
		}
		if err != nil {
			log.Fatalf("could not receive response: %v", err)
		}

		// Display the state
		displayBoardState(resp.GetBoardState())

		// Check if the game is over
		if resp.GetBoardState().GetGameOver() {
			log.Println(resp.GetBoardState().GetMessage())
			break
		}

		moveReq := getMoveRequest(reader, id, resp.GetBoardState())
		if err := stream.Send(&pb.PlayStreamRequest{MoveRequest: moveReq}); err != nil {
			log.Fatalf("could not send move: %v", err)
		}

		// Receive the response from the server
	}
}
