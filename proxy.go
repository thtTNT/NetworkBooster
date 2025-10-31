package main

import (
	"fmt"
	"io"
	"log"
	"net"
	
	"NetworkBooster/shadowsocks"
	"NetworkBooster/socks5"
)

type Proxy struct {
	socks5Server  *socks5.Server
	ssClient      *shadowsocks.Client
}

func NewProxy(socks5ListenAddr, ssServerAddr, ssMethod, ssPassword string) (*Proxy, error) {
	// 创建 SOCKS5 服务器（监听本地）
	socks5Server, err := socks5.NewServer(socks5ListenAddr)
	if err != nil {
		return nil, err
	}
	
	// 创建 Shadowsocks 客户端（连接到远程服务器）
	ssClient, err := shadowsocks.NewClient(ssServerAddr, ssMethod, ssPassword)
	if err != nil {
		socks5Server.Close()
		return nil, err
	}
	
	return &Proxy{
		socks5Server: socks5Server,
		ssClient:     ssClient,
	}, nil
}

// VerifySSServer 验证 Shadowsocks 服务器可用性
func (p *Proxy) VerifySSServer() error {
	log.Printf("正在验证 Shadowsocks 服务器连接...")
	if err := p.ssClient.TestConnection(); err != nil {
		return fmt.Errorf("Shadowsocks 服务器验证失败: %v", err)
	}
	log.Printf("Shadowsocks 服务器连接验证成功")
	return nil
}

func (p *Proxy) Run() error {
	log.Printf("SOCKS5 server listening on %s", p.socks5Server.Addr())
	log.Printf("Forwarding to Shadowsocks server")
	
	for {
		conn, err := p.socks5Server.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		
		go p.handleConnection(conn)
	}
}

func (p *Proxy) Close() error {
	return p.socks5Server.Close()
}

func (p *Proxy) handleConnection(conn net.Conn) {
	defer conn.Close()
	
	// 处理 SOCKS5 连接
	req, err := p.socks5Server.HandleConnection(conn)
	if err != nil {
		log.Printf("Handle SOCKS5 connection error: %v", err)
		return
	}
	
	log.Printf("Connecting to %s via Shadowsocks", req.String())
	
	// 通过 Shadowsocks 连接到目标
	ssConn, err := p.ssClient.Dial(req.Address, req.Port)
	if err != nil {
		log.Printf("Shadowsocks dial error: %v", err)
		p.socks5Server.SendSuccessReply(conn)
		return
	}
	defer ssConn.Close()
	
	// 发送成功响应
	if err := p.socks5Server.SendSuccessReply(conn); err != nil {
		log.Printf("Send reply error: %v", err)
		return
	}
	
	// 双向转发数据
	go io.Copy(ssConn, conn)
	io.Copy(conn, ssConn)
}
