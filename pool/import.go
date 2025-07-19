package pool

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/url"
	"os"
	"proxy/proxy"
	"strconv"
	"strings"

	"github.com/sagernet/sing-box/option"
)

func LoadProxyConfigFromFile(filePath string) []*Node {
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic("讀取 proxy 配置失敗: " + err.Error())
	}
	lines := strings.Split(string(data), "\n")

	nodes, err := ImportFromURILines(lines)
	log.Printf("讀取節點數: %d", len(nodes))
	if err != nil {
		panic("解析 proxy 配置失敗: " + err.Error())
	}
	return nodes
}

func ImportFromURILines(lines []string) ([]*Node, error) {
	nodes := make([]*Node, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch {
		case strings.HasPrefix(line, "vless://"):
			node, err := parseVLESS(line)
			if err == nil {
				nodes = append(nodes, node)
			}

		case strings.HasPrefix(line, "vmess://"):
			node, err := parseVMESS(line)
			if err == nil {
				nodes = append(nodes, node)
			}

		case strings.HasPrefix(line, "ss://"):
			node, err := parseSS(line)
			if err == nil {
				nodes = append(nodes, node)
			}

		case strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://"):
			node, err := parseHTTP(line)
			if err == nil {
				nodes = append(nodes, node)
			}
		}
	}

	return nodes, nil
}

func parseVLESS(uri string) (*Node, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	uuid := u.User.Username()
	query := u.Query()

	host, port, _ := strings.Cut(u.Host, ":")

	return &Node{
		Host: u.Host,
		ProxyServer: &proxy.VlessProxy{
			Address:          host,
			Port:             port,
			Uuid:             uuid,
			Flow:             query.Get("flow"),
			TransportType:    query.Get("type"),
			TransportPath:    query.Get("path"),
			TransportHideUrl: query.Get("host"),
			TlsConfig: option.OutboundTLSOptions{
				Enabled:    query.Get("security") == "tls",
				Insecure:   query.Get("allowInsecure") == "1",
				ServerName: query.Get("sni"),
			},
		},
	}, nil
}

func parseVMESS(uri string) (*Node, error) {
	raw := strings.TrimPrefix(uri, "vmess://")
	raw = strings.TrimSpace(raw)

	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	var conf struct {
		Add  string `json:"add"`
		Port int    `json:"port"`
		ID   string `json:"id"`
		Net  string `json:"net"`
		Path string `json:"path"`
		TLS  string `json:"tls"`
		Aid  int    `json:"aid"`
		Host string `json:"host"`
	}

	if err := json.Unmarshal(decoded, &conf); err != nil {
		return nil, err
	}

	return &Node{
		Host: conf.Add + ":" + strconv.Itoa(conf.Port),
		ProxyServer: &proxy.VmessProxy{
			Address:          conf.Add,
			Port:             strconv.Itoa(conf.Port),
			Uuid:             conf.ID,
			AlterId:          conf.Aid,
			Security:         "auto",
			TransportType:    conf.Net,
			TransportPath:    conf.Path,
			TransportHideUrl: conf.Host,
		},
	}, nil
}

func parseSS(uri string) (*Node, error) {
	uri = strings.TrimPrefix(uri, "ss://")

	// 有些 ss:// URI 可能會帶有插件等資訊，在此先拆掉 query 和 fragment
	if idx := strings.IndexAny(uri, "?#"); idx != -1 {
		uri = uri[:idx]
	}

	parts := strings.SplitN(uri, "@", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid ss format, missing '@'")
	}

	// 先嘗試完整 Base64 解碼（URL safe 也試）
	decodedBytes, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		//log.Println("base64 decode error:", err)
		return nil, err
	}
	decoded := string(decodedBytes)

	// 解析格式 "method:password@host:port"
	atIndex := strings.LastIndex(decoded, ":")
	if atIndex == -1 {
		return nil, errors.New("invalid ss format, missing '@'")
	}

	method := decoded[:atIndex]
	pass := decoded[atIndex+1:]
	addr := parts[1]

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, errors.New("invalid host:port")
	}

	return &Node{
		Host: host + ":" + port,
		ProxyServer: &proxy.SSProxy{
			Address:  host + ":" + port,
			Password: pass,
			Method:   method,
		},
	}, nil
}

func parseHTTP(uri string) (*Node, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	ssl := u.Scheme == "https"

	return &Node{
		Host: u.Host,
		ProxyServer: &proxy.HttpProxy{
			Address: u.Host,
			Ssl:     ssl,
		},
	}, nil
}
