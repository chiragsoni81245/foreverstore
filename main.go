package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
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
    data := []byte("my big data file yo hoooo --")
    dataReader := bytes.NewReader(data)
    if err := fs2.Store(key, dataReader); err != nil {
        log.Fatal(err)
    }

    time.Sleep(1*time.Second)
    fs2.store.Delete(key)

    r, _, err := fs2.Get(key)
    if err != nil {
        log.Fatal(err)
    }
    receivedData := new(bytes.Buffer)
    n, err := io.Copy(receivedData, r)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Received Data (%d bytes): '%s'", n, string(receivedData.Bytes()))

    select {}
}

