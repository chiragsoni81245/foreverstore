package p2p

import (
	"bytes"
	"testing"
)

func TestCopyEncryptDecrypt(t *testing.T) {
    data := "private data code 101"
    src := bytes.NewReader([]byte(data))
    cipherOut := new(bytes.Buffer)
    key := []byte("asdwedscfrgtvfdybhghnjhgrfdrefrd")

    _, err := CopyEncrypt(key, cipherOut, src)
    if err != nil {
        t.Error(err)
    }

    if 16+len(data) != cipherOut.Len() {
        t.Fail()
    }

    decryptOut := new(bytes.Buffer)
    _, err = CopyDecrypt(key, decryptOut, cipherOut)
    if err != nil {
        t.Error(err)
    }

    if decryptOut.String() != data {
        t.Errorf("decryption failed")
    }
}
