package p2p

import (
	"crypto/ecdh"
	"crypto/rand"
	"net"
)

type HandshakeFunc func (net.Conn) (secretKey []byte, err error)

func NOPHandshakeFunc(conn net.Conn) (secretKey []byte, err error) {
    return []byte{}, nil
}

func DiffieHallmanHandshake(conn net.Conn) (secretKey []byte, err error) {
    curve := ecdh.P256()
    privateKey, err := curve.GenerateKey(rand.Reader)
    if err != nil {
        return []byte{}, err
    }
    publicKey := privateKey.PublicKey()
    
    conn.Write(publicKey.Bytes())

    peerPublicKeyBytes := make([]byte, 65) 
    _, err = conn.Read(peerPublicKeyBytes)
    if err != nil {
        return []byte{}, err
    }
    peerPublicKey, err := curve.NewPublicKey(peerPublicKeyBytes)
    if err != nil {
        return []byte{}, err
    }
    
    secretKey, err = privateKey.ECDH(peerPublicKey)
    if err != nil {
        return []byte{}, err
    }

    return secretKey, nil
}
