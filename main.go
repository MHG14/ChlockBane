package main

import (
	"context"
	"log"
	"time"

	"github.com/mhg14/ChlockBane/node"
	"github.com/mhg14/ChlockBane/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	makeNode(":3000", []string{})
	time.Sleep(time.Second)
	makeNode(":4000", []string{":3000"})
	time.Sleep(4 * time.Second)
	makeNode(":5000", []string{":4000"})

	select {}
}

func makeNode(listenAddr string, bootstrapNodes []string) *node.Node {
	n := node.NewNode()
	go n.Start(listenAddr, bootstrapNodes)
	return n
}

func makeTransaction() {
	client, err := grpc.Dial(":3000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	c := proto.NewNodeClient(client)

	version := &proto.Version{
		Version:    "ChlockBane-0.1",
		Height:     1,
		ListenAddr: ":3000",
	}

	_, err = c.Handshake(context.TODO(), version, grpc.EmptyCallOption{})
	if err != nil {
		log.Fatal(err)
	}
}
