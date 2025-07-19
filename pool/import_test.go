package pool

import (
	"fmt"
	"testing"
)

func TestImportFromURILines(t *testing.T) {
	testLines := []string{
		"vless://60834c02-6962-44d6-b1f3-993452abc1b0@localhost:10086?type=ws&host=example.com&path=%2Fvless&security=tls&flow=&sni=example.com&allowInsecure=1",
		"vmess://eyJhZGQiOiJsb2NhbGhvc3QiLCJwb3J0IjoiMTAwODkiLCJpZCI6IjYwODM0YzAyLTY5NjItNDRkNi1iMWYzLTk5MzQ1MmFiYzFiMCIsImFpZCI6MCwibmV0Ijoid3MiLCJwYXRoIjoiL3ZtZXNzIiwidGxzIjoidGxzIiwiaG9zdCI6ImV4YW1wbGUuY29tIn0=",
		"ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@localhost:8388",
		"http://localhost:3128",
		"https://localhost:3129",
	}

	nodes, err := ImportFromURILines(testLines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 5 {
		t.Errorf("expected 5 nodes, got %d", len(nodes))
	}

	types := []string{"*proxy.VlessProxy", "*proxy.VmessProxy", "*proxy.SSProxy", "*proxy.HttpProxy", "*proxy.HttpProxy"}
	for i, node := range nodes {
		actualType := fmt.Sprintf("%T", node.ProxyServer)
		if actualType != types[i] {
			t.Errorf("node %d: expected type %s, got %s", i, types[i], actualType)
		}
	}
}
