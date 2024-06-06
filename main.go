package main

import (
	"fmt"
	"time"

	"github.com/chiragsoni81245/foreverstore/p2p"
)

func OnPeer(peer p2p.Peer) error {
    fmt.Println("Got the peer, doing something with it outside of tcp transport...")
    peer.Close()
    return nil
}

func main(){
    tcpOpts := p2p.TCPTransportOpts{
        ListenAddr: ":4000",
        HandshakeFunc: p2p.NOPHandshakeFunc,
        Decoder: &p2p.DefaultDecoder{},
        // To-Do
        // onpeer func
    }
    tcpTransport := p2p.NewTCPTransport(tcpOpts)

    fileServerOpts := FileServerOpts{
       StorageRoot: "4000_network",
       Transport: tcpTransport,
       PathTransformFunc: CASPathTransformFunc,
    }

    fs := NewFileServer(fileServerOpts)

    go func() {
        time.Sleep(time.Second * 13)
        fs.Stop()
    }()

    fs.Start()
}

