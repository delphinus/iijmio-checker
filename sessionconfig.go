package main

import (
	"encoding/gob"
	"os"
	"path/filepath"

	"github.com/gorilla/securecookie"
)

// SessionConfig stores setting for session
type SessionConfig struct {
	HashKey  []byte
	BlockKey []byte
}

func sessionConfig(filename string) (cfg *SessionConfig, err error) {
	dir := filepath.Dir(filename)
	var st os.FileInfo
	if st, err = os.Stat(dir); os.IsNotExist(err) || !st.IsDir() {
		if err = os.MkdirAll(dir, 0700); err != nil {
			return
		}
	}
	var isNew bool
	var f *os.File
	st, err = os.Stat(filename)
	if err == nil && st.IsDir() || os.IsNotExist(err) {
		f, err = os.Create(filename)
		isNew = true
	} else {
		f, err = os.Open(filename)
	}
	if err != nil {
		return
	}
	defer closer(f, &err)
	if isNew {
		cfg = &SessionConfig{
			HashKey:  securecookie.GenerateRandomKey(64),
			BlockKey: securecookie.GenerateRandomKey(32),
		}
		err = gob.NewEncoder(f).Encode(cfg)
	} else {
		err = gob.NewDecoder(f).Decode(cfg)
	}
	return
}
