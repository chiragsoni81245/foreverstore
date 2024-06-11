package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/chiragsoni81245/foreverstore/p2p"
)


type Message struct {
    Payload any
}

type MessageGetFile struct {
    Key string
}

type MessageStoreFile struct {
    Key string
    Size int64
}

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

func NewFileServer(opts FileServerOpts) *FileServer {
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

func (fs *FileServer) getPeer(addr string) (p2p.Peer, error) {
    fs.peerLock.Lock()
    defer fs.peerLock.Unlock()

    peer, ok := fs.peers[addr]
    if !ok {
        return nil, fmt.Errorf("peer not found in peer list")
    }

    return peer, nil
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

            if err := fs.handleMessage(&rpc, &msg); err != nil {
                log.Fatal(err)
            }
        case <-fs.quitch:
            return
        }
    }
}

func (fs *FileServer) handleMessage(rpc *p2p.RPC, msg *Message) error {
    switch payload := msg.Payload.(type) {
    case *MessageStoreFile:
        return fs.handleStoreFileMessage(rpc, payload)
    case *MessageGetFile:
        return fs.handleGetFileMessage(rpc, payload)
    }
    return nil
}

func (fs *FileServer) handleStoreFileMessage(rpc *p2p.RPC, msgPayload *MessageStoreFile) error {
    log.Printf("store message received: %+v\n", msgPayload)

    peer, err := fs.getPeer(rpc.From)
    if err != nil {
        return err
    }

    if _, err := fs.store.Write(msgPayload.Key, io.LimitReader(peer, msgPayload.Size)); err != nil {
        return err
    }

    // Telling the peer read loop that we have read the stream they can start reading for new data
    peer.CloseStream()

    return nil
}

func (fs *FileServer) handleGetFileMessage(rpc *p2p.RPC, msgPayload *MessageGetFile) error {
    log.Printf("get message received: %+v\n", msgPayload)

    var (
        r io.Reader
        size int64
        err error
    )
    
    if !fs.store.Has(msgPayload.Key) {
        log.Printf("file requested via peer %s not found", rpc.From)
    } else {
        r, size, err = fs.store.Read(msgPayload.Key)
        if err != nil {
            return err
        }
    }

    peer, err := fs.getPeer(rpc.From)
    if err != nil {
        return err
    }

    // Initializing Stream
    // To-Do
    // Find a way so that while sending only IncomingStream type message we don't have to send these default arguments of (nil, 0) for reader and size
    err = peer.Send(p2p.IncomingStream, nil, 0)

    // Sending file size into stream
    err = binary.Write(peer, binary.LittleEndian, size)
    if err != nil { 
        return err
    }
    
    if size != 0 && r != nil {
        // Sending file content
        _, err = io.Copy(peer, r)
    }

    return err
}

func (fs *FileServer) broadcast(msg *Message, r io.Reader) error {
    buf := new(bytes.Buffer)
    if err := gob.NewEncoder(buf).Encode(msg); err != nil {
        return err 
    }
    for _, peer := range fs.peers {
        // Send message
        err := peer.Send(p2p.IncomingMessage, buf, int64(buf.Len()))
        if err != nil {
            log.Printf("error in sending message to peer %s", peer.RemoteAddr().String())
        }

        switch payload := msg.Payload.(type) {
        case *MessageStoreFile:
            // Stream data
            peer.Send(p2p.IncomingStream, r, payload.Size)
        }
    }
    return nil
}

func (fs *FileServer) Get(key string) (io.Reader, int64, error) {
    if fs.store.Has(key) {
        return fs.store.Read(key)
    } else {
        // Local store does not have file associated with key
        // Checking on network peers
        log.Printf("requested file not found on local, checking on peer network")

        getFileMsg := &Message{
            Payload: &MessageGetFile{
                Key: key,
            },
        }
        getFileMsgBuf := new(bytes.Buffer)
        if err := gob.NewEncoder(getFileMsgBuf).Encode(getFileMsg); err != nil {
            return nil, 0, err 
        }

        fs.peerLock.Lock()
        defer fs.peerLock.Unlock()
        
        for _, peer := range fs.peers {
            err := peer.Send(p2p.IncomingMessage, getFileMsgBuf, int64(getFileMsgBuf.Len()))
            if err != nil {
                log.Printf("error in sending message to peer %s", peer.RemoteAddr().String())
            }
            
            // To-Do
            // Try to find a way to check if the peer is in streaming mode or not if not then wait for the stream
            // because if the read loop is not paused so there can be two readers reading from peer at the same time causing a race condition
            // for now using time.Sleep to mimic the wait

            var fileSize int64
            err = binary.Read(peer, binary.LittleEndian, &fileSize)
            if err != nil {
                return nil, 0, err
            }

            if fileSize == 0 {
                log.Printf("file does not exists on peer (%s)", peer.LocalAddr().String())
                peer.CloseStream()
                continue
            } else {
                _, err := fs.store.Write(key, io.LimitReader(peer, fileSize)) 
                if err != nil {
                    return nil, 0, err
                }
                peer.CloseStream()
                return fs.store.Read(key) 
            }
        }

        return nil, 0, fmt.Errorf("file not found")
    }
}

func (fs *FileServer) Store(key string, r io.Reader) error{
    // 1. Store this file to disk
    // 2. Broadcast this file to the peer network
    
    buf := new(bytes.Buffer)
    tee := io.TeeReader(r, buf)

    size, err := fs.store.Write(key, tee); 
    if err != nil {
        return err 
    }

    msg := &Message{
        Payload: &MessageStoreFile{
            Key: key,
            Size: size, 
        },
    }
    
    return fs.broadcast(msg, buf)
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

func init() {
    gob.Register(&MessageGetFile{})
    gob.Register(&MessageStoreFile{})
}

