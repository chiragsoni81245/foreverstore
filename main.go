package main

import (
	"fmt"
	"log"

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
        OnPeer: OnPeer,
    }
    tr := p2p.NewTCPTransport(tcpOpts)
    
    go func () {
        ch := tr.Consume()
        for msg := range ch {
            fmt.Printf("Message: %+v\n", msg)
        }
    }()

    if err := tr.ListenAndAccept(); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Listening...")
    select {}
}

