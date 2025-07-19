package pool

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"proxy/proxy"
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
		Port string `json:"port"`
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
		Host: conf.Add + ":" + conf.Port,
		ProxyServer: &proxy.VmessProxy{
			Address:          conf.Add,
			Port:             conf.Port,
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

	parts := strings.SplitN(uri, "@", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid ss format")
	}

	decoded, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		decoded, err = base64.StdEncoding.DecodeString(parts[0])
		if err != nil {
			return nil, err
		}
	}

	methodPass := string(decoded)
	methodPassParts := strings.SplitN(methodPass, ":", 2)
	if len(methodPassParts) != 2 {
		return nil, errors.New("invalid ss credentials")
	}

	host, port, _ := strings.Cut(parts[1], ":")

	return &Node{
		Host: host + ":" + port,
		ProxyServer: &proxy.SSProxy{
			Address:  host,
			Password: methodPassParts[1],
			Method:   methodPassParts[0],
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
