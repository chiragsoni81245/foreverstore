package p2p

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func NewEncryptionKey() []byte {
    keyBuf := make([]byte, 32)
    io.ReadFull(rand.Reader, keyBuf)
    return keyBuf
}

func CopyDecrypt(key []byte, dst io.Writer, src io.Reader) (int, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return 0, err
    }

    // Read the id from the given src io.Reader and the size of it will be block.BlockSize()
    iv := make([]byte, block.BlockSize())
    if _, err := src.Read(iv); err != nil {
        return 0, err
    }

    totalBytesWrote := 0

    var (
        buf = make([]byte, 32*1024)
        stream = cipher.NewCTR(block, iv)
    )

    for {
        var wn int
        rn, err := src.Read(buf)
        if rn > 0 {
            stream.XORKeyStream(buf, buf[:rn])
            if wn, err = dst.Write(buf[:rn]); err != nil {
                return 0, err
            }

            totalBytesWrote += wn
        }

        if err == io.EOF {
            break
        }
    }
    
    return totalBytesWrote, nil
}

func CopyEncrypt(key []byte, dst io.Writer, src io.Reader) (int, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return 0, err
    }

    iv := make([]byte, block.BlockSize())
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return 0, err
    }

    // Prepend the iv to the file so that it can be used for decryption
    if  _, err := dst.Write(iv); err != nil {
        return 0, err
    }

    totalBytesWrote := len(iv)

    var (
        buf = make([]byte, 32*1024)
        stream = cipher.NewCTR(block, iv)
    )

    for {
        var wn int
        rn, err := src.Read(buf)
        if rn > 0 {
            stream.XORKeyStream(buf, buf[:rn])
            if wn, err = dst.Write(buf[:rn]); err != nil {
                return 0, err
            }

            totalBytesWrote += wn
        }

        if err == io.EOF {
            break
        }
    }

    return totalBytesWrote, nil 
}
