package node

import (
	"context"
	"fmt"

	"github.com/mhg14/ChlockBane/proto"
	"google.golang.org/grpc/peer"
)

// type Server struct {
// 	ln net.Listener
// 	listenAddr string
// }

// func New(listenAddr string) (*Server, error) {
// 	ln, err := net.Listen("tcp", listenAddr)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &Server{
// 		ln: ln,
// 		listenAddr: listenAddr,
// 	}, nil
// }

type Node struct {
	proto.UnimplementedNodeServer
	version string
}

func NewNode() *Node {
	return &Node{
		version: "ChlockBane-0.1",
	}
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)

	fmt.Println("received tx from", peer)
	return &proto.Ack{}, nil
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	ourVersion := &proto.Version{
		Version: n.version,
		Height:  100,
	}

	p, _ := peer.FromContext(ctx)
	fmt.Printf("recieved version from %s: %+v\n", v, p.Addr)
	return ourVersion, nil
}
