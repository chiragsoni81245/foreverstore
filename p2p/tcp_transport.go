package p2p

import (
	"fmt"
	"log"
	"net"
	"sync"
)

// TCPPeer represent the remote node over a TCP established connection
type TCPPeer struct {
    // conn is the underlying connection of the peer
    conn net.Conn

    // if we initiate the connection ==> outbound == false
    // if we accept and retrieve a connection ==> outbound == true
    outbound bool
}

func NewTCPPeer(conn net.Conn, outboud bool) *TCPPeer {
    return &TCPPeer{
        conn: conn,
        outbound: outboud,
    }
}

type TCPTrasport struct {
    listenAddr string
    listener net.Listener
    shakeHands HandshakeFunc
    decoder Decoder

    mu sync.RWMutex // This mutex is to protect the peers in map in concurrent working settings
    peers map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTrasport {
    return &TCPTrasport{
        shakeHands: NOPHandshakeFunc,
        listenAddr: listenAddr,
    }
}

func (t *TCPTrasport) ListenAndAccept() error {
    var err error
    t.listener, err = net.Listen("tcp", t.listenAddr)
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

type Temp struct {}

func (t *TCPTrasport) handleConn(conn net.Conn) {
    _ = NewTCPPeer(conn, true)
    if err := t.shakeHands(conn); err != nil {
        conn.Close()
        log.Fatal(err) 
        return
    }

    // Read loop
    msg := &Temp{}
    for {
        if err := t.decoder.Decode(conn, msg); err != nil {
            fmt.Printf("TCP error: %s\n", err)
            continue
        }
    }
}
