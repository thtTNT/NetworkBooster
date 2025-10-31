package shadowsocks

import (
	"fmt"
	"net"

	"github.com/shadowsocks/go-shadowsocks2/core"
	"github.com/shadowsocks/go-shadowsocks2/socks"
)

type Client struct {
	cipher core.StreamConnCipher
	serverAddr string
}

func NewClient(serverAddr, method, password string) (*Client, error) {
	// 创建 Shadowsocks 加密器
	cipher, err := core.PickCipher(method, []byte{}, password)
	if err != nil {
		return nil, fmt.Errorf("创建加密器失败: %v", err)
	}

	// 类型断言为 StreamConnCipher
	streamCipher, ok := cipher.(core.StreamConnCipher)
	if !ok {
		return nil, fmt.Errorf("加密器不支持流连接: %s", method)
	}

	return &Client{
		cipher:     streamCipher,
		serverAddr: serverAddr,
	}, nil
}

func (c *Client) ServerAddr() string {
	return c.serverAddr
}

// TestConnection 测试 Shadowsocks 服务器连接
func (c *Client) TestConnection() error {
	// 首先测试能否建立 TCP 连接到 Shadowsocks 服务器
	tcpConn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		return fmt.Errorf("无法连接到 Shadowsocks 服务器 %s: %v", c.serverAddr, err)
	}
	tcpConn.Close()
	
	// 然后尝试进行一次完整的 Shadowsocks 连接测试
	// 使用一个快速响应的地址：8.8.8.8:53 (Google DNS)
	testConn, err := c.Dial("8.8.8.8", 53)
	if err != nil {
		return fmt.Errorf("Shadowsocks 协议测试失败: %v (请检查密码和加密方法)", err)
	}
	testConn.Close()
	return nil
}

// Dial 连接到 Shadowsocks 服务器并发送目标地址
func (c *Client) Dial(targetAddr string, targetPort uint16) (net.Conn, error) {
	// 连接到 Shadowsocks 服务器
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		return nil, fmt.Errorf("连接 Shadowsocks 服务器失败: %v", err)
	}

	// 创建加密连接
	ssConn := c.cipher.StreamConn(conn)

	// 构建目标地址
	addr := socks.ParseAddr(net.JoinHostPort(targetAddr, fmt.Sprintf("%d", targetPort)))
	if addr == nil {
		ssConn.Close()
		return nil, fmt.Errorf("无效的目标地址: %s:%d", targetAddr, targetPort)
	}

	// 发送目标地址
	if _, err := ssConn.Write(addr); err != nil {
		ssConn.Close()
		return nil, fmt.Errorf("发送目标地址失败: %v", err)
	}

	return ssConn, nil
}
