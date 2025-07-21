package pool

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"proxy/proxy"
	"time"
)

type HealthChecker struct {
	IPTarget        string
	BandWidthTarget string
}

type IPDataStruct struct {
	IP string `json:"ip"`
}

func (h *HealthChecker) Latency(addr net.Addr) (time.Duration, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr.String(), 10*time.Second)
	if err != nil {
		return -1, err
	}
	conn.Close()
	latency := time.Since(start)
	return latency, nil
}

func (h *HealthChecker) Ip(proxy proxy.Proxy) (string, error) {
	targetUrl, _ := url.Parse(h.IPTarget)
	request := http.Request{
		Method: "GET",
		URL:    targetUrl,
		Header: map[string][]string{
			"User-Agent": {"Mozilla/5.0 (Linux; Android 10; SM-A015AZ) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.116 Mobile Safari/537.36"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan struct{})
	var resp *http.Response
	var err error

	go func() {
		resp, err = proxy.Request(&request)
		close(done)
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-done:
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		var data IPDataStruct
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&data); err != nil {
			log.Println("IP check json decode error:", err, proxy)
			return "", err
		}
		return data.IP, nil
	}
}

func (h *HealthChecker) Bandwidth(proxy proxy.Proxy) (float64, error) {
	const MaxDownloadSize = 10 * 1024 * 1024
	testUrl, _ := url.Parse(h.BandWidthTarget)
	request := http.Request{
		Method: "GET",
		URL:    testUrl,
		Host:   testUrl.Hostname(),
		Header: map[string][]string{
			"User-Agent": {"speed-test"},
		},
	}

	start := time.Now()
	resp, err := proxy.Request(&request)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	// 讀取整個 Response Body
	totalBytes := 0
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		totalBytes += n
		if totalBytes >= MaxDownloadSize {
			break
		}
		if err != nil {
			break
		}
	}
	elapsed := time.Since(start).Seconds()

	if elapsed == 0 {
		return -1, nil
	}
	speedMbps := (float64(totalBytes) * 8) / (elapsed * 1024 * 1024) // Mbps
	return speedMbps, nil
}
