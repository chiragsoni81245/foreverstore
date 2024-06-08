package main

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"
	"sync"

	"github.com/chiragsoni81245/foreverstore/p2p"
)

type FileServerOpts struct {
    StorageRoot string
    PathTransformFunc PathTransformFunc
    Transport p2p.Transport
    BootstrapNodes []string
}

type FileServer struct {
    FileServerOpts

    peerLock sync.Mutex
    peers map[string]p2p.Peer

    store *Store
    quitch chan struct{}
}

type Message struct {
    Payload any
}

type GetMessagePayload struct {
    Key string
}

type StoreMessagePayload struct {
    Key string
    Data []byte
}


func NewFileServer(opts FileServerOpts) *FileServer {
    gob.Register(&GetMessagePayload{})
    gob.Register(&StoreMessagePayload{})
    storeOpts := StoreOpts{
        Root: opts.StorageRoot,
        PathTransformFunc: opts.PathTransformFunc,
    }
    return &FileServer{
        FileServerOpts: opts,
        store: NewStore(storeOpts),
        quitch: make(chan struct{}),
        peers: make(map[string]p2p.Peer),
    }
}


func (fs *FileServer) bootstrapNetwork() error {
    if len(fs.BootstrapNodes)==0 {return nil}
    totalBootstrapedNodes := 0
    wg := sync.WaitGroup{}
    for _, addr := range fs.BootstrapNodes {
        wg.Add(1)
        go func(addr string, totalCount *int) {
            log.Printf("attempting to connect with remote: %s", addr)
            if err := fs.Transport.Dial(addr); err != nil {
                log.Println(err)
                wg.Done()
                return
            }
            totalBootstrapedNodes++
            wg.Done()
        }(addr, &totalBootstrapedNodes)
    }
    wg.Wait()
    return nil
}

func (fs *FileServer) loop() {
    defer func(){
        log.Printf("file server stopped due to user quit action")
        if err := fs.Transport.Close(); err != nil {
            log.Fatal(err)
        }
    }()
    for {
        select {
        case rpc := <-fs.Transport.Consume():
            var msg Message;
            if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
                log.Fatal(err)
            }
            if err := fs.handleMessage(&msg); err != nil {
                log.Fatal(err)
            }
        case <-fs.quitch:
            return
        }
    }
}

func (fs *FileServer) handleMessage(msg *Message) error {
    switch v := msg.Payload.(type) {
    case *StoreMessagePayload:
        log.Printf("store message received: %+v\n", v)
    case *GetMessagePayload:
        log.Printf("get message received: %+v\n", v)
    }
    return nil
}

func (fs *FileServer) broadcast(msg *Message) error {
    buf := new(bytes.Buffer)
    if err := gob.NewEncoder(buf).Encode(msg); err != nil {
        return err 
    }
    for _, peer := range fs.peers {
        peer.Send(buf.Bytes())
    }
    return nil
}

func (fs *FileServer) StoreFile(key string, r io.Reader) error{
    // 1. Store this file to disk
    // 2. Broadcast this file to the peer network
    
    buf := new(bytes.Buffer)
    tee := io.TeeReader(r, buf)

    if err := fs.store.Write(key, tee); err != nil {
        return err 
    }

    msg := &Message{
        Payload: &StoreMessagePayload{
            Key: key,
            Data: buf.Bytes(),
        },
    }

    return fs.broadcast(msg)
}

func (fs *FileServer) OnPeer(peer p2p.Peer) error{
    fs.peerLock.Lock()
    defer fs.peerLock.Unlock()

    fs.peers[peer.RemoteAddr().String()] = peer
    log.Printf("connected with remote %s", peer.RemoteAddr())
    return nil
}

func (fs *FileServer) Start() error{
    if err := fs.Transport.ListenAndAccept(); err != nil {
        return err
    }

    if err := fs.bootstrapNetwork(); err != nil {
        log.Println(err)
    }

    go fs.loop()    

    return nil
}

func (fs *FileServer) Stop() {
    close(fs.quitch)
}

