package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// 新的配置结构
type Config struct {
	Inbound  InboundConfig  `json:"inbound"`
	Outbound OutboundConfig `json:"outbound"`
}

type InboundConfig struct {
	SOCKS5 InboundProtocol `json:"socks5"`
	HTTP   InboundProtocol `json:"http"`
}

type InboundProtocol struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
}

type OutboundConfig struct {
	Shadowsocks ShadowsocksConfig `json:"shadowsocks"`
}

type ShadowsocksConfig struct {
	Server   string `json:"server"`
	Method   string `json:"method"`
	Password string `json:"password"`
}

// 旧配置结构（用于迁移）
type OldConfig struct {
	EnableSOCKS5 bool   `json:"enable_socks5"`
	SOCKS5Addr   string `json:"socks5_addr"`
	EnableHTTP   bool   `json:"enable_http"`
	HTTPAddr     string `json:"http_addr"`
	SSServerAddr string `json:"ss_server_addr"`
	SSMethod     string `json:"ss_method"`
	SSPassword   string `json:"ss_password"`
}

func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 检查是否为旧格式（通过检查是否存在旧字段）
	var rawConfig map[string]interface{}
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 检测是否为新格式（有 inbound 或 outbound 字段）
	_, hasInbound := rawConfig["inbound"]
	_, hasOutbound := rawConfig["outbound"]

	if hasInbound || hasOutbound {
		// 新格式
		var config Config
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("解析新格式配置失败: %v", err)
		}
		return &config, nil
	}

	// 检测到旧格式，进行迁移
	_, hasOldFormat := rawConfig["enable_socks5"]
	if hasOldFormat || rawConfig["ss_server_addr"] != nil {
		fmt.Println("检测到旧格式配置文件，正在迁移...")
		
		var oldConfig OldConfig
		if err := json.Unmarshal(data, &oldConfig); err != nil {
			return nil, fmt.Errorf("解析旧格式配置失败: %v", err)
		}

		config := migrateFromOldConfig(&oldConfig)

		// 备份原文件
		backupPath := configPath + ".backup"
		if err := os.WriteFile(backupPath, data, 0644); err != nil {
			fmt.Printf("警告: 无法创建备份文件: %v\n", err)
		} else {
			fmt.Printf("原配置文件已备份到: %s\n", backupPath)
		}

		// 保存新格式配置
		newData, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("生成新配置失败: %v", err)
		}

		if err := os.WriteFile(configPath, newData, 0644); err != nil {
			return nil, fmt.Errorf("保存新配置失败: %v", err)
		}

		fmt.Println("配置文件已成功迁移为新格式")
		return &config, nil
	}

	// 既不是新格式也不是旧格式
	return nil, fmt.Errorf("无法识别的配置文件格式")
}

func migrateFromOldConfig(old *OldConfig) Config {
	config := Config{
		Inbound: InboundConfig{
			SOCKS5: InboundProtocol{
				Enabled: old.EnableSOCKS5,
				Listen:  old.SOCKS5Addr,
			},
			HTTP: InboundProtocol{
				Enabled: old.EnableHTTP,
				Listen:  old.HTTPAddr,
			},
		},
		Outbound: OutboundConfig{
			Shadowsocks: ShadowsocksConfig{
				Server:   old.SSServerAddr,
				Method:   old.SSMethod,
				Password: old.SSPassword,
			},
		},
	}

	// 设置默认值
	if config.Inbound.SOCKS5.Listen == "" {
		config.Inbound.SOCKS5.Listen = "127.0.0.1:1080"
	}
	if config.Inbound.HTTP.Listen == "" {
		config.Inbound.HTTP.Listen = "127.0.0.1:8080"
	}
	if config.Outbound.Shadowsocks.Method == "" {
		config.Outbound.Shadowsocks.Method = "aes-256-gcm"
	}

	return config
}

func SaveDefaultConfig(configPath string) error {
	// 检查配置文件是否已存在
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("配置文件 %s 已存在，不会覆盖。如需重新生成，请先删除现有配置文件", configPath)
	}

	defaultConfig := Config{
		Inbound: InboundConfig{
			SOCKS5: InboundProtocol{
				Enabled: true,
				Listen:  "127.0.0.1:1080",
			},
			HTTP: InboundProtocol{
				Enabled: false,
				Listen:  "127.0.0.1:8080",
			},
		},
		Outbound: OutboundConfig{
			Shadowsocks: ShadowsocksConfig{
				Server:   "",
				Method:   "aes-256-gcm",
				Password: "",
			},
		},
	}

	data, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("生成默认配置失败: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}
