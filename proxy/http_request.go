package proxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"log"
	"net"
	"net/http"
)

func HttpProxyStartTransfer(r *http.Request, clientConn net.Conn, serverConn net.Conn) error {
	err := r.Write(serverConn)
	if err != nil {
		BadGatewayError(clientConn)
		return err
	}
	return err
}

func GetOriginalHttpRawContext(r *http.Request) string {
	var buf bytes.Buffer
	err := r.Write(&buf)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func ExtractHostAndPort(r *http.Request) (string, string, error) {
	var host string
	var port string

	var err error

	if r.URL.Scheme == "http" {
		host, port, err = net.SplitHostPort(r.URL.Host)
		if err != nil {
			host = r.URL.Host
			port = "80"
		}
	} else {
		host, port, err = net.SplitHostPort(r.URL.Host)
		if err != nil {
			host = r.URL.Host
			port = "443"
		}
	}
	return host, port, nil
}

func SendHttpRequest(r *http.Request, conn net.Conn) (*http.Response, error) {
	raw := GetOriginalHttpRawContext(r)
	if _, err := conn.Write([]byte(raw)); err != nil {
		log.Println("發送請求到目標主機失敗:", err)
		return nil, err
	}

	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, r)
	if err != nil {
		log.Println("讀取目標主機回應失敗:", err)
		return nil, err
	}
	return resp, nil
}

func SendHttpOverTlsRequest(r *http.Request, proxyConn net.Conn) (*http.Response, error) {
	var resp *http.Response
	var err error
	if r.URL.Scheme == "http" {
		resp, err = SendHttpRequest(r, proxyConn)
		if err != nil {
			return nil, err
		}
	} else {
		serverConn := tls.Client(proxyConn, &tls.Config{
			ServerName: r.URL.Hostname(),
		})
		if err := serverConn.Handshake(); err != nil {
			//log.Println("與目標主機握手失敗:", err)
			return nil, err
		}
		defer serverConn.Close()
		resp, err = SendHttpRequest(r, serverConn)
	}
	return resp, err
}
