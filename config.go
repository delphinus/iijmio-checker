package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

var (
	// JST is a time zone for Asia/Tokyo
	JST *time.Location
	// ErrNoToken means config has no token
	ErrNoToken = errors.New("config has no token")
	// ErrTokenExpired means token expired
	ErrTokenExpired = errors.New("token expired")
	// ErrInvalidExpiresIn means expires_in has invalid string
	ErrInvalidExpiresIn = errors.New("expires_in has invalid string")
)

func init() {
	var err error
	JST, err = time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
}

// Config stores config for this app
type Config struct {
	Token     string `json:"token"`
	ExpiresIn string `json:"expires_in"`
}

// ExpiresAt returns time.Time converted from ExpiresIn string
func (c *Config) ExpiresAt() (time.Time, error) {
	return time.ParseInLocation(time.RFC3339, c.ExpiresIn, JST)
}

// Validate validates the config is valid and has not expired.
func (c *Config) Validate() error {
	if c.Token == "" {
		return ErrNoToken
	}
	t, err := c.ExpiresAt()
	if err != nil {
		return ErrInvalidExpiresIn
	}
	if t.Before(time.Now()) {
		return ErrTokenExpired
	}
	return nil
}

func loadConfig(filename string) (cfg *Config, err error) {
	var st os.FileInfo
	if st, err = os.Stat(filename); os.IsNotExist(err) || st.IsDir() {
		return nil, fmt.Errorf("config file not found: %s", filename)
	}
	var f *os.File
	f, err = os.Open(filename)
	if err != nil {
		return
	}
	defer closer(f, &err)
	err = json.NewDecoder(f).Decode(cfg)
	return
}

func saveConfig(
	filename string, token string, expiresInSecond int,
) (err error) {
	dir := filepath.Dir(filename)
	var st os.FileInfo
	if st, err = os.Stat(dir); os.IsNotExist(err) || !st.IsDir() {
		if err = os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	expiresIn := time.Now().Add(time.Second * time.Duration(expiresInSecond))
	cfg := Config{
		Token:     token,
		ExpiresIn: expiresIn.In(JST).Format(time.RFC3339),
	}
	var f *os.File
	f, err = os.Create(filename)
	if err != nil {
		return err
	}
	defer closer(f, &err)
	return json.NewEncoder(f).Encode(&cfg)
}

func closer(c io.Closer, err *error) {
	if cerr := c.Close(); cerr != nil && err == nil {
		*err = cerr
	}
}
