package proxy

import (
	"log"
	"net"
	"net/http"
)

type HttpProxy struct {
	Address string
}

func (p *HttpProxy) Proxy(clientConn net.Conn, r *http.Request) {
	if r.URL.Scheme == "http" {
		p.direct(clientConn, r)
	} else {
		p.connect(clientConn, r)
	}
}

func (p *HttpProxy) newHttpConn(r *http.Request, conn net.Conn) (net.Conn, error) {
	host, port, err := extractHostAndPort(r)
	if err != nil {
		serverError(conn)
		return nil, err
	}

	serverConn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		badGatewayError(conn)
		return nil, err
	}
	return serverConn, nil
}

func (p *HttpProxy) newHttpsConn(r *http.Request, conn net.Conn) (net.Conn, error) {
	host, port, err := extractHostAndPort(r)
	if err != nil {
		serverError(conn)
		return nil, err
	}

	serverConn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		badGatewayError(conn)
		return nil, err
	}
	return serverConn, nil
}

func (p *HttpProxy) direct(clientConn net.Conn, r *http.Request) {
	serverConn, err := p.newHttpConn(r, clientConn)
	if err != nil {
		return
	}
	defer serverConn.Close()

	err = httpProxyStartTransfer(r, clientConn, serverConn)
	if err != nil {
		return
	}

	log.Printf("Client -> Proxy (current) -> %s (http) -> %s (target)", p.Address, r.Host)
	transfer(clientConn, serverConn)
}

func (p *HttpProxy) connect(clientConn net.Conn, r *http.Request) {
	connectionEstablished(clientConn)

	serverConn, err := p.newHttpsConn(r, clientConn)
	if err != nil {
		return
	}
	defer serverConn.Close()

	log.Printf("Client -> Proxy (current) -> %s (https) -> %s (target)", p.Address, r.Host)
	transfer(clientConn, serverConn)
}
