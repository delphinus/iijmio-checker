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
	if st, err = os.Stat(filename); st.IsDir() || os.IsNotExist(err) {
		isNew = true
	}
	var f *os.File
	f, err = os.Open(filename)
	if err != nil {
		return
	}
	defer closer(f, &err)
	if isNew {
		cfg.HashKey = securecookie.GenerateRandomKey(64)
		cfg.BlockKey = securecookie.GenerateRandomKey(32)
		err = gob.NewEncoder(f).Encode(cfg)
	} else {
		err = gob.NewDecoder(f).Decode(cfg)
	}
	return
}
