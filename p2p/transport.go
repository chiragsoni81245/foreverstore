package p2p

import "net"

// Peer is an interface that represents the remote node
type Peer interface {
    Send([]byte) error
    RemoteAddr() net.Addr
    Close() error
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
