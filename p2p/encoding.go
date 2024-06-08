package p2p

import (
	"encoding/binary"
	"io"
)

type Decoder interface {
    Decode(io.Reader, *RPC) error
}

type DefaultDecoder struct {
}

func (dec *DefaultDecoder) Decode(r io.Reader, msg *RPC) error{
    // Get content length
    // then collect streaming data of that content length

    contentLengthBytes := make([]byte, 8)
    n, err := r.Read(contentLengthBytes)
    if err != nil {
        return err
    }
    contentLength := int(binary.BigEndian.Uint64(contentLengthBytes[:n]))

    for {
        buf := make([]byte, 1)
        n, err := r.Read(buf)
        if err != nil {
            return err
        }

        remainingLength := contentLength - len(msg.Payload)
        var payload []byte
        if n > remainingLength {
            payload = buf[:remainingLength]
        } else {
            payload = buf[:n]
        }

        for _, b := range payload {
            msg.Payload = append(msg.Payload, b)
        }

        if contentLength == len(msg.Payload) {
            break
        }
    }
    return nil
}

