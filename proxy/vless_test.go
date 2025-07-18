package proxy

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"testing"
)

func TestVlessRequest(t *testing.T) {
	proxy := VlessProxy{
		Address:       "localhost",
		Port:          "10088",
		Uuid:          "60834c02-6962-44d6-b1f3-993452abc1b0",
		Flow:          "",
		TransportType: "tcp",
		TransportPath: "",
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
