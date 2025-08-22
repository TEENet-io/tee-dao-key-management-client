# TEENet Signature Tool

一个用于TEENet密钥管理系统的签名和验证工具，采用单体架构设计，提供Web界面和REST API。

## 项目结构

```
signature-tool/
├── main.go                    # 主程序和HTTP路由
├── types.go                   # 数据结构定义
├── crypto.go                  # 加密和签名验证逻辑
├── server.go                  # 静态文件服务
├── voting.go                  # 投票处理逻辑
├── go.mod                     # Go模块配置
├── go.sum                     # Go依赖锁定
├── Dockerfile                 # Docker配置
├── frontend/                  # 前端静态文件
│   ├── index.html            # 主页面
│   ├── styles.css            # 样式文件
│   └── app.js                # JavaScript逻辑
└── README.md                  # 项目文档
```

## 功能特性

- **数字签名**: 使用TEE密钥管理系统对消息进行签名
- **签名验证**: 验证数字签名的有效性
- **多方投票签名**: 支持多个TEE节点的投票式签名机制
- **多协议支持**: 支持ECDSA和Schnorr协议
- **多曲线支持**: 支持ED25519、SECP256K1、SECP256R1曲线
- **Web界面**: 直观的Web操作界面
- **REST API**: 完整的API接口
- **Docker支持**: 完整的容器化部署方案
- **容器路径支持**: 支持代理访问场景的动态路径处理

## 系统要求

- Go 1.24 or higher
- TEENet TEE Configuration Server
- 有效的App ID配置

## 安装部署

### Docker部署（推荐）

1. **构建镜像**:
   ```bash
   docker build -t teenet-signature-tool:latest .
   ```

2. **运行容器**:
   ```bash
   docker run -d \
     --name signature-tool \
     -p 8080:8080 \
     -e APP_ID="your-app-id-here" \
     -e TEE_CONFIG_ADDR="your-tee-server:50052" \
     teenet-signature-tool:latest
   ```

### 本地开发

1. **安装依赖**:
   ```bash
   go mod download
   ```

2. **设置环境变量**:
   ```bash
   export APP_ID="your-app-id-here"
   export TEE_CONFIG_ADDR="localhost:50052"
   export PORT="8080"
   ```

3. **构建运行**:
   ```bash
   go build -o signature-tool .
   ./signature-tool
   ```

### 从源码构建Docker镜像

```bash
docker build -t teenet-signature-tool:latest .
```

## 配置说明

### 环境变量

| 变量 | 描述 | 默认值 | 必需 |
|------|------|--------|------|
| `APP_ID` | 应用ID，用于签名操作 | - | 是 |
| `TEE_CONFIG_ADDR` | TEE配置服务器地址 | `localhost:50052` | 否 |
| `PORT` | Web服务器端口 | `8080` | 否 |
| `FRONTEND_PATH` | 前端文件路径 | `./frontend` | 否 |

## 使用说明

### Web界面

访问Web界面：
```
http://localhost:8080
```

或通过代理访问（支持容器路径）：
```
http://your-domain/container/your-app-id/
```

Web界面提供以下功能：

1. **简单签名**: 使用配置的App ID对消息进行签名
2. **简单验证**: 使用App ID验证签名
3. **多方投票签名**: 发起多个TEE节点参与的投票式签名
4. **投票签名验证**: 验证投票签名的有效性
4. **获取公钥**: 获取App ID对应的公钥

### REST API

#### 健康检查
```http
GET /api/health
```

#### 获取配置
```http
GET /api/config
```

#### 获取公钥
```http
POST /api/get-public-key
Content-Type: application/json

{
  "app_id": "your-app-id"
}
```

#### 签名（使用App ID）
```http
POST /api/sign-with-appid
Content-Type: application/json

{
  "app_id": "your-app-id",
  "message": "Hello, World!"
}
```

#### 验证签名（使用App ID）
```http
POST /api/verify-with-appid
Content-Type: application/json

{
  "app_id": "your-app-id",
  "message": "Hello, World!",
  "signature": "hex-encoded-signature"
}
```


#### 多方投票签名
```http
POST /api/vote
Content-Type: application/json

{
  "description": "Hello TEENet! This is a test message for multi-party signing.",
  "target_app_ids": ["app-id-1", "app-id-2", "app-id-3"],
  "required_votes": 2,
  "total_participants": 3
}
```

**响应**:
```json
{
  "success": true,
  "task_id": "voting-task-uuid",
  "message": "Voting completed and signature generated",
  "voting_results": {
    "total_responses": 3,
    "successful_votes": 2,
    "required_votes": 2,
    "voting_complete": true,
    "final_result": "success",
    "vote_details": [...]
  },
  "signature": "hex-encoded-multi-party-signature",
  "timestamp": "2025-08-22T12:34:56Z"
}
```

## 支持的协议和曲线

### 协议
- `ecdsa`: 椭圆曲线数字签名算法
- `schnorr`: Schnorr签名方案

### 曲线
- `ed25519`: Edwards25519曲线
- `secp256k1`: SECP256K1曲线（比特币曲线）
- `secp256r1`: SECP256R1曲线（NIST P-256）

## 错误处理

所有API端点返回一致的错误响应：

```json
{
  "success": false,
  "error": "错误描述"
}
```

常见错误场景：
- 无效的请求格式
- 缺少必需字段
- 无效的公钥格式（必须是base64）
- 无效的签名格式（必须是hex）
- 不支持的协议或曲线
- TEE服务器通信错误

## 安全考虑

1. **环境变量**: 敏感配置存储在环境变量中
2. **网络安全**: 确保TEE配置服务器得到适当保护
3. **输入验证**: 所有输入在处理前都会被验证
4. **错误消息**: 错误消息不暴露敏感信息
5. **CORS**: 为Web界面访问启用了CORS

## 技术特性

### 动态路径处理
- 支持直接访问：`http://localhost:8080/`
- 支持代理访问：`http://domain/container/app-id/`
- CSS和JS文件动态路径解析
- 基于当前URL路径自动适配

### 静态文件服务
- 集成的静态文件服务器
- 支持HTML、CSS、JS、图片等文件类型
- 安全的路径验证，防止目录遍历攻击

## 故障排除

### 常见问题

1. **"APP_ID environment variable is required"**
   - 解决方案: 设置`APP_ID`环境变量

2. **"Failed to initialize TEE client"**
   - 解决方案: 确保TEE配置服务器正在运行且可访问

3. **CSS/JS文件404错误**
   - 解决方案: 确保使用最新版本的镜像，支持动态路径处理

4. **"Invalid public key format"**
   - 解决方案: 确保公钥是base64编码

5. **"Invalid signature format"**
   - 解决方案: 确保签名是hex编码

### 日志
应用程序记录重要事件，包括：
- 服务器启动信息
- API请求处理
- 错误条件
- 签名操作结果

## 多方投票签名流程

### 投票流程图

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend UI   │    │  Signature Tool │    │ TEE DAO Client  │
│                 │    │     Backend     │    │                 │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          │ 1. POST /api/vote    │                      │
          ├─────────────────────►│                      │
          │ {                    │                      │
          │   description,       │                      │
          │   target_app_ids,    │                      │
          │   required_votes     │                      │
          │ }                    │                      │
          │                      │                      │
          │                      │ 2. VotingSign()      │
          │                      ├─────────────────────►│
          │                      │                      │
          │                      │                      │ 3. 并发发送投票请求
          │                      │                      │ ┌─────────────────┐
          │                      │                      │ │                 │
          │                      │                      ├─┤ Target App ID 1 │
          │                      │                      │ │                 │
          │                      │                      │ └─────────────────┘
          │                      │                      │ ┌─────────────────┐
          │                      │                      │ │                 │
          │                      │                      ├─┤ Target App ID 2 │
          │                      │                      │ │                 │
          │                      │                      │ └─────────────────┘
          │                      │                      │ ┌─────────────────┐
          │                      │                      │ │                 │
          │                      │                      ├─┤ Target App ID N │
          │                      │                      │ │                 │
          │                      │                      │ └─────────────────┘
          │                      │                      │
          │                      │                      │ 4. 等待所有投票结果
          │                      │                      │ (不再提前终止)
          │                      │                      │
          │                      │ 5. 投票结果汇总       │
          │                      │ (包含详细投票信息)     │
          │                      │◄─────────────────────┤
          │                      │                      │
          │                      │ 6. 如果通过则生成签名 │
          │                      ├─────────────────────►│ SignWithAppID()
          │                      │                      │
          │                      │ 签名结果              │
          │                      │◄─────────────────────┤
          │                      │ 7. 最终结果          │
          │                      │ {                    │
          │                      │   success,           │
          │                      │   task_id,           │
          │                      │   voting_results: {  │
          │                      │     vote_details[]   │
          │                      │   },                 │
          │                      │   signature          │
          │                      │ }                    │
          │                      │◄─────────────────────┤
          │                      │                      │
          │ 8. 显示投票结果       │                      │
          │ 包括：               │                      │
          │ - 每个节点投票状态    │                      │
          │ - 最终签名结果        │                      │
          │◄─────────────────────┤                      │
          │                      │                      │
```

### 投票机制特点

1. **M-of-N阈值投票**: 配置所需通过票数，如3个节点中需要2票通过
2. **并发投票**: 同时向所有目标App ID发送投票请求
3. **完整收集**: 等待所有投票响应完成，不提前终止
4. **详细记录**: 记录每个节点的投票状态和错误信息
5. **自动签名**: 投票通过后自动使用配置的App ID生成签名
6. **任务跟踪**: 每个投票轮次都有唯一的Task ID用于跟踪

### 投票响应格式

```json
{
  "success": true,
  "task_id": "vote_app-id_1692708123456789",
  "message": "Voting completed and signature generated by app-id for task vote_app-id_1692708123456789",
  "voting_results": {
    "total_responses": 3,
    "successful_votes": 2,
    "required_votes": 2,
    "voting_complete": true,
    "final_result": "APPROVED",
    "vote_details": [
      {
        "client_id": "app-id-1",
        "success": true,
        "response": true,
        "error": ""
      },
      {
        "client_id": "app-id-2", 
        "success": true,
        "response": true,
        "error": ""
      },
      {
        "client_id": "app-id-3",
        "success": true,
        "response": false,
        "error": ""
      }
    ]
  },
  "signature": "3045022100ab1234...def890",
  "timestamp": "2025-08-22T12:34:56Z"
}
```

## 开发指南

### 代码架构

项目采用模块化设计：
- **main.go**: HTTP服务器和API路由定义
- **types.go**: 请求/响应数据结构
- **crypto.go**: 签名验证和加密算法实现
- **server.go**: 静态文件服务和路径处理
- **voting.go**: 投票处理和自定义投票逻辑

### 依赖项
- `github.com/TEENet-io/tee-dao-key-management-client/go`: TEENet密钥管理客户端（使用远程依赖）
- `github.com/gin-gonic/gin`: Web框架
- 标准Go加密库

### 构建
```bash
go build -o signature-tool .
```

### Docker构建
```bash
docker build -t teenet-signature-tool:latest .
```

## 许可证

此项目是TEENet密钥管理客户端的一部分，遵循相同的许可条款。

## 贡献指南

在为此工具做贡献时：
1. 遵循Go编码标准
2. 添加适当的错误处理
3. 为新功能更新文档
4. 提交更改前进行充分测试