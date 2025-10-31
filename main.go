package main

import (
	"flag"
	"log"
)

func main() {
	// 命令行参数
	socks5Addr := flag.String("socks5", "127.0.0.1:1080", "SOCKS5 监听地址")
	ssServerAddr := flag.String("ss-server", "", "Shadowsocks 服务器地址 (例如: example.com:8388)")
	ssMethod := flag.String("ss-method", "aes-256-gcm", "Shadowsocks 加密方法 (aes-256-gcm 或 chacha20-poly1305)")
	ssPassword := flag.String("ss-password", "", "Shadowsocks 密码")
	flag.Parse()
	
	if *ssServerAddr == "" || *ssPassword == "" {
		log.Fatal("请指定 Shadowsocks 服务器地址和密码")
	}
	
	// 创建代理
	proxy, err := NewProxy(*socks5Addr, *ssServerAddr, *ssMethod, *ssPassword)
	if err != nil {
		log.Fatalf("创建代理失败: %v", err)
	}
	defer proxy.Close()
	
	// 验证 Shadowsocks 服务器
	if err := proxy.VerifySSServer(); err != nil {
		log.Fatalf("Shadowsocks 服务器验证失败: %v", err)
	}
	
	// 运行代理
	log.Fatal(proxy.Run())
}


