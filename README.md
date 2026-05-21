# SharedLink 🚀

> 局域网点对点大文件传输工具 — 跨平台、零配置、纯 CLI + TUI

SharedLink 是一个用 Go 编写的局域网 P2P 文件传输工具。无需服务器中转，无需配置网络，两台电脑装上就能互传文件。支持 **Windows / macOS / Linux** 三端互传

---

## ✨ 功能特性

- **点对点直连** — 纯 P2P 传输，文件不经过任何服务器
- **mDNS 自动发现** — 扫描局域网内的发送端，无需手动输入 IP
- **精美 TUI 界面** — 实时进度条、传输速度、剩余时间一目了然
- **大文件支持** — 4MB 分片传输，流式读写，不占用大量内存
- **完整性校验** — SHA-256 逐片校验 + 全文件校验，确保数据完整
- **跨平台** — Windows / macOS / Linux 三端互通，单二进制零依赖
- **断点续传** — 支持传输中断后继续（规划中）

---

## 📦 快速开始

### 下载

从 [Releases](https://github.com/sglwsjxh/SharedLink/releases) 页面下载对应平台的二进制文件，或自行编译。

### 发送文件

```bash
# 发送端监听端口，等待接收端连接
sharedlink send ./bigfile.mp4
```

### 接收文件

```bash
# 方式一：扫描局域网，选择发送端
sharedlink recv

# 方式二：直连指定地址
sharedlink recv 192.168.1.100:53349
```

> **注意**：发送端先运行，接收端后运行。

---

## 🛠 命令说明

| 命令 | 说明 |
|------|------|
| `sharedlink send <文件路径>` | 发送文件，启动 TCP 监听并广播 mDNS 服务 |
| `sharedlink recv` | 扫描局域网内可用的发送端，选择后接收 |
| `sharedlink recv <ip>:<port>` | 直连指定地址接收文件 |

---

## 🔧 自行编译

### 环境要求

- Go 1.21+

### 编译当前平台

```bash
git clone https://github.com/sglwsjxh/SharedLink.git
cd SharedLink
go build -o sharedlink ./cmd/sharedlink
```

### 交叉编译所有平台

```bash
make build-all
```

或手动编译：

```bash
# Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o sharedlink-windows.exe ./cmd/sharedlink

# macOS (Intel)
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o sharedlink-darwin-amd64 ./cmd/sharedlink

# macOS (Apple Silicon)
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o sharedlink-darwin-arm64 ./cmd/sharedlink

# Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o sharedlink-linux ./cmd/sharedlink
```

---

## 🏗 项目结构

```
sharedlink/
├── cmd/
│   └── sharedlink/          CLI 入口（cobra）
│       ├── main.go          程序入口
│       ├── send.go          发送命令
│       └── recv.go          接收命令
├── internal/
│   ├── protocol/            二进制协议编解码
│   ├── transfer/            TCP 传输引擎
│   ├── discover/            mDNS 服务发现
│   └── ui/                  TUI 界面（bubbletea）
├── Makefile                 交叉编译
├── go.mod
├── LICENSE
└── README.md
```

---

## 🧱 技术栈

| 模块 | 技术 |
|------|------|
| 编程语言 | Go |
| CLI 框架 | [Cobra](https://github.com/spf13/cobra) |
| TUI 框架 | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| 样式 | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| 服务发现 | [hashicorp/mdns](https://github.com/hashicorp/mdns) |
| 传输协议 | 自定义二进制 TCP 协议 |

---

## 📄 许可证

本项目基于 MIT 许可证开源 — 详见 [LICENSE](LICENSE) 文件。
