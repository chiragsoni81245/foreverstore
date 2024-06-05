package p2p

import "net"

// RPC holds any arbitrary data that is being send over the
// each transport between two nodes.
type RPC struct {
    From net.Addr
    Payload []byte
}
