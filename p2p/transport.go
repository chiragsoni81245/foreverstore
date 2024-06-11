package p2p

import (
	"io"
	"net"
)

// Peer is an interface that represents the remote node
type Peer interface {
    net.Conn
    Send(t byte, r io.Reader, size int64) error
    CloseStream()
}

// Transport is anything that handles the communication
// between nodes in the network
// This can be of form TCP, UDP, or websockets
type Transport interface {
    ListenAndAccept() error
    Dial(string) error
    Consume() <-chan RPC
    Close() error
}
