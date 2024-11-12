package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	pb "gochat/proto"

	"google.golang.org/grpc"
)

type chatServer struct {
	pb.UnimplementedChatServiceServer
	clients map[string]pb.ChatService_ChatStreamServer
	mu      sync.Mutex
}

// NewChatServer initializes the chat server
func NewChatServer() *chatServer {
	return &chatServer{
		clients: make(map[string]pb.ChatService_ChatStreamServer),
	}
}

// ChatStream handles bidirectional streaming for chat messages
func (s *chatServer) ChatStream(stream pb.ChatService_ChatStreamServer) error {
	// Receive the first message from the client to get their username
	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	username := msg.Username
	fmt.Printf("User %s joined the chat\n", username)

	// Add the client to the server's client map
	s.mu.Lock()
	s.clients[username] = stream
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, username)
		s.mu.Unlock()
		fmt.Printf("User %s left the chat\n", username)
	}()

	// Listen for messages from this client and broadcast them
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		msg.Timestamp = time.Now().Unix()
		s.broadcastMessage(msg)
	}
}

// broadcastMessage sends a message to all connected clients
func (s *chatServer) broadcastMessage(msg *pb.ChatMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for username, clientStream := range s.clients {
		if username != msg.Username { // Skip the sender
			if err := clientStream.Send(msg); err != nil {
				log.Printf("Error sending message to %s: %v", username, err)
			}
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen on port 50051: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterChatServiceServer(grpcServer, NewChatServer())

	fmt.Println("Chat server is running on port 50051...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
