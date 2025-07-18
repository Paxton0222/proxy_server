package proxy

import (
	"io"
	"net"
)

// 建立連線後，讓兩邊交換資料
func transfer(clientConn net.Conn, serverConn net.Conn) {
	go io.Copy(serverConn, clientConn)
	io.Copy(clientConn, serverConn)
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
