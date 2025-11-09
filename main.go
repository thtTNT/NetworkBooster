package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	// 命令行参数（只保留必要的）
	configPath := flag.String("config", "config.json", "配置文件路径")
	initConfig := flag.Bool("init", false, "生成默认配置文件")
	flag.Parse()

	// 如果指定了 -init，生成默认配置文件
	if *initConfig {
		if err := SaveDefaultConfig(*configPath); err != nil {
			log.Fatalf("生成配置文件失败: %v", err)
		}
		fmt.Printf("已生成默认配置文件: %s\n", *configPath)
		fmt.Println("请编辑配置文件并填写 Shadowsocks 服务器信息")
		return
	}

	// 加载配置文件
	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 验证必要参数
	if config.Outbound.Shadowsocks.Server == "" || config.Outbound.Shadowsocks.Password == "" {
		log.Fatal("请指定 Shadowsocks 服务器地址和密码（在配置文件中）")
	}

	// 验证至少启用一个协议
	if !config.Inbound.SOCKS5.Enabled && !config.Inbound.HTTP.Enabled {
		log.Fatal("至少需要启用一个协议（SOCKS5 或 HTTP）")
	}

	// 创建代理
	proxy, err := NewProxy(config)
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


