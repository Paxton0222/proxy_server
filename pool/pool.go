package pool

import (
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"
)

type Pool struct {
	Mu            sync.RWMutex
	ActiveNodes   []*ProxyNode
	Nodes         []*ProxyNode
	HealthChecker HealthChecker
	Random        *rand.Rand
}

func NewPool(
	nodes []*ProxyNode,
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
		p.Direct(w, r)
	}
}

func (p *Pool) StartHealthCheck(interval time.Duration, checkThreads int8) {
	p.refreshActive(checkThreads)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		p.refreshActive(checkThreads)
	}
}

func (p *Pool) Add(node *ProxyNode) {
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

func (p *Pool) GetRandom() *ProxyNode {
	p.Mu.RLock()
	defer p.Mu.RUnlock()
	if len(p.ActiveNodes) == 0 {
		return nil
	}
	index := p.Random.Intn(len(p.ActiveNodes))
	return p.ActiveNodes[index]
}

func (p *Pool) Direct(w http.ResponseWriter, r *http.Request) {
	if r.URL.Scheme == "http" {
		resp, err := http.DefaultTransport.RoundTrip(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			for _, val := range v {
				w.Header().Add(k, val)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	} else {
		destConn, err := net.Dial("tcp", r.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		defer destConn.Close()
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
			return
		}
		clientConn, _, err := hijacker.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		defer clientConn.Close()
		_, _ = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
		go io.Copy(destConn, clientConn)
		go io.Copy(clientConn, destConn)
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

func (p *Pool) refreshActive(checkThreads int8) {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	log.Println("開始進行健康檢查，共有節點：", len(p.Nodes))

	var wg sync.WaitGroup
	activeCh := make(chan *ProxyNode, len(p.Nodes))
	sem := make(chan struct{}, checkThreads)

	for _, node := range p.Nodes {
		wg.Add(1)
		go func(n *ProxyNode) {
			defer wg.Done()
			// 先進入 semaphore（若已滿會等待）
			sem <- struct{}{}
			defer func() { <-sem }() // 執行完畢後釋放一個 slot

			if ip, err := p.HealthChecker.Ip(n.ProxyServer); err == nil {
				activeCh <- n
				n.NodeInfo.IP = ip // 更新 IP
				log.Printf("節點 %s 健康檢查成功 IP: %s", n.Host, ip)
			} else {
				log.Printf("節點 %s 健康檢查失敗: %v", n.Host, err)
			}
		}(node)
	}
	wg.Wait()
	close(activeCh)

	var active []*ProxyNode
	for n := range activeCh {
		active = append(active, n)
	}

	sort.Slice(p.Nodes, func(i, j int) bool { return p.Nodes[i].Host < p.Nodes[j].Host })
	sort.Slice(active, func(i, j int) bool { return active[i].Host < active[j].Host })

	p.ActiveNodes = active
	log.Printf("健康節點更新，健康節點數: %d，不健康節點數: %d", len(active), len(p.Nodes)-len(active))
}
