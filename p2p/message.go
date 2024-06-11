package p2p

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

// RPC holds any arbitrary data that is being send over the
// each transport between two nodes.


type RPC struct {
    From string
    Size int64
    Stream bool
    Payload []byte
}
