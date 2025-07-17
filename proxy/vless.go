package proxy

import (
	"github.com/sagernet/sing-vmess/vless"
	"github.com/sagernet/sing/common/logger"
	"github.com/sagernet/sing/common/metadata"
	"log"
	"net"
	"net/http"
)

type VlessProxy struct {
	Address       string
	Uuid          string
	Flow          string
	TransportType string
	TransportPath string
}

func (v *VlessProxy) Proxy(clientConn net.Conn, r *http.Request) {
	if r.Method == http.MethodConnect {
		v.connect(clientConn, r)
	} else {
		v.direct(clientConn, r)
	}
}

func (v *VlessProxy) direct(clientConn net.Conn, r *http.Request) {
	transportConn, err := newV2RayTransportConn(r, clientConn, v.Address, v.TransportType, v.TransportPath)
	if err != nil {
		return
	}
	defer transportConn.Close()

	serverConn, err := v.newVlessConn(r, clientConn, transportConn)
	if err != nil {
		return
	}
	defer serverConn.Close()

	err = httpProxyStartTransfer(r, clientConn, serverConn)
	if err != nil {
		badGatewayError(clientConn)
		return
	}

	log.Printf("Client -> Proxy (current) -> %s (vless) -> %s (target)", v.Address, r.Host)

	transfer(clientConn, serverConn)
}

func (v *VlessProxy) connect(clientConn net.Conn, r *http.Request) {
	connectionEstablished(clientConn)

	transportConn, err := newV2RayTransportConn(r, clientConn, v.Address, v.TransportType, v.TransportPath)
	if err != nil {
		return
	}
	defer transportConn.Close()

	serverConn, err := v.newVlessConn(r, clientConn, transportConn)
	if err != nil {
		return
	}
	defer serverConn.Close()

	log.Printf("Client -> Proxy (current) -> %s (vless) -> %s (target)", v.Address, r.Host)

	transfer(clientConn, serverConn)
}

func (v *VlessProxy) newVlessConn(r *http.Request, clientConn net.Conn, transportConn net.Conn) (net.Conn, error) {
	client, err := vless.NewClient(
		v.Uuid,
		v.Flow,
		logger.NOP(),
	)
	if err != nil {
		log.Println("Vless client error:", err)
		badGatewayError(clientConn)
		return nil, err
	}

	host, port, err := extractHostAndPort(r)
	if err != nil {
		serverError(clientConn)
		return nil, err
	}

	vlessConn, err := client.DialConn(
		transportConn,
		metadata.ParseSocksaddrHostPortStr(host, port),
	)
	if err != nil {
		log.Println("Vless DialConn error:", err)
		badGatewayError(clientConn)
		return nil, err
	}

	return vlessConn, nil
}
