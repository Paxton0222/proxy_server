package proxy

import (
	"github.com/sagernet/sing-box/common/dialer"
	"github.com/sagernet/sing-box/common/tls"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/transport/v2raywebsocket"
	"github.com/sagernet/sing/common/metadata"
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

// 建立傳輸協議
func newV2RayTransportConn(
	r *http.Request,
	clientConn net.Conn,
	address string,
	transportType string,
	transportPath string,
) (net.Conn, error) {
	var transportConn net.Conn
	var err error

	if transportType == "" {
		transportType = "tcp"
	}

	if transportType == "tcp" {
		transportConn, err = net.Dial("tcp", address)
		if err != nil {
			badGatewayError(clientConn)
			log.Println("Tcp dial error:", err)
			return nil, err
		}
	} else if transportType == "ws" {
		dialerOptions := &option.DialerOptions{}
		d, err := dialer.New(r.Context(), *dialerOptions)
		if err != nil {
			badGatewayError(clientConn)
			log.Println("Websocket dial error:", err)
			return nil, err
		}

		websocketOptions := &option.V2RayWebsocketOptions{
			Path: transportPath,
		}
		tlsConfig := &option.OutboundTLSOptions{}
		tlsOption, err := tls.NewClient(r.Context(), address, *tlsConfig)
		if err != nil {
			badGatewayError(clientConn)
			return nil, err
		}

		wsTarget := metadata.ParseSocksaddr(address)

		transport, err := v2raywebsocket.NewClient(r.Context(), d, wsTarget, *websocketOptions, tlsOption)
		if err != nil {
			badGatewayError(clientConn)
			return nil, err
		}

		transportConn, err = transport.DialContext(r.Context())
		if err != nil {
			badGatewayError(clientConn)
			return nil, err
		}
	}
	return transportConn, nil
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
