# 端到端加密与Heimdall安全网关全面升级指南

本文档详细说明了NewAPI的端到端加密系统、Heimdall安全网关和异常检测机制的实现与使用。

## 目录

1. [系统架构概述](#系统架构概述)
2. [对话历史端到端加密系统](#对话历史端到端加密系统)
3. [Heimdall安全网关](#heimdall安全网关)
4. [异常检测与防滥用机制](#异常检测与防滥用机制)
5. [部署指南](#部署指南)
6. [API文档](#api文档)
7. [前端集成](#前端集成)
8. [安全最佳实践](#安全最佳实践)

## 系统架构概述

```
Client Application
       ↓
  Heimdall Gateway (TLS/HTTPS)
       ↓ (Encrypted)
    NewAPI Backend
       ↓
  ┌────┴────┬──────────┬───────────┐
  │         │          │           │
Database  Redis   AI Models   Anomaly
(Encrypted)              Detector
```

### 核心组件

1. **NewAPI Go Backend**: 主API服务器，处理业务逻辑、加密/解密、用户管理
2. **Heimdall Python Gateway**: 安全网关，提供TLS加密、请求日志和流量分析
3. **加密服务**: AES-256-GCM端到端加密，保护对话历史
4. **异常检测引擎**: 行为分析、风险评分、自动防滥用

## 对话历史端到端加密系统

### 特性

- **用户密钥管理**: 每个用户生成唯一的32字节AES-256密钥
- **不可重置**: 密钥一旦生成，不可重置（防止管理员访问）
- **服务器端加密存储**: 密钥使用PBKDF2+SHA256哈希后存储
- **Key级别控制**: 用户可为每个API密钥启用/禁用对话日志记录
- **客户端解密**: 只有持有密钥的用户可以解密对话历史

### 数据库模型

#### User表扩展
```sql
ALTER TABLE users ADD COLUMN encryption_key_hash VARCHAR(255);
ALTER TABLE users ADD COLUMN encryption_enabled BOOLEAN DEFAULT FALSE;
```

#### Token表扩展
```sql
ALTER TABLE tokens ADD COLUMN conversation_logging_enabled BOOLEAN DEFAULT FALSE;
```

#### ConversationLog表（新建）
```sql
CREATE TABLE conversation_logs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    token_id INT NOT NULL,
    model VARCHAR(255),
    encrypted_data TEXT NOT NULL,
    nonce VARCHAR(64) NOT NULL,
    timestamp BIGINT NOT NULL,
    request_id VARCHAR(64),
    message_count INT DEFAULT 0,
    prompt_tokens INT DEFAULT 0,
    completion_tokens INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_token_id (token_id),
    INDEX idx_timestamp (timestamp)
);
```

### 加密流程

1. **密钥生成**
   ```
   用户请求 → 生成32字节随机密钥 → PBKDF2哈希 → 存储哈希值 → 返回明文密钥（仅一次）
   ```

2. **启用加密**
   ```
   用户提供密钥 → 验证哈希 → 设置encryption_enabled=true
   ```

3. **对话记录**
   ```
   API请求 → 提取对话数据 → AES-256-GCM加密 → 存储密文+nonce
   ```

4. **对话查询**
   ```
   用户请求+密钥 → 验证密钥 → 获取密文 → AES-256-GCM解密 → 返回明文
   ```

### API端点

#### 1. 生成加密密钥
```http
POST /api/encryption/generate-key
Authorization: Bearer <user_token>

Response:
{
  "success": true,
  "message": "Encryption key generated successfully. Please save it securely - it cannot be recovered!",
  "data": {
    "encryption_key": "64个字符的hex密钥"
  }
}
```

#### 2. 启用/禁用加密
```http
POST /api/encryption/enable
Authorization: Bearer <user_token>
Content-Type: application/json

{
  "encryption_key": "用户的加密密钥",
  "enable": true
}

Response:
{
  "success": true,
  "message": "Encryption enabled successfully"
}
```

#### 3. 查询加密状态
```http
GET /api/encryption/status
Authorization: Bearer <user_token>

Response:
{
  "success": true,
  "data": {
    "encryption_enabled": true,
    "has_encryption_key": true
  }
}
```

#### 4. 创建加密对话日志
```http
POST /api/encryption/conversation-log
Authorization: Bearer <user_token>
X-Encryption-Key: <用户加密密钥>
Content-Type: application/json

{
  "token_id": 123,
  "model": "gpt-4",
  "conversation_data": "{JSON格式的对话数据}",
  "prompt_tokens": 100,
  "completion_tokens": 150,
  "message_count": 2,
  "request_id": "req_xxx"
}

Response:
{
  "success": true,
  "message": "Conversation log saved successfully",
  "data": {
    "id": 456
  }
}
```

#### 5. 获取加密对话日志
```http
GET /api/encryption/conversation-logs?page=1&page_size=20&token_id=123
Authorization: Bearer <user_token>
X-Encryption-Key: <用户加密密钥>

Response:
{
  "success": true,
  "data": {
    "logs": [
      {
        "id": 456,
        "token_id": 123,
        "model": "gpt-4",
        "conversation_data": "解密后的对话数据",
        "timestamp": 1701234567,
        "request_id": "req_xxx",
        "message_count": 2,
        "prompt_tokens": 100,
        "completion_tokens": 150
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 20
  }
}
```

#### 6. 删除对话日志
```http
DELETE /api/encryption/conversation-log/:id
Authorization: Bearer <user_token>

Response:
{
  "success": true,
  "message": "Conversation log deleted successfully"
}
```

## Heimdall安全网关

### 核心功能

1. **TLS/HTTPS加密**: 保护客户端到网关的通信
2. **请求日志记录**: 完整捕获请求元数据
3. **透明代理**: 无缝转发请求到NewAPI
4. **数据采集**: IP、设备指纹、API调用频率统计

### 部署Heimdall

#### 方法1: 直接运行
```bash
cd heimdall
pip install -r requirements.txt

# 配置环境变量
export NEWAPI_BASE_URL="http://localhost:3000"
export HEIMDALL_PORT="8000"
export TLS_CERT_PATH="/path/to/cert.pem"  # 可选
export TLS_KEY_PATH="/path/to/key.pem"    # 可选

# 运行
python main.py
```

#### 方法2: Docker
```bash
docker build -t heimdall ./heimdall
docker run -d \
  -p 8000:8000 \
  -e NEWAPI_BASE_URL=http://newapi:3000 \
  -e TLS_CERT_PATH=/certs/cert.pem \
  -e TLS_KEY_PATH=/certs/key.pem \
  -v /path/to/certs:/certs:ro \
  heimdall
```

#### 方法3: Docker Compose
```bash
# 编辑 .env 文件
TLS_CERT_PATH=/path/to/cert.pem
TLS_KEY_PATH=/path/to/key.pem

# 启动
docker-compose -f docker-compose.heimdall.yml up -d
```

### TLS证书设置

#### 生成自签名证书（测试用）
```bash
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes \
  -subj "/CN=localhost"
```

#### 使用Let's Encrypt（生产用）
```bash
# 安装certbot
apt-get install certbot

# 获取证书
certbot certonly --standalone -d api.yourdomain.com

# 证书路径
export TLS_CERT_PATH="/etc/letsencrypt/live/api.yourdomain.com/fullchain.pem"
export TLS_KEY_PATH="/etc/letsencrypt/live/api.yourdomain.com/privkey.pem"
```

### Nginx反向代理配置

```nginx
upstream heimdall {
    server localhost:8000;
}

server {
    listen 80;
    server_name api.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://heimdall;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
    }
}
```

### 数据采集字段

Heimdall采集以下数据并发送到NewAPI：

| 字段 | 描述 | 来源 |
|------|------|------|
| user_id | 用户ID | 从token解析 |
| token_key | API密钥 | Authorization header |
| request_path | 请求路径 | URL |
| request_method | HTTP方法 | GET/POST/etc |
| real_ip | 真实IP | X-Forwarded-For（第一个值） |
| forwarded_for | 转发链 | X-Forwarded-For完整值 |
| user_agent | 用户代理 | User-Agent header |
| request_headers | 请求头 | 完整headers (JSON) |
| request_body | 请求体 | Body (限制10KB) |
| content_fingerprint | 内容指纹 | MD5哈希 |
| device_fingerprint | 设备指纹 | SHA256哈希 |
| cookies | Cookie | Cookie header (JSON) |
| response_status | 响应状态码 | HTTP status |
| response_time | 响应时间 | 毫秒 |
| timestamp | 时间戳 | Unix时间戳 |

## 异常检测与防滥用机制

### 检测算法

#### 1. 设备聚合分析
- 查询同一设备指纹的所有请求
- 计算访问频率和时间间隔
- 识别异常行为模式

#### 2. IP聚合分析
- 考虑NAT网关共享IP的情况
- 降低IP风险权重（相比设备）
- 追踪IP信誉历史

#### 3. 行为模式分析
异常类型包括：

- **high_frequency**: 高频率请求（平均间隔<1秒）
- **data_spike**: 数据突变（访问量/登录次数 > 50）
- **suspicious_pattern**: 可疑行为模式
- **no_api_activity**: 高访问但无API调用

### 风险评分算法

风险分数 (0-100) 计算公式：

```
风险分数 = 访问比例权重(40%) + 无API活动权重(30%) + 高频率权重(20%) + 访问量权重(10%)

其中：
- 访问比例权重 = min(40, 40 * (访问量/登录次数) / 阈值)
- 无API活动权重 = 访问量 > 100 且 API调用 = 0 ? 30 : 0
- 高频率权重 = 平均间隔 < 1秒 ? 20 * (1 - 平均间隔) : 0
- 访问量权重 = min(10, 10 * (访问量 / 最小阈值 - 1))
```

### 触发动作

根据风险分数自动触发：

| 风险分数 | 动作 | 描述 |
|----------|------|------|
| < 70 | none | 无动作 |
| 70-79 | alert | 发送警报 |
| 80-89 | rate_limit | 限流 |
| ≥ 90 | block | 阻止访问 |

### 异常检测API

#### 1. 获取异常检测列表
```http
GET /api/anomaly/detections?page=1&page_size=20&status=detected&min_risk_score=70
Authorization: Bearer <user_token>

Response:
{
  "success": true,
  "data": {
    "anomalies": [
      {
        "id": 1,
        "user_id": 123,
        "device_fingerprint": "abc123...",
        "ip_address": "1.2.3.4",
        "anomaly_type": "high_frequency",
        "risk_score": 85.5,
        "login_count": 5,
        "total_access_count": 500,
        "access_ratio": 100.0,
        "api_request_count": 0,
        "time_window_start": 1701234567,
        "time_window_end": 1701238167,
        "average_interval": 0.5,
        "status": "detected",
        "action": "rate_limit",
        "description": "High frequency requests: avg interval 0.50 seconds",
        "detected_at": 1701238167
      }
    ],
    "total": 10,
    "page": 1,
    "page_size": 20
  }
}
```

#### 2. 获取异常详情
```http
GET /api/anomaly/detections/:id
Authorization: Bearer <user_token>

Response:
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 123,
    ...
  }
}
```

#### 3. 更新异常状态（管理员）
```http
PUT /api/anomaly/detections/:id/status
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "status": "resolved"
}

可选状态: "detected", "reviewing", "resolved", "false_positive"

Response:
{
  "success": true,
  "message": "Anomaly status updated successfully"
}
```

#### 4. 更新异常动作（管理员）
```http
PUT /api/anomaly/detections/:id/action
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "action": "block"
}

可选动作: "none", "alert", "rate_limit", "block"

Response:
{
  "success": true,
  "message": "Anomaly action updated successfully"
}
```

#### 5. 获取异常统计
```http
GET /api/anomaly/statistics?start_time=1701234567&end_time=1701238167
Authorization: Bearer <user_token>

Response:
{
  "success": true,
  "data": {
    "total_anomalies": 25,
    "high_risk_count": 5,
    "avg_risk_score": 72.3
  }
}
```

#### 6. 设备聚合查询（管理员）
```http
GET /api/anomaly/device-aggregation?device_fingerprint=abc123...&page=1&page_size=20
Authorization: Bearer <admin_token>

Response:
{
  "success": true,
  "data": {
    "logs": [...],           // 该设备的所有请求日志
    "logs_total": 500,
    "anomalies": [...],      // 该设备的异常记录
    "anomaly_total": 3,
    "page": 1,
    "page_size": 20
  }
}
```

#### 7. IP聚合查询（管理员）
```http
GET /api/anomaly/ip-aggregation?ip=1.2.3.4&page=1&page_size=20
Authorization: Bearer <admin_token>

Response:
{
  "success": true,
  "data": {
    "logs": [...],           // 该IP的所有请求日志
    "logs_total": 1000,
    "anomalies": [...],      // 该IP的异常记录
    "anomaly_total": 5,
    "page": 1,
    "page_size": 20
  }
}
```

#### 8. 手动触发异常检测（管理员）
```http
POST /api/anomaly/trigger/:user_id
Authorization: Bearer <admin_token>

Response:
{
  "success": true,
  "message": "Anomaly detection completed successfully"
}
```

## 部署指南

### 完整部署流程

#### 1. 准备环境
```bash
# 克隆仓库
git clone <repository>
cd new-api

# 安装Go依赖
go mod download

# 构建Go后端
go build -o new-api

# 安装Python依赖（Heimdall）
cd heimdall
pip install -r requirements.txt
cd ..
```

#### 2. 配置数据库
```bash
# 设置数据库连接
export SQL_DSN="mysql://user:pass@localhost:3306/newapi"

# 运行数据库迁移
./new-api migrate
```

#### 3. 配置TLS证书
```bash
# 生成或获取TLS证书
export TLS_CERT_PATH="/path/to/cert.pem"
export TLS_KEY_PATH="/path/to/key.pem"
```

#### 4. 启动服务
```bash
# 启动NewAPI
./new-api &

# 启动Heimdall
cd heimdall
export NEWAPI_BASE_URL="http://localhost:3000"
python main.py &
```

#### 5. 使用Docker Compose（推荐）
```bash
# 配置.env文件
cat > .env << EOF
SQL_DSN=mysql://user:pass@mysql:3306/newapi
REDIS_CONN_STRING=redis://redis:6379
SESSION_SECRET=your-secret-key
TLS_CERT_PATH=/path/to/cert.pem
TLS_KEY_PATH=/path/to/key.pem
EOF

# 启动所有服务
docker-compose -f docker-compose.heimdall.yml up -d
```

### 监控与维护

#### 查看日志
```bash
# NewAPI日志
docker logs newapi

# Heimdall日志
docker logs heimdall

# 或直接查看文件
tail -f heimdall/heimdall.log
```

#### 数据库备份
```bash
# 备份加密密钥哈希（重要）
mysqldump -u user -p newapi users --where="encryption_key_hash IS NOT NULL" > encryption_keys_backup.sql

# 备份对话日志
mysqldump -u user -p newapi conversation_logs > conversation_logs_backup.sql

# 备份异常检测数据
mysqldump -u user -p newapi anomaly_detections heimdall_logs > security_logs_backup.sql
```

## 前端集成

### 1. 加密密钥管理页面

```javascript
// 生成加密密钥
async function generateEncryptionKey() {
  const response = await fetch('/api/encryption/generate-key', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${userToken}`
    }
  });
  
  const data = await response.json();
  if (data.success) {
    // 显示密钥，提示用户保存
    alert(`您的加密密钥: ${data.data.encryption_key}\n\n请妥善保存！此密钥不可恢复！`);
    
    // 提供下载功能
    downloadKey(data.data.encryption_key);
  }
}

// 启用加密
async function enableEncryption(encryptionKey) {
  const response = await fetch('/api/encryption/enable', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${userToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      encryption_key: encryptionKey,
      enable: true
    })
  });
  
  const data = await response.json();
  if (data.success) {
    alert('加密已启用！');
  }
}
```

### 2. 对话日志查看页面

```javascript
// 获取加密对话日志
async function fetchConversationLogs(encryptionKey, page = 1) {
  const response = await fetch(
    `/api/encryption/conversation-logs?page=${page}&page_size=20`,
    {
      headers: {
        'Authorization': `Bearer ${userToken}`,
        'X-Encryption-Key': encryptionKey
      }
    }
  );
  
  const data = await response.json();
  if (data.success) {
    displayLogs(data.data.logs);
  }
}

// 显示对话日志
function displayLogs(logs) {
  logs.forEach(log => {
    const conversationData = JSON.parse(log.conversation_data);
    console.log('Model:', log.model);
    console.log('Messages:', conversationData.messages);
    console.log('Timestamp:', new Date(log.timestamp * 1000));
  });
}
```

### 3. 安全中心仪表板

```javascript
// 获取异常检测统计
async function fetchAnomalyStatistics() {
  const endTime = Math.floor(Date.now() / 1000);
  const startTime = endTime - 86400; // 最近24小时
  
  const response = await fetch(
    `/api/anomaly/statistics?start_time=${startTime}&end_time=${endTime}`,
    {
      headers: {
        'Authorization': `Bearer ${userToken}`
      }
    }
  );
  
  const data = await response.json();
  if (data.success) {
    displayStatistics(data.data);
  }
}

// 获取异常列表
async function fetchAnomalies(page = 1, minRiskScore = 70) {
  const response = await fetch(
    `/api/anomaly/detections?page=${page}&page_size=20&min_risk_score=${minRiskScore}`,
    {
      headers: {
        'Authorization': `Bearer ${userToken}`
      }
    }
  );
  
  const data = await response.json();
  if (data.success) {
    displayAnomalies(data.data.anomalies);
  }
}
```

### 4. Token创建时的加密选项

```javascript
// 创建Token时添加对话日志选项
async function createToken(tokenData) {
  const response = await fetch('/api/token', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${userToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      name: tokenData.name,
      remain_quota: tokenData.quota,
      conversation_logging_enabled: tokenData.enableLogging, // 新增字段
      // ... 其他字段
    })
  });
  
  const data = await response.json();
  return data;
}
```

## 安全最佳实践

### 1. 密钥管理

✅ **应该做**:
- 提示用户立即保存密钥到安全位置
- 提供密钥下载功能（加密文件）
- 建议用户使用密码管理器
- 实施密钥验证流程

❌ **不应该做**:
- 在浏览器localStorage存储明文密钥
- 通过不安全渠道传输密钥
- 提供密钥重置功能（会破坏端到端加密）

### 2. TLS/HTTPS

✅ **应该做**:
- 生产环境必须使用TLS
- 使用Let's Encrypt免费证书
- 配置strong cipher suites
- 启用HSTS（HTTP Strict Transport Security）
- 定期更新证书

### 3. 日志管理

✅ **应该做**:
- 定期清理旧日志（实施retention policy）
- 限制请求体大小（默认10KB）
- 敏感数据脱敏
- 实施访问控制

❌ **不应该做**:
- 记录完整的密码或API密钥
- 无限制存储日志
- 向非授权用户暴露日志

### 4. 异常检测

✅ **应该做**:
- 定期审查异常报告
- 调整风险阈值以减少误报
- 对高风险异常及时响应
- 记录处理决策

❌ **不应该做**:
- 完全依赖自动化阻止
- 忽略false positives
- 设置过低的阈值（造成正常用户体验下降）

### 5. 数据保护

✅ **应该做**:
- 定期备份加密数据
- 实施多层安全防护
- 加密数据库连接
- 使用强密码和2FA

### 6. 合规性

- **GDPR**: 用户可删除自己的对话日志
- **数据驻留**: 可配置数据存储位置
- **访问审计**: 记录所有访问操作
- **数据最小化**: 只收集必要的数据

## 故障排除

### 问题1: 无法生成加密密钥
```
错误: Failed to generate encryption key
解决: 检查系统随机数生成器是否可用，确保有足够的熵
```

### 问题2: 解密失败
```
错误: Decryption failed
可能原因:
1. 密钥不正确
2. 数据已损坏
3. 密钥已更换（不支持）
解决: 验证用户提供的密钥是否正确
```

### 问题3: Heimdall连接NewAPI失败
```
错误: Connection refused
解决:
1. 检查NEWAPI_BASE_URL配置
2. 确保NewAPI正在运行
3. 检查网络连通性
4. 查看防火墙规则
```

### 问题4: TLS证书错误
```
错误: Certificate verification failed
解决:
1. 检查证书有效期
2. 验证证书域名匹配
3. 确保证书链完整
4. 检查文件路径和权限
```

### 问题5: 异常检测产生过多误报
```
解决:
1. 调整风险阈值（提高HighRiskScoreThreshold）
2. 增加MinAccessCountThreshold
3. 调整时间窗口大小
4. 审查并标记false positives
```

## 性能优化

### 1. 数据库索引
确保以下索引存在：
```sql
-- Conversation logs
CREATE INDEX idx_conversation_logs_user_timestamp ON conversation_logs(user_id, timestamp);
CREATE INDEX idx_conversation_logs_token_timestamp ON conversation_logs(token_id, timestamp);

-- Heimdall logs
CREATE INDEX idx_heimdall_logs_user_timestamp ON heimdall_logs(user_id, timestamp);
CREATE INDEX idx_heimdall_logs_device ON heimdall_logs(device_fingerprint);
CREATE INDEX idx_heimdall_logs_ip ON heimdall_logs(real_ip);

-- Anomaly detections
CREATE INDEX idx_anomaly_detections_user_score ON anomaly_detections(user_id, risk_score);
CREATE INDEX idx_anomaly_detections_status ON anomaly_detections(status);
```

### 2. 缓存策略
- 使用Redis缓存用户加密状态
- 缓存频繁访问的异常检测结果
- 实施Token缓存策略

### 3. 批量处理
- 批量写入日志（减少数据库压力）
- 异步处理异常检测
- 定期归档旧数据

## 总结

本系统实现了三大核心功能：

1. **端到端加密**: 保护用户对话历史，管理员无法访问
2. **Heimdall网关**: 提供TLS加密、全面日志记录和流量分析
3. **异常检测**: 自动识别和防御滥用行为

通过这些功能，NewAPI达到了最高级别的安全性和可靠性。

## 支持

如有问题，请：
1. 查阅本文档
2. 检查GitHub Issues
3. 联系技术支持

## 更新日志

- 2025-02-07: 初始版本发布
  - 端到端加密系统
  - Heimdall安全网关
  - 异常检测引擎
