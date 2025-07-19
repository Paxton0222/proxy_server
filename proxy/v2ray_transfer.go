package proxy

import (
	"fmt"
	"github.com/sagernet/sing-box/common/dialer"
	"github.com/sagernet/sing-box/common/tls"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/transport/v2raywebsocket"
	"github.com/sagernet/sing/common/json/badoption"
	"github.com/sagernet/sing/common/metadata"
	"log"
	"net"
	"net/http"
)

// 建立 v2ray 傳輸協議
func newV2RayTransportConn(r *http.Request, host string, port string, transportType string, transportHideUrl string, transportPath string, tlsOptions option.OutboundTLSOptions) (net.Conn, error) {
	var transportConn net.Conn
	var err error

	if transportType == "" {
		transportType = "tcp"
	}

	if transportType == "tcp" {
		transportConn, err = net.Dial("tcp", host+":"+port)
		if err != nil {
			log.Println("Tcp dial error:", err)
			return nil, err
		}
	} else if transportType == "ws" {
		dialerOptions := &option.DialerOptions{}
		d, err := dialer.New(r.Context(), *dialerOptions)
		if err != nil {
			log.Println("Websocket dial error:", err)
			return nil, err
		}

		// 處理偽裝域名
		var wsHost string
		if transportHideUrl != "" {
			wsHost = transportHideUrl
		} else {
			wsHost = host
		}

		wsTarget := metadata.ParseSocksaddr(wsHost + ":" + port)

		websocketOptions := &option.V2RayWebsocketOptions{
			Path: transportPath,
			Headers: badoption.HTTPHeader{
				"Host": {wsHost},
			},
		}

		tlsOption, err := tls.NewClient(
			r.Context(),
			wsHost,
			*&tlsOptions,
		)
		if err != nil {
			return nil, err
		}

		transport, err := v2raywebsocket.NewClient(r.Context(), d, wsTarget, *websocketOptions, tlsOption)
		if err != nil {
			return nil, err
		}

		transportConn, err = transport.DialContext(r.Context())
		if err != nil {
			return nil, err
		}
	}
	if transportConn == nil {
		return nil, fmt.Errorf("newV2RayTransportConn: 建立連線失敗（conn 為 nil）")
	}

	return transportConn, nil
}
