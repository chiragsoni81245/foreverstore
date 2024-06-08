package p2p

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
)


// TCPPeer represent the remote node over a TCP established connection
type TCPPeer struct {
    net.Conn

    // if we initiate the connection ==> outbound == false
    // if we accept and retrieve a connection ==> outbound == true
    outbound bool
}

// Write function takes bytes slice and first send the content length of those bytes to peer and then stream the bytes
func (peer *TCPPeer) Send(b []byte) error {
    contentLength := uint64(len(b))
    contentLengthBytes := make([]byte, 8)
    binary.BigEndian.PutUint64(contentLengthBytes, contentLength)

    _, err := peer.Conn.Write(contentLengthBytes)
    if err != nil {
        return err
    }

    _, err = peer.Conn.Write(b)
    return err
}

func NewTCPPeer(conn net.Conn, outboud bool) *TCPPeer {
    return &TCPPeer{
        Conn: conn,
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
// for reading incoming messages received from another peer in the network
func (t *TCPTrasport) Consume() <-chan RPC {
    return t.rpcch
}

// Close implements the Transport interface
func (t *TCPTrasport) Close() error {
    return t.listener.Close()
}

func (t *TCPTrasport) ListenAndAccept() error {
    var err error
    t.listener, err = net.Listen("tcp", t.ListenAddr)
    if err!=nil {
        return err
    }

    go t.startAcceptConnLoop() 

    log.Printf("TCP transport listening on: %s\n", t.ListenAddr)

    return nil
}

// Dial implements the Transport interface
func (t *TCPTrasport) Dial(addr string) error {
    conn, err := net.Dial("tcp", addr)
    if err != nil {
        return err
    }
    go t.handleConn(conn, true)

    return nil
}

func (t *TCPTrasport) startAcceptConnLoop() {
    for {
        conn, err := t.listener.Accept()
        if errors.Is(err, net.ErrClosed) {
            return
        }
        if err!=nil {
            fmt.Printf("TCP accept error: %s\n", err)
        }

        go t.handleConn(conn, false)
    }
}

func (t *TCPTrasport) handleConn(conn net.Conn, outbound bool) {
    var err error
    defer func() {
        fmt.Printf("dropping peer connection: %s\n" , err) 
        conn.Close()
    }()

    peer := NewTCPPeer(conn, outbound)

    if err = t.HandshakeFunc(peer); err != nil {
        return
    }

    if t.OnPeer != nil {
        if err = t.OnPeer(peer); err != nil { 
            return 
        }
    }

    // Read loop
    rpc := RPC{From: conn.RemoteAddr().String()}
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
