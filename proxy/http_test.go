package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	url "net/url"
	"testing"
)

func TestHttpProxy_Request(t *testing.T) {
	proxy := HttpProxy{
		Address: "localhost:10808",
		Ssl:     false,
	}
	targetUrl, _ := url.Parse("https://api64.ipify.org?format=json")
	request := http.Request{
		Method: "GET",
		URL:    targetUrl,
		Header: map[string][]string{
			"User-Agent": {"Mozilla/5.0 (Linux; Android 10; SM-A015AZ) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.116 Mobile Safari/537.36"},
		},
	}
	resp, _ := proxy.Request(&request)
	defer resp.Body.Close()

	type IPDataStruct struct {
		IP string `json:"ip"`
	}

	var data IPDataStruct
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&data); err != nil {
		fmt.Println(err)
	}
	fmt.Println(data)
}
