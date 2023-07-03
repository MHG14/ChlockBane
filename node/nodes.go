package node

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/mhg14/ChlockBane/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

type Node struct {
	proto.UnimplementedNodeServer

	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version

	version    string
	listenAddr string
	logger     *zap.SugaredLogger
}

func (n *Node) Start(listenAddr string) error {
	n.listenAddr = listenAddr
	grpcServer := grpc.NewServer(grpc.EmptyServerOption{})

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	proto.RegisterNodeServer(grpcServer, n)
	n.logger.Infow("node started...", "port", n.listenAddr)
	return grpcServer.Serve(ln)
}

func NewNode() *Node {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.TimeKey = ""
	logger, _ := loggerConfig.Build()

	return &Node{
		version: "ChlockBane-0.1",
		peers:   make(map[proto.NodeClient]*proto.Version),
		logger:  logger.Sugar(),
	}
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)

	fmt.Println("received tx from", peer)
	return &proto.Ack{}, nil
}

func (n *Node) BootstrapNetwork(addr []string) error {
	for _, address := range addr {
		c, err := makeNodeClient(address)
		if err != nil {
			return err
		}

		v, err := c.Handshake(context.Background(), n.getVersion())
		if err != nil {
			n.logger.Error("handshake error:", err)
			continue
		}

		n.addPeer(c, v)
	}
	return nil
}

func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	n.logger.Debugw("new peer connected", "dest", n.listenAddr, "height", v.Height)

	n.peers[c] = v
}

func (n *Node) removePeer(c proto.NodeClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	delete(n.peers, c)
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	c, err := makeNodeClient(v.ListenAddr)
	if err != nil {
		return nil, err
	}

	n.addPeer(c, v)

	return n.getVersion(), nil
}

func makeNodeClient(listenAddr string) (proto.NodeClient, error) {
	c, err := grpc.Dial(listenAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return proto.NewNodeClient(c), nil
}

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    "ChlockBane-0.1",
		Height:     0,
		ListenAddr: n.listenAddr,
	}
}
