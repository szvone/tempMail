# TempMail - 多存储模式临时邮箱系统

修改自 [@Jlan45](https://github.com/Jlan45/temporaryMailbox)

极简临时邮箱系统，支持三种存储模式：内存、Redis、MySQL，支持多域名，邮件阅后即焚。

## 功能特性

- **多存储模式支持**：内存、Redis、MySQL三种存储方式
- **智能清理机制**：
  - 内存模式：定时清理防止内存泄露
  - Redis模式：自动5分钟过期
  - MySQL模式：软删除，保留历史记录
- **多域名支持**：支持配置多个邮箱域名
- **HTTPS支持**：可配置SSL证书启用HTTPS
- **阅后即焚**：邮件提取后自动删除/标记
- **RESTful API**：提供简洁的HTTP接口

## 快速开始

### 环境要求

- Go 1.16+
- Redis (如使用Redis存储模式)
- MySQL (如使用MySQL存储模式)

### 安装运行

1. 克隆项目
```bash
git clone <repository-url>
cd tempMail-main
```

2. 配置环境变量
复制 `.env` 文件并根据需要修改配置

3. 运行应用
```bash
go run .
```

## 配置说明

### 环境变量配置 (.env)

```bash
# 允许的域名,英文逗号分隔
ALLOWED_DOMAINS=domain1,domain2,domain3

# SMTP 和 HTTP 服务端口 ，默认即可，不建议修改
SMTP_PORT=25
HTTP_PORT=80

# HTTPS 配置
ENABLE_HTTPS=true
HTTPS_PORT=443
CERT_FILE=./certs/server.pem
KEY_FILE=./certs/server.key

# 存储模式配置 (memory, redis, mysql)
STORAGE_MODE=memory

# Redis配置 (仅当STORAGE_MODE=redis时生效)
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# MySQL配置 (仅当STORAGE_MODE=mysql时生效)
MYSQL_DSN=root:password@tcp(localhost:3306)/temp_mail?charset=utf8mb4&parseTime=True&loc=Local

# 内存存储清理间隔（分钟）
MEMORY_CLEANUP_INTERVAL=5
```

### 存储模式说明

#### 1. 内存模式 (memory)
- **特点**：数据存储在内存中，重启后丢失
- **清理机制**：定时清理，防止内存泄露
- **适用场景**：开发测试、临时使用

#### 2. Redis模式 (redis)
- **特点**：数据存储在Redis中，支持持久化
- **清理机制**：自动5分钟过期，无需手动清理
- **适用场景**：生产环境，需要高性能

#### 3. MySQL模式 (mysql)
- **特点**：数据存储在MySQL数据库
- **清理机制**：软删除，保留历史记录
- **适用场景**：需要数据持久化和查询功能

## API接口

### 1. 获取允许的域名列表

**请求**
```http
GET /getAllowedDomains
```

**响应**
```json
{
  "allowedDomains": ["domain1", "domain2", "domain3"]
}
```

### 2. 获取邮件

**请求**
```http
GET /getMail/{邮箱地址}
```

**参数**
- `{邮箱地址}`：完整的邮箱地址，如 `test@domain1`

**响应**
- 成功 (200)
```json
{
  "mail": {
    "from": "sender@example.com",
    "title": "邮件标题",
    "TextContent": "邮件文本内容",
    "HtmlContent": "邮件HTML内容"
  }
}
```

- 无邮件 (201)
```json
{
  "mail": "没有邮件"
}
```

- 错误 (500)
```json
{
  "error": "错误信息"
}
```

## 部署说明

### DNS配置

1. **主域名**：解析A记录到服务器IP
2. **MX记录**：为每个邮箱域名配置MX记录指向服务器

### 证书配置

如需启用HTTPS，需要配置SSL证书：

1. 将证书文件放置于 `./certs/` 目录
2. 修改 `.env` 中的证书路径配置
3. 设置 `ENABLE_HTTPS=true`

### 存储模式选择建议

- **开发测试**：使用内存模式，无需额外依赖
- **小型部署**：使用Redis模式，性能好且自动清理
- **生产环境**：使用MySQL模式，数据持久化且可查询历史

## 技术架构

- **SMTP服务器**：基于 `github.com/alash3al/go-smtpsrv/v3`
- **HTTP服务器**：基于 `github.com/gin-gonic/gin`
- **存储抽象**：统一的Storage接口，支持多存储后端
- **配置管理**：环境变量配置，支持 `.env` 文件

## 存储模式对比

| 特性 | 内存模式 | Redis模式 | MySQL模式 |
|------|----------|-----------|-----------|
| 性能 | 最高 | 高 | 中等 |
| 持久化 | 否 | 可选 | 是 |
| 数据保留 | 定时清理 | 5分钟过期 | 永久保留 |
| 部署复杂度 | 简单 | 中等 | 中等 |
| 适用场景 | 测试开发 | 生产环境 | 数据审计 |

## 故障排除

### 常见问题

1. **SMTP服务无法启动**
   - 检查端口25是否被占用
   - 检查防火墙设置

2. **邮件无法接收**
   - 检查MX记录配置
   - 检查域名解析

3. **存储连接失败**
   - 检查Redis/MySQL服务状态
   - 验证连接配置参数

### 日志查看

应用会输出详细的运行日志，包括：
- 存储模式初始化信息
- 邮件接收和提取记录
- 清理任务执行情况
- 错误和异常信息

## 许可证

本项目基于开源许可证发布，具体参见LICENSE文件。

## 贡献

欢迎提交Issue和Pull Request来改进本项目。