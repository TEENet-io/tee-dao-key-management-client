# TEENet Signature Tool

一个基于TEENet密钥管理系统的分布式签名和多方投票工具，提供Web界面和REST API。

## 功能特性

- **单方签名**: 使用TEE密钥管理系统对消息进行签名
- **签名验证**: 验证数字签名的有效性  
- **分布式投票签名**: 支持多个TEE节点的M-of-N阈值投票机制
- **多协议支持**: 支持ECDSA和Schnorr协议
- **多曲线支持**: 支持ED25519、SECP256K1、SECP256R1曲线
- **Web界面**: 直观的Web操作界面
- **REST API**: 完整的API接口
- **无缓存部署**: 支持零缓存部署，无需手动清除浏览器缓存

## 项目结构

```
signature-tool/
├── main.go                    # 主程序和HTTP路由
├── types.go                   # 数据结构定义
├── crypto.go                  # 加密和签名验证逻辑
├── server.go                  # 静态文件服务（支持无缓存）
├── voting.go                  # 投票处理逻辑
├── go.mod                     # Go模块配置
├── go.sum                     # Go依赖锁定
├── Dockerfile                 # Docker构建配置
├── frontend/                  # 前端静态文件
│   ├── index.html            # 主页面
│   ├── styles.css            # 样式文件
│   └── app.js                # JavaScript逻辑
└── README.md                  # 项目文档
```

## 系统要求

- Go 1.24 or higher
- TEENet TEE Configuration Server
- 有效的App ID配置
- 服务器端配置的VotingSign项目（用于自动获取投票目标和阈值）

## 安装部署

### Docker部署（推荐）

1. **构建镜像**:
   ```bash
   cd /path/to/signature-tool
   docker build -t teenet-signature-tool .
   ```

2. **运行容器**:
   ```bash
   docker run -d \
     --name signature-tool \
     -p 8080:8080 \
     -e APP_ID="your-app-id-here" \
     -e TEE_CONFIG_ADDR="your-tee-server:50052" \
     teenet-signature-tool
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

Web界面提供以下功能：

1. **单方签名**: 使用配置的App ID对消息进行签名
2. **签名验证**: 验证签名的有效性
3. **分布式投票签名**: 发起多个TEE节点参与的投票式签名
4. **投票签名验证**: 验证投票签名的有效性
5. **获取公钥**: 获取App ID对应的公钥信息

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

#### 单方签名
```http
POST /api/sign-with-appid
Content-Type: application/json

{
  "app_id": "your-app-id",
  "message": "Hello, World!"
}
```

#### 验证签名
```http
POST /api/verify-with-appid
Content-Type: application/json

{
  "app_id": "your-app-id",
  "message": "Hello, World!",
  "signature": "hex-encoded-signature"
}
```

#### 分布式投票签名
```http
POST /api/vote
Content-Type: application/json

{
  "message": "dGVzdA==",
  "signer_app_id": "app-1"
}
```

**注意**: 目标应用ID列表（target_app_ids）和所需票数（required_votes）现在由服务器根据VotingSign项目配置自动获取，无需在请求中指定。

**响应格式**:
```json
{
  "success": true,
  "approved": true,
  "app_id": "app-1",
  "message": "APPROVED",
  "voting_results": {
    "voting_complete": true,
    "successful_votes": 2,
    "required_votes": 2,
    "total_targets": 3,
    "final_result": "APPROVED",
    "vote_details": [
      {
        "client_id": "app-1",
        "success": true,
        "response": true
      },
      {
        "client_id": "app-2",
        "success": true,
        "response": true
      },
      {
        "client_id": "app-3",
        "success": true,
        "response": false
      }
    ]
  },
  "signature": "3045022100ab1234...def890",
  "timestamp": "2025-09-01T15:23:45Z"
}
```

## 分布式投票机制

### 核心特性

1. **M-of-N阈值投票**: 服务器自动配置所需通过票数
2. **自动目标发现**: 从服务器获取参与投票的节点列表
3. **并发投票处理**: 同时向所有目标应用发送投票请求
4. **完整响应收集**: 等待所有投票响应，提供详细的投票状态
5. **自动签名生成**: 投票通过后自动生成签名
6. **防循环机制**: 通过`is_forwarded`标记防止投票请求无限循环

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
          │   message,           │                      │
          │   signer_app_id,     │                      │
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
          │                      │                      │ (完整收集响应)
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
          │ 7. 最终响应           │                      │
          │ {                    │                      │
          │   success,           │                      │
          │   approved,          │                      │
          │   voting_results: {  │                      │
          │     vote_details[]   │                      │
          │   },                 │                      │
          │   signature          │                      │
          │ }                    │                      │
          │◄─────────────────────┤                      │
          │                      │                      │
          │ 8. 显示投票结果       │                      │
          │ 包括：               │                      │
          │ - 每个节点投票状态    │                      │
          │ - 最终签名结果        │                      │
          │                      │                      │
```

### 投票流程步骤

1. **发起投票**: 客户端向签名工具发送投票请求
2. **分发投票**: 签名工具向所有目标应用并发发送投票请求
3. **本地决策**: 每个应用根据自定义逻辑（如消息包含"test"）做出投票决定
4. **收集结果**: 等待所有目标应用响应投票
5. **阈值判断**: 检查是否满足所需投票数量
6. **生成签名**: 如果投票通过，使用签名者App ID生成最终签名

### 投票决策逻辑

当前实现的投票决策：
- 如果消息内容包含"test"（不区分大小写），则投票通过
- 可以通过修改`main.go`中的投票逻辑来自定义决策规则

### 架构优化

最新版本进行了以下优化：

1. **简化API**: 移除了`isForwarded`参数，由客户端库自动处理
2. **统一投票方法**: 只保留一个`VotingSign`方法
3. **正确的签名者**: 使用`signer_app_id`作为签名生成者，而非接收者
4. **清理数据结构**: 移除了不必要的字段如`TaskID`和`TotalParticipants`

## 支持的协议和曲线

### 协议
- `ecdsa`: 椭圆曲线数字签名算法
- `schnorr`: Schnorr签名方案

### 曲线
- `ed25519`: Edwards25519曲线
- `secp256k1`: SECP256K1曲线（比特币曲线）
- `secp256r1`: SECP256R1曲线（NIST P-256）

## 技术特性

### 无缓存部署
服务器设置了以下HTTP头来防止浏览器缓存：
- `Cache-Control: no-cache, no-store, must-revalidate`
- `Pragma: no-cache`
- `Expires: 0`

部署新版本后无需手动清除浏览器缓存。

### 静态文件服务
- 集成的静态文件服务器
- 支持HTML、CSS、JS、图片等文件类型
- 安全的路径验证，防止目录遍历攻击

## 安全考虑

1. **环境变量**: 敏感配置存储在环境变量中
2. **输入验证**: 所有输入在处理前都会被验证
3. **错误消息**: 错误消息不暴露敏感信息
4. **CORS**: 为Web界面访问启用了CORS
5. **防循环**: 投票请求包含防循环机制

## 故障排除

### 常见问题

1. **"APP_ID environment variable is required"**
   - 解决方案: 设置`APP_ID`环境变量

2. **"Failed to initialize TEE client"**
   - 解决方案: 确保TEE配置服务器正在运行且可访问

3. **投票请求失败**
   - 检查目标应用是否正在运行
   - 确认网络连通性和部署配置

4. **签名验证失败**
   - 确保消息和签名格式正确
   - 验证使用的App ID是否正确

### 日志
应用程序记录重要事件，包括：
- 服务器启动信息
- 投票流程详细日志
- API请求处理
- 错误条件和签名操作结果

## 开发指南

### 代码架构

项目采用模块化设计：
- **main.go**: HTTP服务器和API路由定义
- **types.go**: 请求/响应数据结构（简化后的结构）
- **crypto.go**: 签名验证和加密算法实现
- **server.go**: 静态文件服务和无缓存处理
- **voting.go**: 投票处理和自定义投票逻辑

### 依赖项
- `github.com/TEENet-io/tee-dao-key-management-client/go`: TEENet密钥管理客户端
- `github.com/gin-gonic/gin`: Web框架
- 标准Go加密库

### 自定义投票逻辑

可以通过修改`main.go`中的投票决策逻辑来自定义投票行为：

```go
// 当前逻辑：消息包含"test"则通过
localApproval := strings.Contains(strings.ToLower(messageStr), "test")

// 可以修改为其他逻辑，例如：
// - 基于消息长度
// - 基于时间戳
// - 基于消息哈希
// - 基于外部API调用
```

## 许可证

此项目是TEENet密钥管理客户端的一部分，遵循相同的许可条款。