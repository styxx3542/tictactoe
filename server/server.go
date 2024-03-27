package main

import (
	"log"
	"net"
	"sync"
	"os"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/styxx3542/tictactoe/gameService"
)

// Define the server struct
type server struct {
	pb.UnimplementedTicTacToeServiceServer
	mu         sync.Mutex
	boardState [3][3]string
	clients    []*Client
	clientMu   sync.Mutex
}

// Client represents a connected client
type Client struct {
	id       string
	stream   pb.TicTacToeService_PlayStreamServer
	waitCh   chan bool
	isActive bool
}

// flattenBoardState returns a flattened representation of the board state
func (s *server) flattenBoardState() []string {
	var flatBoardState []string
	for _, row := range s.boardState {
		for _, cell := range row {
			flatBoardState = append(flatBoardState, cell)
		}
	}
	return flatBoardState
}

// checkGameOver checks if the game is over and returns the winner or an empty string if it's a draw
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

// MakeMove processes a move request and updates the board state
func (s *server) MakeMove(req *pb.MoveRequest) (*pb.BoardState, error) {
	s.mu.Lock()

	row, column := req.GetRow(), req.GetColumn()
	if row < 0 || row >= 3 || column < 0 || column >= 3 || s.boardState[row][column] != "" {
		return &pb.BoardState{
			Message: "Invalid move",
		}, nil
	}

	log.Printf("Player %s made a move at row %d, column %d", req.GetPlayerId(), row, column)

	s.boardState[row][column] = req.GetPlayerId()
	s.mu.Unlock()
	return s.GetBoardState(),nil
}

func (s *server) GetBoardState() *pb.BoardState {
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

	return &pb.BoardState{
		Message:    message,
		Board: s.flattenBoardState(),
		GameOver:   gameOver,
	}
}

// PlayStream handles the bi-directional streaming RPC for the game
func (s *server) PlayStream(stream pb.TicTacToeService_PlayStreamServer) error {
	s.clientMu.Lock()

	// Check if there are already two clients connected
	log.Println("Client connected")
	if len(s.clients) >= 2 {
		return status.Errorf(codes.Unavailable, "Server is full. Please try again later.")
	}
	var id string
	if len(s.clients) == 1 {
		id = "O"
	} else {
		id = "X"
	}

	// Create a new client and add it to the slice
	client := &Client{
		id:       id, // You need to implement this function
		stream:   stream,
		waitCh:   make(chan bool),
		isActive: false, // The first client is inactive by default
	}
	s.clients = append(s.clients, client)
	if len(s.clients) == 2 {
		log.Println("Game started")
		for _, c := range s.clients {
			if err := c.stream.Send(&pb.PlayStreamResponse{ClientId: c.id}); err != nil {
				return err
			}
		}
		s.clients[0].isActive = true // Set the first client as active
		s.clients[0].waitCh <- true
	}
	s.clientMu.Unlock()
	return s.processClientMoves(client)
}

func (s* server) terminateGame() {
	s.clientMu.Lock()
	defer s.clientMu.Unlock()
	for _, c := range s.clients {
		c.stream.Send(&pb.PlayStreamResponse{BoardState: s.GetBoardState()})
	}
	s.clients = nil
	s.boardState = [3][3]string{}
	log.Println("Game finished")
}

func (s *server) processClientMoves(client *Client) error {
	for {
		// Check if the client is active
		if !client.isActive {
			<-client.waitCh
		}

		// Send the current board state to the client
		boardState := s.GetBoardState()
		if boardState.GetGameOver() {
			// Remove the clients from the slice
			s.terminateGame()
			return nil
		}

		if err := client.stream.Send(&pb.PlayStreamResponse{BoardState: boardState}); err != nil {
			return err
		}

		// Receive the move from the client
		req, err := client.stream.Recv()
		if err != nil {
			// Handle error or client disconnection
			return err
		}

		// Process the move
		_, err = s.MakeMove(req.GetMoveRequest())
		if err != nil {
			// Handle error
			return err
		}

		// Check if the game is over
			// Switch to the next active client
		s.clientMu.Lock()
		for i := range s.clients {
			if s.clients[i].id == client.id {
				s.clients[i].isActive = false
				j := (i + 1) % len(s.clients)
				s.clients[j].isActive = true
				s.clients[j].waitCh <- true
				break
			}
		}
		s.clientMu.Unlock()
	}
}

func main() {
	// Create a gRPC server instance
	grpcServer := grpc.NewServer()
	serverPort := os.Args[1]

	// Register the TicTacToeService server with the gRPC server
	s := &server{}
	pb.RegisterTicTacToeServiceServer(grpcServer, s)
	file, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer file.Close()

	// Set the log output to the custom log file
	log.SetOutput(file)
	// Listen for incoming connections
	lis, err := net.Listen("tcp", ":"+serverPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Start the gRPC server
	log.Println("Starting gRPC server on port", serverPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
