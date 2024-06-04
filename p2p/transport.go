package p2p

// Peer is an interface that represents the remote node
type Peer interface {
}

// Transport is anything that handles the communication
// between nodes in the network
// This can be of form TCP, UDP, or websockets
type Transport interface {
    ListenAndAccept() error
}
