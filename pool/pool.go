package pool

import (
	"log"
	"math/rand"
	"net"
	"net/http"
	"proxy/proxy"
	"sort"
	"sync"
	"time"
)

type Pool struct {
	Mu            sync.RWMutex
	ActiveNodes   []*Node
	Nodes         []*Node
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
			IPTarget:        "https://api64.ipify.org?format=json",
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
		resultCh := make(chan *Node)

		go p.refreshActive(checkThreads, resultCh)

		var active []*Node
		for n := range resultCh {
			// 實時收到一個健康節點就加入 active，這裡加鎖更新
			p.Mu.Lock()
			active = append(active, n)
			p.ActiveNodes = active
			p.Mu.Unlock()
		}

		// 排序健康節點列表（也可放在上面）
		p.Mu.Lock()
		sort.Slice(active, func(i, j int) bool { return active[i].Host < active[j].Host })
		p.ActiveNodes = active
		p.Mu.Unlock()

		log.Printf("健康節點更新，健康節點數: %d，不健康節點數: %d", len(active), len(p.Nodes)-len(active))

		time.Sleep(interval)
	}
}

func (p *Pool) Add(node *Node) {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.Nodes = append(p.Nodes, node)
}

func (p *Pool) Remove(host string) {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.removeActiveNode(host)
	p.removeNode(host)
}

func (p *Pool) GetRandom() *Node {
	p.Mu.RLock()
	defer p.Mu.RUnlock()
	if len(p.ActiveNodes) == 0 {
		return nil
	}
	index := p.Random.Intn(len(p.ActiveNodes))
	return p.ActiveNodes[index]
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

func (p *Pool) removeActiveNode(host string) {
	sort.Slice(p.ActiveNodes, func(i, j int) bool {
		return p.ActiveNodes[i].Host < p.ActiveNodes[j].Host
	})

	index := sort.Search(len(p.ActiveNodes), func(i int) bool {
		return p.ActiveNodes[i].Host >= host
	})

	if index < len(p.ActiveNodes) && p.ActiveNodes[index].Host == host {
		p.ActiveNodes = append(p.ActiveNodes[:index], p.ActiveNodes[index+1:]...)
	}
}

func (p *Pool) removeNode(host string) {
	sort.Slice(p.Nodes, func(i, j int) bool {
		return p.Nodes[i].Host < p.Nodes[j].Host
	})

	index := sort.Search(len(p.Nodes), func(i int) bool {
		return p.Nodes[i].Host >= host
	})

	if index < len(p.Nodes) && p.Nodes[index].Host == host {
		p.Nodes = append(p.Nodes[:index], p.Nodes[index+1:]...)
	}
}

func (p *Pool) refreshActive(checkThreads int8, resultCh chan<- *Node) {
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
				log.Printf("節點 %s 健康檢查成功 IP: %s", n.Host, ip)
				resultCh <- n
			} else {
				log.Printf("節點 %s 健康檢查失敗: %v", n.Host, err)
			}
		}(node)
	}

	wg.Wait()
	close(resultCh)
}
