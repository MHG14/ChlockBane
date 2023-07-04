package node

import (
	"context"
	"encoding/hex"
	"net"
	"sync"
	"time"

	"github.com/mhg14/ChlockBane/crypto"
	"github.com/mhg14/ChlockBane/proto"
	"github.com/mhg14/ChlockBane/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

const blockTime = time.Second * 5

type Mempool struct {
	lock sync.RWMutex
	txx  map[string]*proto.Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		txx: make(map[string]*proto.Transaction),
	}
}

func (mp *Mempool) Clear() []*proto.Transaction {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	txx := make([]*proto.Transaction, len(mp.txx))
	it := 0

	for k, v := range mp.txx {
		delete(mp.txx, k)
		txx[it] = v
		it++
	}

	return txx
}

func (mp *Mempool) Len() int {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return len(mp.txx)
}

func (mp *Mempool) Has(tx *proto.Transaction) bool {
	mp.lock.RLock()
	defer mp.lock.RUnlock()
	txHash := hex.EncodeToString(types.HashTransaction(tx))
	_, ok := mp.txx[txHash]
	return ok
}

func (mp *Mempool) Add(tx *proto.Transaction) bool {
	if mp.Has(tx) {
		return false
	}

	mp.lock.Lock()
	defer mp.lock.Unlock()

	txHash := hex.EncodeToString(types.HashTransaction(tx))
	mp.txx[txHash] = tx
	return true
}

type Node struct {
	proto.UnimplementedNodeServer

	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version

	logger  *zap.SugaredLogger
	mempool *Mempool
	ServerConfig
}

type ServerConfig struct {
	Version    string
	ListenAddr string
	PrivateKey *crypto.PrivateKey
}

func NewNode(cfg ServerConfig) *Node {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.TimeKey = ""
	logger, _ := loggerConfig.Build()

	return &Node{
		peers:        make(map[proto.NodeClient]*proto.Version),
		logger:       logger.Sugar(),
		mempool:      NewMempool(),
		ServerConfig: cfg,
	}
}

func (n *Node) Start(listenAddr string, bootstrapNodes []string) error {
	n.ListenAddr = listenAddr
	grpcServer := grpc.NewServer(grpc.EmptyServerOption{})

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	proto.RegisterNodeServer(grpcServer, n)
	n.logger.Infow("node started...", "port", n.ListenAddr)

	if len(bootstrapNodes) > 0 {
		go n.bootstrapNetwork(bootstrapNodes)
	}

	if n.PrivateKey != nil {
		go n.validatorLoop()
	}

	return grpcServer.Serve(ln)
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)
	hash := hex.EncodeToString(types.HashTransaction(tx))

	if n.mempool.Add(tx) {
		n.logger.Debugw("reciecved tx", "we", n.ListenAddr, "from", peer.Addr, "hash", hash)

		go func() {
			if err := n.broadcast(tx); err != nil {
				n.logger.Errorw("broadcast error", "err", err)
			}
		}()
	}

	return &proto.Ack{}, nil
}

func (n *Node) broadcast(msg any) error {
	for peer := range n.peers {
		switch v := msg.(type) {
		case *proto.Transaction:
			_, err := peer.HandleTransaction(context.Background(), v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	c, err := makeNodeClient(v.ListenAddr)
	if err != nil {
		return nil, err
	}

	n.addPeer(c, v)

	return n.getVersion(), nil
}

func (n *Node) validatorLoop() {
	n.logger.Infow("starting validator loop", "pubKey", n.PrivateKey.Public(), "blockTime", blockTime)
	ticker := time.NewTicker(blockTime)
	for {
		<-ticker.C

		txx := n.mempool.Clear()

		n.logger.Debugw("time to create a new block", "lenTx", len(txx))


	}
}

func (n *Node) bootstrapNetwork(addr []string) error {
	for _, address := range addr {
		if !n.canConnectWith(address) {
			continue
		}
		n.logger.Debugw("dialing remote node", "we", n.ListenAddr, "remoteNode", address)
		if address == n.ListenAddr {
			continue
		}
		c, v, err := n.dialRemoteNode(address)
		if err != nil {
			return err
		}
		n.addPeer(c, v)
	}
	return nil
}

func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	n.peers[c] = v

	// Connect to all peers in the recieved list of peers
	if len(v.PeerList) > 0 {
		go n.bootstrapNetwork(v.PeerList)
	}

	n.logger.Debugw("new peer successfully connected",
		"we", n.ListenAddr,
		"remoteNode", v.ListenAddr,
		"height", v.Height,
	)
}

// func (n *Node) removePeer(c proto.NodeClient, v *proto.Version) {
// 	n.peerLock.Lock()
// 	defer n.peerLock.Unlock()

// 	delete(n.peers, c)
// }

func (n *Node) dialRemoteNode(addr string) (proto.NodeClient, *proto.Version, error) {
	c, err := makeNodeClient(addr)
	if err != nil {
		return nil, nil, err
	}

	v, err := c.Handshake(context.Background(), n.getVersion())
	if err != nil {
		return nil, nil, err
	}

	return c, v, nil
}

func makeNodeClient(listenAddr string) (proto.NodeClient, error) {
	c, err := grpc.Dial(listenAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return proto.NewNodeClient(c), nil
}

func (n *Node) getPeerList() []string {
	n.peerLock.RLock()
	defer n.peerLock.RUnlock()

	peers := []string{}

	for _, version := range n.peers {
		peers = append(peers, version.ListenAddr)
	}

	return peers
}

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    "ChlockBane-0.1",
		Height:     0,
		ListenAddr: n.ListenAddr,
		PeerList:   n.getPeerList(),
	}
}

func (n *Node) canConnectWith(addr string) bool {
	if n.ListenAddr == addr {
		return false
	}
	connectedPeers := n.getPeerList()
	for _, connectedAddr := range connectedPeers {
		if addr == connectedAddr {
			return false
		}
	}
	return true
}
