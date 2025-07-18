package proxy

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"testing"
)

func TestVmessRequest(t *testing.T) {
	proxy := VmessProxy{
		Address:          "35.212.178.40",
		Port:             "8880",
		Uuid:             "482c7152-b91b-4081-b1fc-5a0cf13c6635",
		AlterId:          0,
		Security:         "auto",
		TransportType:    "ws",
		TransportHideUrl: "",
		TransportPath:    "482c7152-b91b-4081-b1fc-5a0cf13c6635-vm",
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
