package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/mhg14/ChlockBane/node"
	"github.com/mhg14/ChlockBane/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	node := node.NewNode()
	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer(grpc.EmptyServerOption{})
	proto.RegisterNodeServer(grpcServer, node)
	fmt.Println("node running on port", ":3000")

	go func() {
		for {
			time.Sleep(2 * time.Second)
			makeTransaction()
		}
	}()

	grpcServer.Serve(ln)
}

func makeTransaction() {
	client, err := grpc.Dial(":3000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	c := proto.NewNodeClient(client)

	version := &proto.Version{
		Version: "ChlockBane-0.1",
		Height:  1,
	}

	_, err = c.Handshake(context.TODO(), version, grpc.EmptyCallOption{})
	if err != nil {
		log.Fatal(err)
	}
}
