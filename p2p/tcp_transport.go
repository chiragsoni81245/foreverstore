package p2p

import (
	"fmt"
	"net"
)

// TCPPeer represent the remote node over a TCP established connection.
type TCPPeer struct {
    // conn is the underlying connection of the peer
    conn net.Conn

    // if we initiate the connection ==> outbound == false
    // if we accept and retrieve a connection ==> outbound == true
    outbound bool
}

// Close implements the peer interface.
func (peer *TCPPeer) Close() error {
    return peer.conn.Close()
}

func NewTCPPeer(conn net.Conn, outboud bool) *TCPPeer {
    return &TCPPeer{
        conn: conn,
        outbound: outboud,
    }
}

type TCPTransportOpts struct {
    ListenAddr string
    HandshakeFunc HandshakeFunc
    Decoder Decoder
    OnPeer func(Peer) error
}

type TCPTrasport struct {
    TCPTransportOpts
    listener net.Listener
    rpcch chan RPC
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTrasport {
    if opts.HandshakeFunc == nil {
        opts.HandshakeFunc = NOPHandshakeFunc
    }
    if  opts.Decoder == nil {
        opts.Decoder = &DefaultDecoder{}
    }
    return &TCPTrasport{
        TCPTransportOpts: opts,
        rpcch: make(chan RPC),
    }
}

// Consume implements the transport interface, which will return a read only channel
// for reading incoming messages received from another peer in the network.
func (t *TCPTrasport) Consume() <-chan RPC {
    return t.rpcch
}

func (t *TCPTrasport) ListenAndAccept() error {
    var err error
    t.listener, err = net.Listen("tcp", t.ListenAddr)
    if err!=nil {
        return err
    }

    go t.startAcceptConnLoop() 

    return nil
}

func (t *TCPTrasport) startAcceptConnLoop() {
    for {
        conn, err := t.listener.Accept()
        if err!=nil {
            fmt.Printf("TCP accept error: %s\n", err)
        }

        go t.handleConn(conn)
    }
}

func (t *TCPTrasport) handleConn(conn net.Conn) {
    var err error
    defer func() {
        fmt.Printf("dropping peer connection: %s\n" , err) 
        conn.Close()
    }()

    peer := NewTCPPeer(conn, true)

    if err = t.HandshakeFunc(peer); err != nil {
        return
    }

    if t.OnPeer != nil {
        if err = t.OnPeer(peer); err != nil { 
            return 
        }
    }

    // Read loop
    rpc := RPC{From: conn.RemoteAddr()}
    for {
        // To-Do -  
        // figure out a way to identify the error type
        // so we can break the read loop only for case 
        // when a connection is already closed while reading from it
        if err = t.Decoder.Decode(conn, &rpc); err != nil { 
            return 
        }

        t.rpcch <- rpc 
    }
}
