# 安全中心、计费功能和排行榜 - 快速参考指南

## 📋 文档导航

本目录包含关于New-API三大核心功能的详细文档：

### [完整报告](./SECURITY_BILLING_LEADERBOARD_REPORT.md)
包含安全中心、计费功能和排行榜的详细介绍、API文档、实现细节和最佳实践。

---

## 🔒 安全中心 - 快速查询

### 关键功能
| 功能 | 描述 | API端点 |
|------|------|--------|
| 获取仪表板 | 获取安全统计数据 | `GET /api/admin/security/dashboard` |
| 查询违规 | 分页查询违规记录 | `GET /api/admin/security/violations` |
| 查询用户 | 获取有违规的用户列表 | `GET /api/admin/security/users` |
| 封禁用户 | 封禁用户账户 | `POST /api/admin/security/users/{id}/ban` |
| 解除封禁 | 解除用户封禁 | `POST /api/admin/security/users/{id}/unban` |
| 设置重定向 | 设置用户模型重定向 | `POST /api/admin/security/users/{id}/redirect` |
| 获取设置 | 获取安全配置 | `GET /api/admin/security/settings` |
| 更新设置 | 更新安全配置 | `PUT /api/admin/security/settings` |

### 核心文件
- `service/security.go` - 业务逻辑
- `controller/security.go` - HTTP接口
- `model/security_violation.go` - 违规记录数据模型
- `model/user_security.go` - 用户安全状态

### 关键指标
- **ViolationCount**: 用户违规次数
- **IsBanned**: 是否被封禁
- **RedirectModel**: 重定向目标模型
- **LastViolationAt**: 最后违规时间

---

## 💰 计费功能 - 快速查询

### 关键概念
| 概念 | 说明 | 单位 |
|------|------|------|
| Quota | 账户余额 | 1 USD = 500,000 Quota |
| Token | 模型输入输出字符单位 | 根据模型计算 |
| Plan | 套餐订阅 | 周期性配额 |
| TopUp | 充值订单 | USD/本地货币 |

### 计费流程
```
1. 预扣费 (PreConsumeQuota)
   ├─ 检查余额
   └─ 冻结预估金额

2. 请求执行
   └─ 调用AI模型

3. 后扣费 (Post*ConsumeQuota)
   ├─ 统计实际消耗
   └─ 补扣或退款
```

### 主要API
| 功能 | API端点 | 方法 |
|------|--------|------|
| 获取充值配置 | `/api/topup/info` | GET |
| 创建充值订单 | `/api/topup/request/epay` | POST |
| 计算金额 | `/api/topup/request/amount` | POST |
| 查询充值记录 | `/api/topup` | GET |
| 管理员补单 | `/api/topup/admin/complete` | POST |
| Stripe支付 | `/api/stripe/checkout` | POST |
| 兑换优惠券 | `/api/voucher/redeem` | POST |

### 核心文件
- `service/token_counter.go` - Token统计
- `service/quota.go` - 额度扣费
- `service/billing_engine.go` - 新版计费引擎
- `controller/topup.go` - 充值接口
- `model/topup.go` - 充值订单模型
- `model/plan.go` - 套餐定义
- `model/plan_assignment.go` - 套餐分配

### 支付网关
- **易支付（EPay）**: 支付宝、微信、QQ钱包等
- **Stripe**: 国际信用卡支付

---

## 🏆 排行榜 - 快速查询

### 时间窗口
| 窗口 | 范围 |
|------|------|
| `1h` | 最近1小时 |
| `24h` | 最近24小时（默认） |
| `7d` | 最近7天 |
| `30d` | 最近30天 |
| `all` | 所有历史 |

### 排行指标
| 指标 | 说明 |
|------|------|
| RequestCount | 请求数 |
| TokenCount | Token总数 |
| QuotaConsumed | 消耗配额 |
| UniqueModels | 使用的不同模型数 |

### 主要API
| 功能 | API端点 | 方法 |
|------|--------|------|
| 获取排行榜 | `/api/leaderboard/users?window=24h&limit=100` | GET |
| 获取个人统计 | `/api/user/stats?window=24h` | GET |
| Token IP统计 | `/api/leaderboard/token-ips?token_id={id}` | GET |

### 核心文件
- `service/leaderboard.go` - 排行榜逻辑
- `controller/leaderboard.go` - HTTP接口

---

## 🔄 系统集成点

### 请求处理流程
```
用户请求
  ↓
安全检查 (middleware/security.go)
  ├─ 检查封禁
  └─ 检查重定向
  ↓
计费预授权
  ├─ 检查配额
  └─ 预扣费
  ↓
请求转发 (relay/)
  ↓
Token统计
  ↓
违规检查
  ├─ 检测违规
  └─ 记录日志
  ↓
计费提交
  ├─ 后扣费
  └─ 更新余额
  ↓
返回响应
```

---

## ⚙️ 配置速查

### 安全配置
```
SECURITY_AUTO_BAN_ENABLED=true         # 自动封禁开关
SECURITY_AUTO_BAN_THRESHOLD=10         # 封禁阈值（违规次数）
```

### 计费配置
```
QUOTA_DISPLAY_MODE=USD                 # USD/CNY/Tokens
QUOTA_PER_UNIT=500000                  # 1USD = 500,000 Quota
```

### 支付网关 - 易支付
```
OPERATION_SETTING_EPAY_ID=xxx          # 商户号
OPERATION_SETTING_EPAY_KEY=xxx         # 商户密钥
OPERATION_SETTING_PAY_ADDRESS=xxx      # 网关地址
```

### 支付网关 - Stripe
```
SETTING_STRIPE_API_SECRET=sk_xxx       # API密钥
SETTING_STRIPE_WEBHOOK_SECRET=wh_xxx   # Webhook密钥
SETTING_STRIPE_PRICE_ID=price_xxx      # 价格ID
```

---

## 🐛 常见问题快速解决

### 安全中心
- **用户被误封禁**: 使用解除封禁API恢复
- **重定向模型不生效**: 检查模型是否存在
- **违规记录堆积**: 定期清理历史记录

### 计费系统
- **充值失败**: 检查支付网关配置和签名
- **余额计算错误**: 查看RequestLog进行人工对账
- **配额不足**: 提示用户充值

### 排行榜
- **数据不更新**: 检查日志表是否有新数据
- **查询缓慢**: 添加索引，缩小时间范围
- **数据差异**: 验证日志表的数据完整性

---

## 📚 资源链接

### 源代码
- **安全中心**: `service/security.go`, `controller/security.go`
- **计费系统**: `service/billing_engine.go`, `controller/topup.go`
- **排行榜**: `service/leaderboard.go`, `controller/leaderboard.go`

### 数据模型
- `model/security_violation.go` - 违规记录
- `model/user_security.go` - 用户安全状态
- `model/topup.go` - 充值订单
- `model/plan.go` - 套餐
- `model/usage_counter.go` - 使用计数

### 前端
- `web/src/pages/TopUp/` - 充值页面
- `web/src/pages/Setting/Payment/` - 支付设置

---

## 📞 技术支持

遇到问题？请参考：
1. [完整报告](./SECURITY_BILLING_LEADERBOARD_REPORT.md) - 详细文档
2. 源代码中的注释
3. 测试文件（如存在）
4. Git日志和提交信息

---

**最后更新**: 2024年11月  
**维护者**: Development Team  
**版本**: 1.0
