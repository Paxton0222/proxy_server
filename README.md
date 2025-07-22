# 代理池

## 支援協議
1. [x] http
2. [x] https
3. [x] shadowsocks
4. [x] vless
5. [x] vmess
6. [ ] trojan
7. [ ] socks5
8. [ ] hysteria2

## proxy 檔案設定

proxy.txt 支持標準 uri 導入

## 部署方式

```shell
cp ./proxy_example.txt proxy.txt
cp. ./env.example .env
docker-compose up -d
```