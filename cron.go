package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"gopkg.in/urfave/cli.v2"
)

const (
	// APIURL stores URL for IIJmio API
	APIURL = "https://api.iijmio.jp/mobile/d/v2/log/packet/"
)

func cron(cc *cli.Context) (err error) {
	developerID := os.Getenv(envName)
	if developerID == "" {
		return fmt.Errorf("set developerID in %s", envName)
	}
	var cfg *Config
	cfg, err = loadConfig(cc.String("config"))
	if err != nil {
		return err
	} else if err = cfg.Validate(); err != nil {
		return fmt.Errorf("トークンが異常です: %v", err)
	}
	var req *http.Request
	req, err = http.NewRequest("GET", APIURL, nil)
	if err != nil {
		return err
	}
	req.Header.Add("X-IIJmio-Developer", developerID)
	req.Header.Add("X-IIJmio-Authorization", cfg.Token)
	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer closer(resp.Body, &err)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Fprintf(cc.App.Writer, "%s", body)
	return nil
}
