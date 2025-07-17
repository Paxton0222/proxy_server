package proxy

import (
	"github.com/sagernet/sing-vmess"
	"github.com/sagernet/sing/common/metadata"
	"log"
	"net"
	"net/http"
)

type VmessProxy struct {
	Address       string
	Uuid          string
	Security      string
	AlterId       int
	TransportType string
	TransportPath string
}

func (v *VmessProxy) Proxy(clientConn net.Conn, r *http.Request) {
	if r.Method == http.MethodConnect {
		v.connect(clientConn, r)
	} else {
		v.direct(clientConn, r)
	}
}

func (v *VmessProxy) direct(clientConn net.Conn, r *http.Request) {
	transportConn, err := newV2RayTransportConn(r, clientConn, v.Address, v.TransportType, v.TransportPath)
	if err != nil {
		return
	}
	defer transportConn.Close()

	serverConn, err := v.newVmessConn(r, clientConn, transportConn)
	if err != nil {
		return
	}
	defer serverConn.Close()

	err = httpProxyStartTransfer(r, clientConn, serverConn)
	if err != nil {
		badGatewayError(clientConn)
		return
	}

	log.Printf("Client -> Proxy (current) -> %s (vmess) -> %s (target)", v.Address, r.Host)

	transfer(clientConn, serverConn)
}

func (v *VmessProxy) connect(clientConn net.Conn, r *http.Request) {
	connectionEstablished(clientConn)

	transportConn, err := newV2RayTransportConn(r, clientConn, v.Address, v.TransportType, v.TransportPath)
	if err != nil {
		return
	}
	defer transportConn.Close()

	serverConn, err := v.newVmessConn(r, clientConn, transportConn)
	if err != nil {
		return
	}
	defer serverConn.Close()

	log.Printf("Client -> Proxy (current) -> %s (vmess) -> %s (target)", v.Address, r.Host)
	transfer(clientConn, serverConn)
}

func (v *VmessProxy) newVmessConn(r *http.Request, clientConn net.Conn, transportConn net.Conn) (net.Conn, error) {
	client, err := vmess.NewClient(
		v.Uuid,
		v.Security,
		v.AlterId,
	)
	if err != nil {
		badGatewayError(clientConn)
		log.Println("VMess client error:", err)
		return nil, err
	}

	host, port, err := extractHostAndPort(r)
	if err != nil {
		serverError(clientConn)
		return nil, err
	}

	serverConn, err := client.DialConn(
		transportConn,
		metadata.ParseSocksaddrHostPortStr(host, port),
	)
	if err != nil {
		log.Println("VMess DialConn error:", err)
		badGatewayError(clientConn)
		return nil, err
	}

	return serverConn, nil
}
