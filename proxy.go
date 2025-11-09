package main

import (
	"fmt"
	"io"
	"log"
	"net"
	
	"NetworkBooster/http"
	"NetworkBooster/shadowsocks"
	"NetworkBooster/socks5"
)

type Proxy struct {
	socks5Server *socks5.Server
	httpServer   *http.Server
	ssClient     *shadowsocks.Client
}

func NewProxy(config *Config) (*Proxy, error) {
	proxy := &Proxy{}

	// 创建 Shadowsocks 客户端（连接到远程服务器）
	ssClient, err := shadowsocks.NewClient(
		config.Outbound.Shadowsocks.Server,
		config.Outbound.Shadowsocks.Method,
		config.Outbound.Shadowsocks.Password,
	)
	if err != nil {
		return nil, err
	}
	proxy.ssClient = ssClient

	// 创建 SOCKS5 服务器（如果启用）
	if config.Inbound.SOCKS5.Enabled {
		socks5Server, err := socks5.NewServer(config.Inbound.SOCKS5.Listen)
		if err != nil {
			return nil, fmt.Errorf("创建 SOCKS5 服务器失败: %v", err)
		}
		proxy.socks5Server = socks5Server
	}

	// 创建 HTTP 代理服务器（如果启用）
	if config.Inbound.HTTP.Enabled {
		httpServer, err := http.NewServer(config.Inbound.HTTP.Listen)
		if err != nil {
			if proxy.socks5Server != nil {
				proxy.socks5Server.Close()
			}
			return nil, fmt.Errorf("创建 HTTP 代理服务器失败: %v", err)
		}
		proxy.httpServer = httpServer
	}

	return proxy, nil
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
	log.Printf("Forwarding to Shadowsocks server")
	
	// 启动 SOCKS5 服务器
	if p.socks5Server != nil {
		log.Printf("SOCKS5 server listening on %s", p.socks5Server.Addr())
		go p.runSOCKS5Server()
	}
	
	// 启动 HTTP 代理服务器
	if p.httpServer != nil {
		log.Printf("HTTP proxy server listening on %s", p.httpServer.Addr())
		go p.runHTTPServer()
	}
	
	// 等待（主 goroutine 保持运行）
	select {}
}

func (p *Proxy) runSOCKS5Server() {
	for {
		conn, err := p.socks5Server.Accept()
		if err != nil {
			log.Printf("SOCKS5 Accept error: %v", err)
			continue
		}
		
		go p.handleSOCKS5Connection(conn)
	}
}

func (p *Proxy) runHTTPServer() {
	for {
		conn, err := p.httpServer.Accept()
		if err != nil {
			log.Printf("HTTP Accept error: %v", err)
			continue
		}
		
		go p.handleHTTPConnection(conn)
	}
}

func (p *Proxy) Close() error {
	var errs []error
	if p.socks5Server != nil {
		if err := p.socks5Server.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if p.httpServer != nil {
		if err := p.httpServer.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("关闭服务器时出错: %v", errs)
	}
	return nil
}

func (p *Proxy) handleSOCKS5Connection(conn net.Conn) {
	defer conn.Close()
	
	// 处理 SOCKS5 连接
	req, err := p.socks5Server.HandleConnection(conn)
	if err != nil {
		log.Printf("Handle SOCKS5 connection error: %v", err)
		return
	}
	
	log.Printf("SOCKS5: Connecting to %s via Shadowsocks", req.String())
	
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

func (p *Proxy) handleHTTPConnection(conn net.Conn) {
	defer conn.Close()
	
	// 处理 HTTP 代理连接
	req, err := p.httpServer.HandleConnection(conn)
	if err != nil {
		log.Printf("Handle HTTP connection error: %v", err)
		p.httpServer.SendErrorReply(conn, 400, "Bad Request")
		return
	}
	
	log.Printf("HTTP: Connecting to %s via Shadowsocks", req.String())
	
	// 通过 Shadowsocks 连接到目标
	ssConn, err := p.ssClient.Dial(req.Address, req.Port)
	if err != nil {
		log.Printf("Shadowsocks dial error: %v", err)
		p.httpServer.SendErrorReply(conn, 502, "Bad Gateway")
		return
	}
	defer ssConn.Close()
	
	// 发送成功响应
	if err := p.httpServer.SendSuccessReply(conn); err != nil {
		log.Printf("Send reply error: %v", err)
		return
	}
	
	// 双向转发数据
	go io.Copy(ssConn, conn)
	io.Copy(conn, ssConn)
}
