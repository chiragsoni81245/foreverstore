package p2p

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

// TCPPeer represent the remote node over a TCP established connection
type TCPPeer struct {
    net.Conn

    // streamTriggerCh is to send the events when a stream start and stops
    streamTriggerCh chan struct{}

    // secretKey will be the key to used in encryption and decryption of traffic 
    iv []byte
    peerIV []byte
    secretKey []byte

    // if we initiate the connection ==> outbound == false
    // if we accept and retrieve a connection ==> outbound == true
    outbound bool
}

// Write function takes type byte, io.Reader and that reader's size first send the byte type 
// so remote peer will be ready for steam or normal message based on this type 
// if t == IncomingMessage --> then content length will be transferred and then the actual message bytes will be sent
// if t == IncomingStream --> then directly the stream bytes will be sent
func (peer *TCPPeer) Send(t byte, r io.Reader, size int64) error {
    // Send incoming data type to remote peer
    _, err := peer.Write([]byte{t})
    if err != nil {
        return err
    }

    if r == nil || size == 0 { return nil }

    if t == IncomingMessage {
        err := binary.Write(peer, binary.LittleEndian, size)
        if err != nil {
            return err
        }
    }
    
    // Send message or stream bytes to remote peer
    _, err = io.Copy(peer, r)
    if err != nil {
        return err
    }

    return nil
}

func (peer *TCPPeer) Write(b []byte) (n int, err error) {
    cb := new(bytes.Buffer)
    n, err = CopyEncrypt(peer.secretKey, cb, bytes.NewReader(b), peer.iv)
    if(err != nil) {
        return 0, err
    }

    _, err = peer.Conn.Write(cb.Bytes())
    if(err != nil) {
        return 0, err
    }
    
    return len(b), nil
}

func (peer *TCPPeer) Read(b []byte) (n int, err error) {
    buf := new(bytes.Buffer) 
    cb := make([]byte, cap(b)) 

    n, err = peer.Conn.Read(cb)
    if(err != nil) {
        return 0, err
    }

    n, err = CopyDecrypt(peer.secretKey, buf, bytes.NewReader(cb[:n]), peer.peerIV)
    if(err != nil) {
        return 0, err
    }
    copy(b, buf.Bytes())

    return buf.Len(), nil
}

// ReadStream function implements Peer interface
// it will be used when user want to use a specific size of stream
func (peer *TCPPeer) ReadStream(size int64) io.Reader {
    <- peer.streamTriggerCh
    return io.LimitReader(peer, size) 
}

// CloseStream function implements Peer interface
// it is used to continue the read loop of messages after reading the stream of previous message from peer connection reader
func (peer *TCPPeer) CloseStream() {
    peer.streamTriggerCh <- struct{}{} 
}

func NewTCPPeer(conn net.Conn, outboud bool) *TCPPeer {
    return &TCPPeer{
        Conn: conn,
        outbound: outboud,
        streamTriggerCh: make(chan struct{}),
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

    secretKey, iv, peerIV, err := t.HandshakeFunc(conn)
    peer.secretKey = secretKey
    peer.peerIV = peerIV
    peer.iv = iv 

    if err != nil {
        return
    }

    if t.OnPeer != nil {
        if err = t.OnPeer(peer); err != nil { 
            return 
        }
    }

    // Read loop
    for {
        rpc := RPC{}
        // To-Do -  
        // figure out a way to identify the error type
        // so we can break the read loop only for case 
        // when a connection is already closed while reading from it
        if err = t.Decoder.Decode(peer, &rpc); err != nil { 
            return 
        }

        rpc.From = conn.RemoteAddr().String()

        if rpc.Stream {
            peer.streamTriggerCh <- struct{}{} // Send the trigger which will result in stream start
            <- peer.streamTriggerCh // Wait for the trigger which will result in stream stop
            continue
        }

        t.rpcch <- rpc 
    }

}
