# 简易聊天程序

作者：池松泽

学号：523031910367

这是一个基于 Go 语言开发的简易聊天程序，包含服务端和基于 Fyne GUI 的客户端。

## 项目结构

```
├── source/
│   ├── cmd/
│   │   ├── client/            # 客户端代码 
│   │   │   ├── main.go
│   │   │   └── gui.go
│   │   └── server/            # 服务端代码
│   │       └── main.go
│   ├── utils/                 # 通用工具包 (消息定义、编解码)
│   │       └── utils.go
│   ├── go.sum                 # 校验和文件
│   └── go.mod                 # 依赖管理
├── bin/
│   ├── client_linux_amd64.exe # 客户端程序 linux amd64
│   ├── server_linux_amd64.exe # 服务端程序 linux amd64
│   ├── client_win_64.exe      # 客户端程序 win64
│   └── server_win_64.exe      # 服务端程序 win64
├── readme.md                  # 说明文件
└── 项目文档.pdf
```

## 功能特性

- **用户登录**: 支持自定义用户名登录。
- **实时聊天**: 支持多用户实时一对一或者一对多在线聊天。
- **系统通知**: 用户加入或离开时会有系统广播。
- **GUI 界面**: 客户端使用 Fyne 框架构建，提供图形化操作界面。
- **通信协议**: 使用 TCP 连接，消息采用 JSON 格式并经过 Base64 编码传输。

## 环境要求

- Go 1.25+
- Fyne 依赖 (Windows 上通常无需额外安装，Linux/macOS 可能需要安装 GCC 和图形库开发包)

## 快速开始

### 1. 启动服务端

服务端默认监听 `:8080` 端口。

```bash
# 直接运行代码
cd source/cmd/server
go run main.go

# 运行编译好的可执行程序
cd bin
./server_win_64.exe
# 或者指定端口
./server_win_64.exe -port 9000
```

### 2. 启动客户端

客户端启动后会显示登录界面，需输入服务端 IP、端口和用户名。

```bash
# 直接运行代码
cd source/cmd/client
go run ./

# 运行编译好的可执行程序
cd bin
./client_win_64.exe
```

**登录参数说明:**
- **Server IP**: 服务端 IP 地址 (本地测试可用 `localhost`)
- **Server Port**: 服务端端口 (默认 `8080`)
- **Username**: 你的聊天昵称

## 通信协议

客户端与服务端之间通过 TCP Socket 通信。每条消息以换行符 `\n` 分隔。

**消息格式 (JSON):**
```json
{
    "type": "text",        // 消息类型: login, text, broadcast, error 等
    "content": "你好",     // 消息内容
    "sender": "user1",     // 发送者
    "receiver": "user2"    // 接收者
}
```
*注意：实际传输时，上述 JSON 字符串会被 Base64 编码。*

## 开发与构建

**构建客户端 (Windows):**
```bash
cd source/cmd/client
go build -ldflags "-s -w" -o client.exe .
```

**构建服务端:**
```bash
cd source/cmd/server
go build -ldflags "-s -w" -o server.exe .
```

