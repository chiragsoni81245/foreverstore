package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/chiragsoni81245/foreverstore/p2p"
)


func makeserver(listenAddr string, nodes ...string) *FileServer {
    tcpOpts := p2p.TCPTransportOpts{
        ListenAddr: listenAddr,
        HandshakeFunc: p2p.DiffieHallmanHandshake,
        Decoder: &p2p.DefaultDecoder{},
    }
    tcpTransport := p2p.NewTCPTransport(tcpOpts)

    fileServerOpts := FileServerOpts{
       StorageRoot: fmt.Sprintf("storage/%s_network", listenAddr),
       Transport: tcpTransport,
       PathTransformFunc: CASPathTransformFunc,
       BootstrapNodes: nodes,
       EncryptionKey: []byte("rptreftgrtgfrefrdeswfrdefrdejtkg"),
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
    
    key := "myprivatedata"
    f, err := os.Open("test_file")
    if err := fs2.Store(key, f); err != nil {
        log.Fatal(err)
    }
    f.Close()

    time.Sleep(1*time.Second)
    fs2.store.Delete(key)

    r, _, err := fs2.Get(key)
    if err != nil {
        log.Fatal(err)
    }

    f, err = os.Create("output_file")
    n, err := io.Copy(f, r)
    if err != nil {
        log.Fatal(err)
    }
    f.Close()
    
    log.Printf("Received %d bytes", n)

    select {}
}

