package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "gochat/proto"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewChatServiceClient(conn)

	// Start streaming with the server
	stream, err := client.ChatStream(context.Background())
	if err != nil {
		log.Fatalf("Error creating stream: %v", err)
	}

	// Get username
	fmt.Print("Enter your username: ")
	reader := bufio.NewReader(os.Stdin)
	username, _ := reader.ReadString('\n')
	username = username[:len(username)-1] // Remove newline

	// Send the first message with the username
	stream.Send(&pb.ChatMessage{Username: username, Message: "joined the chat!"})

	// Handle incoming messages in a separate goroutine
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				log.Fatalf("Error receiving message: %v", err)
			}
			fmt.Printf("[%s] %s: %s\n", time.Unix(msg.Timestamp, 0).Format("15:04"), msg.Username, msg.Message)
		}
	}()

	// Send messages from the user
	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		text = text[:len(text)-1] // Remove newline
		stream.Send(&pb.ChatMessage{Username: username, Message: text})
	}
}
