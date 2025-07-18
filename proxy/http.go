package proxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type HttpProxy struct {
	Address string
	Ssl     bool
}

func (p *HttpProxy) Proxy(clientConn net.Conn, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.direct(clientConn, r)
	} else {
		p.connect(clientConn, r)
	}
}

func (p *HttpProxy) Request(r *http.Request) (*http.Response, error) {
	proxyConn, err := p.newHttpConn(r)
	if err != nil {
		return nil, err
	}
	defer proxyConn.Close()

	serverConn := tls.Client(proxyConn, &tls.Config{
		ServerName: r.URL.Hostname(),
	})
	if err := serverConn.Handshake(); err != nil {
		fmt.Println("與目標主機握手失敗:", err)
		return nil, err
	}
	defer serverConn.Close()

	return sendHttpRequest(r, serverConn)
}

func (p *HttpProxy) newHttpConn(r *http.Request) (net.Conn, error) {
	// 連接 HTTPS Proxy 伺服器
	var proxyConn net.Conn
	var err error
	if p.Ssl {
		proxyConn, err = tls.Dial("tcp", p.Address, &tls.Config{
			InsecureSkipVerify: true, // 可改成嚴格驗證
		})
		if err != nil {
			return nil, err
		}
	} else {
		proxyConn, err = net.Dial("tcp", p.Address)
		if err != nil {
			return nil, err
		}
	}

	target := r.URL.Host
	if !strings.Contains(target, ":") {
		target += ":443"
	}

	// 建立 CONNECT 請求
	connectReq := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Opaque: target},
		Host:   target,
		Header: make(http.Header),
	}
	connectReq.Header.Set("Proxy-Connection", "keep-alive")

	var buf bytes.Buffer
	if err := connectReq.Write(&buf); err != nil {
		fmt.Println("構建 CONNECT 請求失敗:", err)
		return nil, err
	}
	if _, err := proxyConn.Write(buf.Bytes()); err != nil {
		fmt.Println("傳送 CONNECT 請求失敗:", err)
		return nil, err
	}

	// 讀取 CONNECT 回應
	resp, err := http.ReadResponse(bufio.NewReader(proxyConn), r)
	if err != nil {
		fmt.Println("讀取 CONNECT 回應失敗:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("CONNECT 被拒絕:", resp.Status)
		return nil, nil
	}
	return proxyConn, err
}

func (p *HttpProxy) direct(clientConn net.Conn, r *http.Request) {
	serverConn, err := p.newHttpConn(r)
	if err != nil {
		badGatewayError(clientConn)
		return
	}
	defer serverConn.Close()

	err = httpProxyStartTransfer(r, clientConn, serverConn)
	if err != nil {
		return
	}

	log.Printf("Client <-> Proxy (current) <-> %s (http) <-> %s (target)", p.Address, r.Host)
	transfer(clientConn, serverConn)
}

func (p *HttpProxy) connect(clientConn net.Conn, r *http.Request) {
	connectionEstablished(clientConn)

	serverConn, err := p.newHttpConn(r)
	if err != nil {
		badGatewayError(clientConn)
		return
	}
	defer serverConn.Close()

	log.Printf("Client <-> Proxy (current) <-> %s (https) <-> %s (target)", p.Address, r.Host)
	transfer(clientConn, serverConn)
}
