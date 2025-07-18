package proxy

import (
	"io"
	"log"
	"net"
	"net/http"
)

func extractHostAndPort(r *http.Request) (string, string, error) {
	var host string
	var port string

	var err error

	if r.URL.Scheme == "http" {
		host, port, err = net.SplitHostPort(r.Host)
		if err != nil {
			host = r.Host
			port = "80"
		}
	} else {
		host, port, err = net.SplitHostPort(r.Host)
		if err != nil {
			log.Println("SplitHostPort error:", err)
			return "", "", err
		}
	}
	return host, port, nil
}

// 建立連線後，讓兩邊交換資料
func transfer(clientConn net.Conn, serverConn net.Conn) {
	go io.Copy(serverConn, clientConn)
	io.Copy(clientConn, serverConn)
}

func httpProxyStartTransfer(r *http.Request, clientConn net.Conn, serverConn net.Conn) error {
	err := r.Write(serverConn)
	if err != nil {
		badGatewayError(clientConn)
		return err
	}
	return err
}

func connectionEstablished(conn net.Conn) {
	_, err := conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		return
	}
}

func badGatewayError(conn net.Conn) {
	_, err := conn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
	if err != nil {
		return
	}
}

func serverError(conn net.Conn) {
	_, err := conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
	if err != nil {
		return
	}
}
