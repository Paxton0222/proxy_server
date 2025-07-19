package pool

import (
	"log"
	"math/rand"
	"net"
	"net/http"
	"proxy/proxy"
	"reflect"
	"sync"
	"time"
)

type Pool struct {
	Mu            sync.RWMutex
	Nodes         []*Node
	Groups        map[string][]*Node // IP -> []*Node
	HealthChecker HealthChecker
	Random        *rand.Rand
}

func NewPool(
	nodes []*Node,
) *Pool {
	p := &Pool{
		Random: rand.New(rand.NewSource(time.Now().UnixNano())),
		Nodes:  nodes,
		HealthChecker: HealthChecker{
			IPTarget: "https://api.ipify.org?format=json",
			//IPTarget:        "https://lgxhuyzqm0pci0mof9hc75zpz3e7f0c0.edns.ip-api.com/json",
			BandWidthTarget: "http://speedtest.tyo11.jp.leaseweb.net/10mb.bin",
		},
	}
	return p
}

func (p *Pool) Handle(w http.ResponseWriter, r *http.Request) {
	clientConn, err, done := p.newHijackClientConn(w)
	if done || err != nil {
		return
	}
	defer clientConn.Close()

	selected := p.GetRandom()
	if selected != nil {
		selected.GetProxyServer().Proxy(clientConn, r)
	} else {
		p.Direct(clientConn, r)
	}
}

func (p *Pool) StartHealthCheck(interval time.Duration, checkThreads int8) {
	for {
		newGroups := make(map[string][]*Node)
		var wg sync.WaitGroup
		sem := make(chan struct{}, checkThreads)

		for _, node := range p.Nodes {
			wg.Add(1)
			go func(n *Node) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				if ip, err := p.HealthChecker.Ip(n.ProxyServer); err == nil {
					n.NodeInfo.IP = ip
					proxyName := reflect.TypeOf(n.GetProxyServer())
					log.Printf("節點 %s (%s) 健康檢查成功 IP: %s", n.Host, proxyName, ip)
					p.Mu.Lock()
					newGroups[ip] = append(newGroups[ip], n)
					p.Mu.Unlock()
				} else {
					log.Printf("節點 %s 健康檢查失敗: %v", n.Host, err)
				}
			}(node)
		}

		wg.Wait()

		p.Mu.Lock()
		p.Groups = newGroups
		p.Mu.Unlock()

		log.Printf("健康節點分組完成，Group 數量: %d", len(p.Groups))
		for ip, nodes := range p.Groups {
			log.Printf("Group IP: %s，節點數: %d", ip, len(nodes))
			//for _, node := range nodes {
			//	log.Printf("  - Host: %s, IP: %s", node.Host, ip)
			//}
		}
		time.Sleep(interval)
	}
}

func (p *Pool) GetRandom() *Node {
	p.Mu.RLock()
	defer p.Mu.RUnlock()

	if len(p.Groups) == 0 {
		return nil
	}

	groupIPs := make([]string, 0, len(p.Groups))
	for ip, nodes := range p.Groups {
		if len(nodes) > 0 {
			groupIPs = append(groupIPs, ip)
		}
	}
	if len(groupIPs) == 0 {
		return nil
	}

	randomGroup := groupIPs[p.Random.Intn(len(groupIPs))]
	//log.Printf("Use IP: %s", randomGroup)
	nodes := p.Groups[randomGroup]
	return nodes[p.Random.Intn(len(nodes))]
}

func (p *Pool) Direct(clientConn net.Conn, r *http.Request) {
	if r.Method == "CONNECT" {
		destConn, err := net.Dial("tcp", r.Host)
		if err != nil {
			proxy.BadGatewayError(clientConn)
			return
		}
		proxy.ConnectionEstablished(clientConn)
		log.Printf(
			"Client <-> ProxyServer (current - https) <-> %s (target)",
			r.Host,
		)
		proxy.Transfer(clientConn, destConn)
	} else {
		host, port, err := proxy.ExtractHostAndPort(r)
		if err != nil {
			proxy.BadGatewayError(clientConn)
			return
		}
		destConn, err := net.Dial("tcp", net.JoinHostPort(host, port))
		if err != nil {
			proxy.BadGatewayError(clientConn)
			return
		}
		defer destConn.Close()

		err = proxy.HttpProxyStartTransfer(r, clientConn, destConn)
		if err != nil {
			return
		}

		log.Printf(
			"Client <-> ProxyServer (current - http) <-> %s (target)",
			r.Host,
		)
		proxy.Transfer(clientConn, destConn)
	}
}

// 劫持客戶端連線
func (p *Pool) newHijackClientConn(w http.ResponseWriter) (net.Conn, error, bool) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return nil, nil, true
	}
	clientConn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, "Hijack failed: "+err.Error(), http.StatusInternalServerError)
		return nil, nil, true
	}
	return clientConn, err, false
}
