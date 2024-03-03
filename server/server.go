package main

import (
	"context"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"

	pb "github.com/styxx3542/tictactoe/gameService"
)

// Define the server struct
type server struct {
	pb.UnimplementedTicTacToeServiceServer;
	mu         sync.Mutex
	boardState [3][3]string
}

func (s *server) flattenBoardState() []string {
	var flatBoardState []string
	for _, row := range s.boardState {
		for _, cell := range row {
			flatBoardState = append(flatBoardState, cell)
		}
	}
	return flatBoardState
}

func (s *server) checkGameOver() (bool, string) {
	// Check rows
	for _, row := range s.boardState {
		if row[0] != "" && row[0] == row[1] && row[1] == row[2] {
			return true, row[0]
		}
	}
	// Check columns
	for i := 0; i < 3; i++ {
		if s.boardState[0][i] != "" && s.boardState[0][i] == s.boardState[1][i] && s.boardState[1][i] == s.boardState[2][i] {
			return true, s.boardState[0][i]
		}
	}
	// Check diagonals
	if s.boardState[0][0] != "" && s.boardState[0][0] == s.boardState[1][1] && s.boardState[1][1] == s.boardState[2][2] {
		return true, s.boardState[0][0]
	}
	if s.boardState[0][2] != "" && s.boardState[0][2] == s.boardState[1][1] && s.boardState[1][1] == s.boardState[2][0] {
		return true, s.boardState[0][2]
	}
	// Check for Draw
	for _, row := range s.boardState {
		for _, cell := range row {
			if cell == "" {
				return false, ""
			}
		}
	}
	return true, ""
}

// Implement the MakeMove RPC method
func (s *server) MakeMove(ctx context.Context, req *pb.MoveRequest) (*pb.MoveResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	row, column := req.GetRow(), req.GetColumn()
	if row < 0 || row >= 3 || column < 0 || column >= 3 || s.boardState[row][column] != "" {
		return &pb.MoveResponse{
			Success: false,
			Message: "Invalid move",
		}, nil
	}
	s.boardState[row][column] = req.GetPlayerId()
	gameOver, winner := s.checkGameOver()
	var message string
	if gameOver {
		if winner == "" {
			message = "Game over: Draw"
		} else {
			message = "Game over: Player " + winner + " wins"
		}
	} else {
		message = "Move successfully made"
	}
	return &pb.MoveResponse{
		Success:    true,
		Message:    message,
		BoardState: s.flattenBoardState(),
	}, nil
}

func main() {
	// Create a gRPC server instance
	grpcServer := grpc.NewServer()

	// Register the TicTacToeService server with the gRPC server
	pb.RegisterTicTacToeServiceServer(grpcServer, &server{})

	// Listen for incoming connections
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Start the gRPC server
	log.Println("Starting gRPC server...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
