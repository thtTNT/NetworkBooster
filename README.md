# NetworkBooster

一个支持 SOCKS5 和 HTTP 代理到 Shadowsocks 的代理工具，将本地代理连接转换为 Shadowsocks 协议。

## 功能特性

- ✅ 本地 SOCKS5 服务器（监听本地连接）
- ✅ 本地 HTTP 代理服务器（支持 HTTP/HTTPS）
- ✅ 可配置启用/禁用 SOCKS5 和 HTTP 代理
- ✅ Shadowsocks 客户端（连接到远程服务器）
- ✅ 自动验证 Shadowsocks 服务器可用性
- ✅ 支持配置文件（inbound/outbound 结构）
- ✅ 自动配置迁移（旧格式自动升级）

## 安装

```bash
go build
```

## 使用方法

### 1. 生成配置文件

首次使用，生成默认配置文件：

```bash
./NetworkBooster -init
```

**注意**: 如果配置文件已存在，不会覆盖。如需重新生成，请先删除现有配置文件。

或指定配置文件路径：

```bash
./NetworkBooster -init -config my-config.json
```

### 2. 编辑配置文件

编辑 `config.json`，填写 Shadowsocks 服务器信息：

```json
{
  "inbound": {
    "socks5": {
      "enabled": true,
      "listen": "127.0.0.1:1080"
    },
    "http": {
      "enabled": false,
      "listen": "127.0.0.1:8080"
    }
  },
  "outbound": {
    "shadowsocks": {
      "server": "your-server.com:8388",
      "method": "aes-256-gcm",
      "password": "your-password"
    }
  }
}
```

配置项说明：

**inbound（入站配置）**：
- `inbound.socks5.enabled`: 是否启用 SOCKS5 代理（默认: `true`）
- `inbound.socks5.listen`: SOCKS5 监听地址（默认: `127.0.0.1:1080`）
- `inbound.http.enabled`: 是否启用 HTTP 代理（默认: `false`）
- `inbound.http.listen`: HTTP 代理监听地址（默认: `127.0.0.1:8080`）

**outbound（出站配置）**：
- `outbound.shadowsocks.server`: Shadowsocks 服务器地址（必填，格式: `host:port`）
- `outbound.shadowsocks.method`: 加密方法（默认: `aes-256-gcm`，可选: `chacha20-poly1305` 等）
- `outbound.shadowsocks.password`: Shadowsocks 密码（必填）

**注意**: 
- 至少需要启用一个 inbound 协议（`inbound.socks5.enabled` 或 `inbound.http.enabled` 至少一个为 `true`）
- 如果使用旧格式配置文件，程序会自动迁移到新格式（原文件会备份为 `.backup`）

### 3. 运行程序

使用配置文件运行：

```bash
./NetworkBooster
```

或指定配置文件路径：

```bash
./NetworkBooster -config my-config.json
```

### 4. 配置文件迁移

如果检测到旧格式配置文件，程序会自动迁移到新格式：

1. 自动备份原配置文件为 `config.json.backup`
2. 将旧格式转换为新的 `inbound`/`outbound` 结构
3. 保存新格式配置文件

**旧格式示例**（会自动迁移）：
```json
{
  "enable_socks5": true,
  "socks5_addr": "127.0.0.1:1080",
  "enable_http": false,
  "http_addr": "127.0.0.1:8080",
  "ss_server_addr": "example.com:8388",
  "ss_method": "aes-256-gcm",
  "ss_password": "password"
}
```

**新格式**（推荐）：
```json
{
  "inbound": {
    "socks5": {
      "enabled": true,
      "listen": "127.0.0.1:1080"
    },
    "http": {
      "enabled": false,
      "listen": "127.0.0.1:8080"
    }
  },
  "outbound": {
    "shadowsocks": {
      "server": "example.com:8388",
      "method": "aes-256-gcm",
      "password": "password"
    }
  }
}
```

## 示例

### 使用配置文件

```bash
# 1. 生成配置文件
./NetworkBooster -init

# 2. 编辑 config.json，填写服务器信息

# 3. 运行
./NetworkBooster
```

### 使用自定义配置文件路径

```bash
./NetworkBooster -config /path/to/my-config.json
```

### 同时启用 SOCKS5 和 HTTP

编辑 `config.json`：

```json
{
  "inbound": {
    "socks5": {
      "enabled": true,
      "listen": "127.0.0.1:1080"
    },
    "http": {
      "enabled": true,
      "listen": "127.0.0.1:8080"
    }
  },
  "outbound": {
    "shadowsocks": {
      "server": "example.com:8388",
      "method": "aes-256-gcm",
      "password": "your-password"
    }
  }
}
```

## 配置应用使用代理

程序启动后，配置您的应用程序使用代理：

### SOCKS5 代理
- **代理地址**: `127.0.0.1:1080`（可在配置文件中修改）
- **代理类型**: SOCKS5

### HTTP 代理
- **代理地址**: `127.0.0.1:8080`（可在配置文件中修改）
- **代理类型**: HTTP/HTTPS

**提示**: 可以同时启用 SOCKS5 和 HTTP 代理，不同应用程序可以选择使用不同的协议。

## 故障排除

### Shadowsocks 服务器验证失败

程序启动时会自动验证服务器连接，如果失败会显示详细错误信息：

```
Shadowsocks 服务器验证失败: 无法连接到 Shadowsocks 服务器 xxx:xxx: 连接被拒绝
```

常见原因：
1. 服务器地址或端口错误
2. 网络连接问题
3. 密码或加密方法不匹配
4. 防火墙阻止连接

### 查看帮助

```bash
./NetworkBooster -h
```

