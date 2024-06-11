package p2p

import (
	"encoding/binary"
	"io"
	"log"
)

type Decoder interface {
    Decode(io.Reader, *RPC) error
}

type DefaultDecoder struct {
}

func (dec *DefaultDecoder) Decode(r io.Reader, rpc *RPC) error{
    // Get content length
    // then collect streaming data of that content length

    var incomingDataType byte
    incomingDataTypeBytes := make([]byte, 1)
    n, err := r.Read(incomingDataTypeBytes)
    if err != nil {
        log.Println(err)
        return err
    }
    incomingDataType = incomingDataTypeBytes[:n][0]
    if incomingDataType==IncomingStream {
        rpc.Stream = true
        return nil 
    }

    var contentLength int64
    binary.Read(r, binary.LittleEndian, &contentLength)
    rpc.Size = contentLength

    buf := make([]byte, contentLength)
    n, err = r.Read(buf)
    if err != nil {
        return err
    }
    rpc.Payload = buf[:n]

    return nil
}

