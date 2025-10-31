package socks5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

type Server struct {
	listener net.Listener
}

func NewServer(listenAddr string) (*Server, error) {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	
	return &Server{
		listener: listener,
	}, nil
}

func (s *Server) Accept() (net.Conn, error) {
	return s.listener.Accept()
}

func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}

// HandleConnection 处理 SOCKS5 连接
func (s *Server) HandleConnection(conn net.Conn) (*Request, error) {
	// SOCKS5 握手
	if err := s.handshake(conn); err != nil {
		conn.Close()
		return nil, err
	}
	
	// 读取连接请求
	req, err := s.readRequest(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	
	req.Conn = conn
	return req, nil
}

func (s *Server) handshake(conn net.Conn) error {
	// 读取客户端认证方法
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return err
	}
	
	if buf[0] != 0x05 {
		return errors.New("invalid SOCKS version")
	}
	
	numMethods := int(buf[1])
	if numMethods == 0 {
		return errors.New("no authentication methods")
	}
	
	methods := make([]byte, numMethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return err
	}
	
	// 检查是否支持无需认证
	noAuth := false
	for _, method := range methods {
		if method == 0x00 {
			noAuth = true
			break
		}
	}
	
	if !noAuth {
		// 返回不接受任何方法
		conn.Write([]byte{0x05, 0xFF})
		return errors.New("authentication required")
	}
	
	// 发送响应：无需认证
	_, err := conn.Write([]byte{0x05, 0x00})
	return err
}

func (s *Server) readRequest(conn net.Conn) (*Request, error) {
	// 读取请求
	buf := make([]byte, 4)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, err
	}
	
	if buf[0] != 0x05 {
		return nil, errors.New("invalid SOCKS version")
	}
	
	if buf[1] != 0x01 { // 只支持 CONNECT
		s.sendReply(conn, 0x07) // 不支持的命令
		return nil, fmt.Errorf("unsupported command: %d", buf[1])
	}
	
	// 读取地址
	req := &Request{}
	addrType := buf[3]
	
	switch addrType {
	case 0x01: // IPv4
		ip := make([]byte, 4)
		if _, err := io.ReadFull(conn, ip); err != nil {
			return nil, err
		}
		req.Address = net.IP(ip).String()
		
	case 0x03: // Domain
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return nil, err
		}
		domainLen := int(lenBuf[0])
		domain := make([]byte, domainLen)
		if _, err := io.ReadFull(conn, domain); err != nil {
			return nil, err
		}
		req.Address = string(domain)
		
	case 0x04: // IPv6
		ip := make([]byte, 16)
		if _, err := io.ReadFull(conn, ip); err != nil {
			return nil, err
		}
		req.Address = net.IP(ip).String()
		
	default:
		s.sendReply(conn, 0x08) // 不支持的地址类型
		return nil, fmt.Errorf("unsupported address type: %d", addrType)
	}
	
	// 读取端口
	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBuf); err != nil {
		return nil, err
	}
	req.Port = binary.BigEndian.Uint16(portBuf)
	
	// 发送成功响应（延迟到实际连接成功后再发送）
	// 这里暂时不发送，由上层在连接成功后发送
	
	return req, nil
}

func (s *Server) sendReply(conn net.Conn, replyCode byte) error {
	// 发送 SOCKS5 响应
	// 格式: VER | REP | RSV | ATYP | BND.ADDR | BND.PORT
	resp := []byte{0x05, replyCode, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	_, err := conn.Write(resp)
	return err
}

// SendSuccessReply 发送成功响应
func (s *Server) SendSuccessReply(conn net.Conn) error {
	return s.sendReply(conn, 0x00)
}

// Request 表示 SOCKS5 连接请求
type Request struct {
	Address string
	Port    uint16
	Conn    net.Conn
}

func (r *Request) String() string {
	return net.JoinHostPort(r.Address, fmt.Sprintf("%d", r.Port))
}

