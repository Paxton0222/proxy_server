package proxy

import (
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-vmess"
	"github.com/sagernet/sing/common/metadata"
	"log"
	"net"
	"net/http"
)

type VmessProxy struct {
	Address          string
	Port             string
	Uuid             string
	Security         string
	AlterId          int
	TransportType    string
	TransportHideUrl string
	TransportPath    string
	TlsConfig        option.OutboundTLSOptions
}

func (v *VmessProxy) Proxy(clientConn net.Conn, r *http.Request) {
	if r.Method == http.MethodConnect {
		v.connect(clientConn, r)
	} else {
		v.direct(clientConn, r)
	}
}

func (v *VmessProxy) Request(r *http.Request) (*http.Response, error) {
	transportConn, err := newV2RayTransportConn(r, v.Address, v.Port, v.TransportType, v.TransportHideUrl, v.TransportPath, v.TlsConfig)
	if err != nil {
		return nil, err
	}
	defer transportConn.Close()

	serverConn, err := v.newVmessConn(r, transportConn)
	if err != nil {
		return nil, err
	}
	defer serverConn.Close()

	return sendHttpOverTlsRequest(r, serverConn)
}

func (v *VmessProxy) direct(clientConn net.Conn, r *http.Request) {
	transportConn, err := newV2RayTransportConn(r, v.Address, v.Port, v.TransportType, v.TransportHideUrl, v.TransportPath, v.TlsConfig)
	if err != nil {
		badGatewayError(clientConn)
		return
	}
	defer transportConn.Close()

	serverConn, err := v.newVmessConn(r, transportConn)
	if err != nil {
		badGatewayError(clientConn)
		return
	}
	defer serverConn.Close()

	err = httpProxyStartTransfer(r, clientConn, serverConn)
	if err != nil {
		badGatewayError(clientConn)
		return
	}

	log.Printf("Client <-> Proxy (current) <-> %s:%s (vmess) <-> %s (target)", v.Address, v.Port, r.Host)

	transfer(clientConn, serverConn)
}

func (v *VmessProxy) connect(clientConn net.Conn, r *http.Request) {
	connectionEstablished(clientConn)

	transportConn, err := newV2RayTransportConn(r, v.Address, v.Port, v.TransportType, v.TransportHideUrl, v.TransportPath, v.TlsConfig)
	if err != nil {
		badGatewayError(clientConn)
		return
	}
	defer transportConn.Close()

	serverConn, err := v.newVmessConn(r, transportConn)
	if err != nil {
		badGatewayError(clientConn)
		return
	}
	defer serverConn.Close()

	log.Printf("Client <-> Proxy (current) <-> %s:%s (vmess) <-> %s (target)", v.Address, v.Port, r.Host)
	transfer(clientConn, serverConn)
}

func (v *VmessProxy) newVmessConn(r *http.Request, transportConn net.Conn) (net.Conn, error) {
	client, err := vmess.NewClient(
		v.Uuid,
		v.Security,
		v.AlterId,
	)
	if err != nil {
		log.Println("VMess client error:", err)
		return nil, err
	}

	host, port, err := extractHostAndPort(r)
	if err != nil {
		return nil, err
	}

	serverConn, err := client.DialConn(
		transportConn,
		metadata.ParseSocksaddrHostPortStr(host, port),
	)
	if err != nil {
		log.Println("VMess DialConn error:", err)
		return nil, err
	}

	return serverConn, nil
}
