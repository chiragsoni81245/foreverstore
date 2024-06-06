package main

import (
	"fmt"
	"log"

	"github.com/chiragsoni81245/foreverstore/p2p"
)

type FileServerOpts struct {
    StorageRoot string
    PathTransformFunc PathTransformFunc
    Transport p2p.Transport
}

type FileServer struct {
    FileServerOpts
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
    }
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
        case msg := <-fs.Transport.Consume():
            fmt.Println(msg)
        case <-fs.quitch:
            return
        }
    }
}

func (fs *FileServer) Stop() {
    close(fs.quitch)
}

func (fs *FileServer) Start() {
    if err := fs.Transport.ListenAndAccept(); err != nil {
        log.Fatal(err)
        return
    }

    fs.loop()
}
