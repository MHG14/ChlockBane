package main

import (
	"context"

	"log"
	"time"

	"github.com/mhg14/ChlockBane/crypto"
	"github.com/mhg14/ChlockBane/node"
	"github.com/mhg14/ChlockBane/proto"
	"github.com/mhg14/ChlockBane/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	makeNode(":3000", []string{}, true)
	time.Sleep(time.Second)
	makeNode(":4000", []string{":3000"}, false)
	time.Sleep(4 * time.Second)
	makeNode(":5000", []string{":4000"}, false)

	for {
		time.Sleep(time.Second)
		makeTransaction()
	}
}

func makeNode(listenAddr string, bootstrapNodes []string, isValidator bool) *node.Node {
	cfg := node.ServerConfig{
		Version:    "ChlockBane-0.1",
		ListenAddr: listenAddr,
	}
	if isValidator {
		cfg.PrivateKey = crypto.GeneratePrivateKey()
	}
	n := node.NewNode(cfg)
	go n.Start(listenAddr, bootstrapNodes)
	return n
}

func makeTransaction() {
	client, err := grpc.Dial(":3000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	c := proto.NewNodeClient(client)

	privKey := crypto.GeneratePrivateKey()
	pubKey := privKey.Public()
	tx := &proto.Transaction{
		Version: 1,
		Inputs: []*proto.TxInput{{
			PrevTxHash:   util.RandomHash(),
			PrevOutIndex: 0,
			PublicKey:    pubKey.Bytes(),
		}},
		Outputs: []*proto.TxOutput{{
			Amount:  99,
			Address: pubKey.Address().Bytes(),
		}},
	}

	_, err = c.HandleTransaction(context.TODO(), tx, grpc.EmptyCallOption{})
	if err != nil {
		log.Fatal(err)
	}
}
