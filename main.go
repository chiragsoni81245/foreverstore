package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/chiragsoni81245/foreverstore/p2p"
)

func OnPeer(peer p2p.Peer) error {
    fmt.Println("Got the peer, doing something with it outside of tcp transport...")
    peer.Close()
    return nil
}

func makeserver(listenAddr string, nodes ...string) *FileServer {
    tcpOpts := p2p.TCPTransportOpts{
        ListenAddr: listenAddr,
        HandshakeFunc: p2p.NOPHandshakeFunc,
        Decoder: &p2p.DefaultDecoder{},
    }
    tcpTransport := p2p.NewTCPTransport(tcpOpts)

    fileServerOpts := FileServerOpts{
       StorageRoot: fmt.Sprintf("%s_network", listenAddr),
       Transport: tcpTransport,
       PathTransformFunc: CASPathTransformFunc,
       BootstrapNodes: nodes,
    }
    fs := NewFileServer(fileServerOpts)
    tcpTransport.OnPeer = fs.OnPeer
    return fs
}

func main(){
    fs1 := makeserver(":4000")
    fs1.Start()

    time.Sleep(1*time.Second)

    fs2 := makeserver(":5000", ":4000")
    fs2.Start()
    time.Sleep(1*time.Second)
    
    data := bytes.NewReader([]byte("my big data file!"))
    if err := fs2.StoreFile("myprivatedata", data); err != nil {
        log.Fatal(err)
    }
    select {}
}

