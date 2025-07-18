package proxy

import (
	"github.com/shadowsocks/go-shadowsocks2/core"
	"github.com/shadowsocks/go-shadowsocks2/socks"
	"log"
	"net"
	"net/http"
	"time"
)

type SSProxy struct {
	Address  string
	Method   string
	Password string
}

func (s *SSProxy) Proxy(clientConn net.Conn, r *http.Request) {
	if r.Method == http.MethodConnect {
		s.connect(clientConn, r)
	} else {
		s.direct(clientConn, r)
	}
}

func (s *SSProxy) direct(clientConn net.Conn, r *http.Request) {
	serverConn, err := s.newSsConn(r, clientConn)
	if err != nil {
		return
	}
	defer serverConn.Close()

	err = httpProxyStartTransfer(r, clientConn, serverConn)
	if err != nil {
		return
	}

	log.Printf("Client <-> Proxy (current) <-> %s (ss) <-> %s (target)", s.Address, r.Host)

	transfer(clientConn, serverConn)
}

func (s *SSProxy) connect(clientConn net.Conn, r *http.Request) {
	connectionEstablished(clientConn)

	serverConn, err := s.newSsConn(r, clientConn)
	if err != nil {
		return
	}
	defer serverConn.Close()

	log.Printf("Client <-> Proxy (current) <-> %s (ss) <-> %s (target)", s.Address, r.Host)

	transfer(clientConn, serverConn)
}

func (s *SSProxy) newSsConn(r *http.Request, conn net.Conn) (net.Conn, error) {
	cipher, err := core.PickCipher(s.Method, nil, s.Password)
	if err != nil {
		badGatewayError(conn)
		return nil, err
	}

	ssConn, err := core.Dial("tcp", s.Address, cipher)
	if err != nil {
		badGatewayError(conn)
		return nil, err
	}

	var host string

	host, port, err := extractHostAndPort(r)
	if err != nil {
		serverError(conn)
		return nil, err
	}

	_, err = ssConn.Write(socks.ParseAddr(host + ":" + port))
	if err != nil {
		badGatewayError(conn)
		return nil, err
	}

	err = ssConn.SetDeadline(time.Now().Add(15 * time.Second))
	if err != nil {
		badGatewayError(conn)
		return nil, err
	}

	return ssConn, nil
}
