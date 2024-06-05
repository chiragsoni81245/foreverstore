package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"
)

func TestPathTransformFunc(t *testing.T) {
    key := "yo its a dummy string"
    pathKey := CASPathTransformFunc(key)
    filename := "90555eb6014736bedad917345c0193cd1f638ad6"
    expectedPathName := "90555/eb601/4736b/edad9/17345/c0193/cd1f6/38ad6"
    if pathKey.Filename != filename { 
        t.Errorf("have filename %s want %s", pathKey.Filename, filename)
    }
    if pathKey.Path != expectedPathName { 
        t.Errorf("have path name %s want %s", pathKey.Path, expectedPathName)
    }
}

func TestStore(t *testing.T) {
    store := newStore(t)
    defer teardown(t, store)

    for i:=0; i < 50; i++ {
        key := fmt.Sprintf("foo_%d", i)
        data := []byte("some random bytes")
        if err := store.writeStream(key, bytes.NewReader(data)); err != nil {
            t.Error(err)
        }

        if !store.Has(key) {
            t.Errorf("expected to have key %s", key)
        }

        r, err := store.Read(key)
        if err != nil {
            t.Error(err)
        }

        b, err := io.ReadAll(r)
        if err != nil {
            t.Error(err)
        }

        if string(b) != string(data) {
            t.Errorf("want \"%s\" have \"%s\"", data, b)
        }

        if err := store.Delete(key); err != nil {
            t.Error(err)
        }

        if store.Has(key) {
            t.Errorf("not expecting the key %s", key)
        }
    }
}

func newStore(t *testing.T) *Store {
    opts := StoreOpts{
        Root: fmt.Sprintf("./storage/%d", time.Now().UnixMilli()),
        PathTransformFunc: CASPathTransformFunc,        
    }
    store, err := NewStore(opts)
    if err != nil {
        t.Error(err)
    }
    return store
}

func teardown(t *testing.T, s *Store) {
    if err := s.Clear(); err != nil {
        t.Error(err)
    }
}

