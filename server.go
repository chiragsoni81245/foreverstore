package main

import "github.com/chiragsoni81245/foreverstore/p2p"

type FileServerOpts struct {
    StorageRoot string
    PathTransformFunc PathTransformFunc
    Transport p2p.Transport
}

type FileServer struct {
    FileServerOpts
    store *Store
}

func NewFileServer(opts FileServerOpts) *FileServer {
    storeOpts := StoreOpts{
        Root: opts.StorageRoot,
        PathTransformFunc: opts.PathTransformFunc,
    }
    return &FileServer{
        FileServerOpts: opts,
        store: NewStore(storeOpts),
    }
}

func (fs *FileServer) Start() error {
    if err := fs.Transport.ListenAndAccept(); err != nil {
        return err
    }

    return nil
}
