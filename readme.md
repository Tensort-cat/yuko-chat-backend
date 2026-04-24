# yuko-chat-backend

yuko-chat 是一个基于 Golang 实现的前后端分离仿微信 IM 系统，本仓库为其后端服务。

前端项目地址：
👉 [https://github.com/Tensort-cat/yuko-chat-frontend](https://github.com/Tensort-cat/yuko-chat-frontend)

---

## 🚀 技术栈

* Golang
* Gin
* Gorm
* Zap
* MySQL
* Redis
* Kafka
* Vue（前端）
* Nginx

---

## 📦 项目结构

```
yuko-chat-backend
├─ cmd                 # 程序入口
├─ configs             # 配置文件
├─ internal            # 核心业务代码
│  ├─ controller       # 控制层
│  ├─ service          # 业务逻辑层
│  ├─ dao              # 数据访问层（MySQL / Redis / Kafka）
│  ├─ model            # 数据模型 & DDL
│  ├─ dto              # 请求/响应结构
│  ├─ route            # 路由注册
│  └─ config           # 配置加载
├─ pkg                 # 公共组件（工具类、日志等）
├─ test                # 测试代码
├─ logs                # 日志输出
```

---

## ⚙️ 运行说明

### 1. 修改配置

启动前请根据实际环境修改：

* `configs/config.toml`
* `internal/config/config.go`

---

### 2. 初始化数据库

数据库建表 SQL 位于：

```
/internal/model/tables.sql
```

---

### 3. 启动服务

```bash
go run cmd/main.go
```

---

## ✨ 功能概述

* 用户注册 / 登录（支持验证码）
* 好友 / 群组管理
* 会话管理
* 单聊 / 群聊
* WebSocket 实时通信
* Kafka 异步消息处理
* Redis 缓存支持

---

## 🧪 测试

项目提供部分基础测试：

```
/test
```

可自行运行：

```bash
go test ./...
```

---