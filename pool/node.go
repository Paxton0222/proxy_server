package pool

import (
	"proxy/proxy"
	"time"
)

type ProxyNode struct {
	Host        string
	ProxyServer proxy.Proxy
	NodeInfo    NodeInfo
}

type NodeInfo struct {
	Active    bool
	IP        string
	Latency   time.Duration
	Bandwidth float64
}

func (n *ProxyNode) GetProxyServer() proxy.Proxy {
	return n.ProxyServer
}
