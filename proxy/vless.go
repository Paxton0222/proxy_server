package proxy

import (
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-vmess/vless"
	"github.com/sagernet/sing/common/logger"
	"github.com/sagernet/sing/common/metadata"
	"log"
	"net"
	"net/http"
)

type VlessProxy struct {
	Address          string
	Port             string
	Uuid             string
	Flow             string
	TransportType    string
	TransportHideUrl string
	TransportPath    string
	TlsConfig        option.OutboundTLSOptions
}

func (v *VlessProxy) Proxy(clientConn net.Conn, r *http.Request) {
	if r.Method == http.MethodConnect {
		v.connect(clientConn, r)
	} else {
		v.direct(clientConn, r)
	}
}

func (v *VlessProxy) Request(r *http.Request) (*http.Response, error) {
	transportConn, err := newV2RayTransportConn(r, v.Address, v.Port, v.TransportType, v.TransportHideUrl, v.TransportPath, v.TlsConfig)
	if err != nil {
		return nil, err
	}
	defer transportConn.Close()

	serverConn, err := v.newVlessConn(r, transportConn)
	if err != nil {
		return nil, err
	}
	defer serverConn.Close()

	return sendHttpOverTlsRequest(r, serverConn)
}

func (v *VlessProxy) direct(clientConn net.Conn, r *http.Request) {
	transportConn, err := newV2RayTransportConn(r, v.Address, v.Port, v.TransportType, v.TransportHideUrl, v.TransportPath, v.TlsConfig)
	if err != nil {
		badGatewayError(clientConn)
		return
	}
	defer transportConn.Close()

	serverConn, err := v.newVlessConn(r, transportConn)
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

	log.Printf("Client <-> ProxyServer (current) <-> %s:%s (vless) <-> %s (target)", v.Address, v.Port, r.Host)
	transfer(clientConn, serverConn)
}

func (v *VlessProxy) connect(clientConn net.Conn, r *http.Request) {
	connectionEstablished(clientConn)

	transportConn, err := newV2RayTransportConn(r, v.Address, v.Port, v.TransportType, v.TransportHideUrl, v.TransportPath, v.TlsConfig)
	if err != nil {
		badGatewayError(clientConn)
		return
	}
	defer transportConn.Close()

	serverConn, err := v.newVlessConn(r, transportConn)
	if err != nil {
		badGatewayError(clientConn)
		return
	}
	defer serverConn.Close()

	log.Printf("Client <-> ProxyServer (current) <-> %s:%s (vless) <-> %s (target)", v.Address, v.Port, r.Host)
	transfer(clientConn, serverConn)
}

func (v *VlessProxy) newVlessConn(r *http.Request, transportConn net.Conn) (net.Conn, error) {
	client, err := vless.NewClient(
		v.Uuid,
		v.Flow,
		logger.NOP(),
	)
	if err != nil {
		log.Println("Vless client error:", err)
		return nil, err
	}

	host, port, err := extractHostAndPort(r)
	if err != nil {
		return nil, err
	}

	vlessConn, err := client.DialConn(
		transportConn,
		metadata.ParseSocksaddrHostPortStr(host, port),
	)
	if err != nil {
		log.Println("Vless DialConn error:", err)
		return nil, err
	}

	return vlessConn, nil
}
