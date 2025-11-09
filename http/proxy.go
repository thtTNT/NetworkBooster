package http

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
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

// HandleConnection 处理 HTTP 代理连接
func (s *Server) HandleConnection(conn net.Conn) (*Request, error) {
	reader := bufio.NewReader(conn)

	// 读取 HTTP 请求行
	requestLine, _, err := reader.ReadLine()
	if err != nil {
		return nil, err
	}

	// 解析请求行: METHOD URL HTTP/VERSION
	parts := strings.Fields(string(requestLine))
	if len(parts) < 2 {
		return nil, errors.New("invalid HTTP request")
	}

	method := parts[0]
	requestURL := parts[1]

	var host string
	var port uint16

	// 处理 CONNECT 方法（用于 HTTPS）
	if method == "CONNECT" {
		// CONNECT 方法的 URL 格式: host:port
		host, portStr, err := net.SplitHostPort(requestURL)
		if err != nil {
			return nil, fmt.Errorf("解析 CONNECT 目标地址失败: %v", err)
		}

		// 读取并消耗掉 HTTP 头部（直到空行）
		for {
			line, _, err := reader.ReadLine()
			if err != nil {
				return nil, err
			}
			if len(line) == 0 {
				break
			}
		}

		port = parsePort(portStr)
		return &Request{
			Address: host,
			Port:    port,
			Conn:    conn,
		}, nil
	}

	// 处理其他 HTTP 方法（GET, POST 等）
	parsedURL, err := parseURL(requestURL)
	if err != nil {
		return nil, fmt.Errorf("解析 URL 失败: %v", err)
	}

	host = parsedURL.host
	port = parsedURL.port

	// 读取并消耗掉 HTTP 头部（直到空行）
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if len(line) == 0 {
			break
		}
		// 检查 Host 头，可能包含端口信息
		if strings.HasPrefix(strings.ToLower(string(line)), "host:") {
			hostPart := strings.TrimSpace(string(line[5:]))
			if strings.Contains(hostPart, ":") {
				h, p, err := net.SplitHostPort(hostPart)
				if err == nil {
					host = h
					port = parsePort(p)
				}
			} else {
				host = hostPart
			}
		}
	}

	return &Request{
		Address: host,
		Port:    port,
		Conn:    conn,
	}, nil
}

func parsePort(portStr string) uint16 {
	var port uint16
	fmt.Sscanf(portStr, "%d", &port)
	return port
}

type urlParts struct {
	host string
	port uint16
}

func parseURL(rawURL string) (*urlParts, error) {
	// 简单 URL 解析：http://host:port/path 或 https://host:port/path
	if !strings.Contains(rawURL, "://") {
		// 可能是相对 URL，需要从 Host 头获取
		return nil, errors.New("需要绝对 URL 或 Host 头")
	}

	parts := strings.SplitN(rawURL, "://", 2)
	if len(parts) != 2 {
		return nil, errors.New("无效的 URL 格式")
	}

	scheme := parts[0]
	rest := parts[1]

	// 查找第一个 / 或 ?
	idx := len(rest)
	for i, c := range rest {
		if c == '/' || c == '?' || c == '#' {
			idx = i
			break
		}
	}

	authority := rest[:idx]

	var host string
	var port uint16

	if strings.Contains(authority, ":") {
		host, portStr, err := net.SplitHostPort(authority)
		if err != nil {
			return nil, err
		}
		port = parsePort(portStr)
		host = host
	} else {
		host = authority
		if scheme == "https" {
			port = 443
		} else {
			port = 80
		}
	}

	return &urlParts{
		host: host,
		port: port,
	}, nil
}

// SendSuccessReply 发送成功响应
func (s *Server) SendSuccessReply(conn net.Conn) error {
	_, err := conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	return err
}

// SendErrorReply 发送错误响应
func (s *Server) SendErrorReply(conn net.Conn, code int, message string) error {
	response := fmt.Sprintf("HTTP/1.1 %d %s\r\n\r\n", code, message)
	_, err := conn.Write([]byte(response))
	return err
}

// Request 表示 HTTP 代理请求
type Request struct {
	Address string
	Port    uint16
	Conn    net.Conn
}

func (r *Request) String() string {
	return net.JoinHostPort(r.Address, fmt.Sprintf("%d", r.Port))
}

