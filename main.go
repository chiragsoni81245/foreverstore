package main

import (
	"fmt"
	"log"

	"github.com/chiragsoni81245/foreverstore/p2p"
)

func main(){
    tr := p2p.NewTCPTransport(":4000")
    
    if err := tr.ListenAndAccept(); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Listening...")
    select {}
}

