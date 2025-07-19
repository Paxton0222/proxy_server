package pool

import (
	"proxy/proxy"
	"time"
)

type Node struct {
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

func (n *Node) GetProxyServer() proxy.Proxy {
	return n.ProxyServer
}
