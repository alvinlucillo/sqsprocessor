package main

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/alvinlucillo/sqs-processor/protogen/sqs"
)

type Client struct {
	pb.SQSServiceClient
}

var addr string = "0.0.0.0:50051"

func main() {
	var conn *grpc.ClientConn

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}

	// perform close at the end of the function
	defer conn.Close()

	c := pb.NewSQSServiceClient(conn)

	res, err := c.ReceiveMessage(context.Background(), &pb.SQSReceiveMessageRequest{VisibilityTimeout: 5, WaitTime: 5})
	if err != nil {
		log.Fatalf("Failed to call receive message: %v\n", err)
	}

	log.Printf("Message: %s\n", res.Messages)

	res1, err := c.DeleteMessage(context.Background(), &pb.SQSDeleteMessageRequest{MessageID: res.Messages[0].MessageID})
	if err != nil {
		log.Fatalf("Failed to call delete message: %v\n", err)
	}

	log.Printf("Result: %s\n", res1)

}
