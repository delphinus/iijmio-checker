package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"gopkg.in/urfave/cli.v2"
)

const (
	// APIURL stores URL for IIJmio API
	APIURL = "https://api.iijmio.jp/mobile/d/v2/log/packet/"
	// THRESHOLD is a threshold to report over limit (unit: MB)
	THRESHOLD = 150
)

type apiErrorResponse struct {
	ReturnCode string `json:"returnCode"`
}

type apiResponse struct {
	ReturnCode    string `json:"returnCode"`
	PacketLogInfo []struct {
		HDDServiceCode string `json:"hddServiceCode"`
		Plan           string `json:"plan"`
		HDOInfo        []struct {
			HDOServiceCode string      `json:"hdoServiceCode"`
			PacketLog      []PacketLog `json:"packetLog"`
		} `json:"hdoInfo"`
		HDUInfo []struct {
			HDUServiceCode string      `json:"hduServiceCode"`
			PacketLog      []PacketLog `json:"packetLog"`
		} `json:"hduInfo"`
	} `json:"packetLogInfo"`
}

type PacketLog struct {
	Date          Date `json:"date"`
	WithCoupon    int  `json:"withCoupon"`
	WithoutCoupon int  `json:"withoutCoupon"`
}

type Date string

func (d *Date) isToday() bool {
	return string(*d) == time.Now().In(JST).Format("20060102")
}

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
	res, err := request(developerID, cfg.Token)
	if err != nil {
		return err
	}
	defer res.close(&err)
	return res.Show(cc.App.Writer)
}

func request(developerID, token string) (res *result, err error) {
	var req *http.Request
	req, err = http.NewRequest("GET", APIURL, nil)
	if err != nil {
		return
	}
	req.Header.Add("X-IIJmio-Developer", developerID)
	req.Header.Add("X-IIJmio-Authorization", token)
	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	return &result{resp.StatusCode, resp.Body}, nil
}

type result struct {
	status int
	reader io.ReadCloser
}

func (res *result) close(err *error) { closer(res.reader, err) }

// Show shows the warning if needed
func (res *result) Show(out io.Writer) error {
	if res.status != http.StatusOK {
		var resp apiErrorResponse
		if err := json.NewDecoder(res.reader).Decode(&resp); err != nil {
			return err
		}
		msg := fmt.Sprintf("error: status => %d, message => %s",
			res.status, resp.ReturnCode)
		printResult(out, &msg)
		return nil
	}
	var resp apiResponse
	if err := json.NewDecoder(res.reader).Decode(&resp); err != nil {
		return err
	}
	if len(resp.PacketLogInfo) == 0 {
		return errors.New("no PacketLogInfo")
	}
	if len(resp.PacketLogInfo[0].HDOInfo) == 0 {
		return errors.New("no PacketLogInfo[0].HDOInfo")
	}
	packetLogs := resp.PacketLogInfo[0].HDOInfo[0].PacketLog
	if len(packetLogs) == 0 {
		return errors.New("no PacketLogInfo[0].HDOInfo[0].PacketLog")
	}
	var today *PacketLog
	for _, pl := range packetLogs {
		if pl.Date.isToday() {
			today = &pl
			break
		}
	}
	if today == nil {
		return errors.New("no PacketLog for today is found")
	}
	if today.WithCoupon > THRESHOLD {
		msg := fmt.Sprintf("you use %d MB with the coupon", today.WithCoupon)
		printResult(out, &msg)
	}
	return nil
}

func printResult(out io.Writer, body *string) {
	fmt.Fprintf(out, "iijmio-checker: %s", *body)
}
