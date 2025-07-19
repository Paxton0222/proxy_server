package proxy

import (
	"io"
	"net"
)

// Transfer 建立連線後，讓兩邊交換資料
func Transfer(clientConn net.Conn, serverConn net.Conn) {
	go io.Copy(serverConn, clientConn)
	io.Copy(clientConn, serverConn)
}

func ConnectionEstablished(conn net.Conn) {
	_, err := conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		return
	}
}

func BadGatewayError(conn net.Conn) {
	_, err := conn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
	if err != nil {
		return
	}
}

func ServerError(conn net.Conn) {
	_, err := conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
	if err != nil {
		return
	}
}
