package proxy

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"testing"
)

func TestSSRequest(t *testing.T) {
	proxy := SSProxy{
		Address:  "localhost:8388",
		Method:   "aes-256-gcm",
		Password: "1234",
	}

	targetUrl, _ := url.Parse("https://api64.ipify.org?format=json")
	request := http.Request{
		Method: "GET",
		URL:    targetUrl,
		Header: map[string][]string{
			"User-Agent": {"Mozilla/5.0 (Linux; Android 10; SM-A015AZ) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.116 Mobile Safari/537.36"},
		},
	}
	resp, err := proxy.Request(&request)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	type IPDataStruct struct {
		IP string `json:"ip"`
	}

	var data IPDataStruct
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&data); err != nil {
		log.Println(err)
	}
	fmt.Println(data)
}
