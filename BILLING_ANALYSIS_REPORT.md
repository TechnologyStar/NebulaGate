# New-API 计费系统分析报告

## 概述

本报告详细分析了 new-api 项目中的计费系统实现，包括 Token 计费、积分包、套餐订阅、支付集成等核心功能。

---

## 1. Token 计费机制

### 1.1 Token 统计服务
**文件位置**: `service/token_counter.go`

**核心功能**:
- 支持多模型 Token 统计（OpenAI、Claude、Gemini、Realtime 等）
- 多媒体类型支持：文本、图片、音频、视频
- 图片 Token 计算：支持 tile-based 和 patch-based 两种算法
- 音频 Token 计算：基于采样率和格式

**关键函数**:
- `CountRequestToken()`: 请求阶段 Token 统计
- `CountTokenClaudeRequest()`: Claude 专用统计
- `CountTokenRealtime()`: 实时音频统计
- `getImageToken()`: 图片 Token 计算（支持 GPT-4o、GPT-5、O1 等模型的不同算法）

### 1.2 Quota 扣费逻辑
**文件位置**: `service/quota.go`, `service/pre_consume_quota.go`

**扣费流程**:
1. **预扣费阶段** (`PreConsumeQuota`):
   - 检查用户余额是否充足
   - 信任额度优化：高额度用户跳过预扣
   - 新版 Billing 模式下不预扣，仅做授权检查

2. **后扣费阶段** (`Post*ConsumeQuota`):
   - `PostClaudeConsumeQuota()`: Claude 模型
   - `PostAudioConsumeQuota()`: 音频模型
   - `PostWssConsumeQuota()`: WebSocket 实时连接
   - 计算实际消耗，补扣或退还预扣费

**计费公式**:
```
quota = (input_tokens + output_tokens * completion_ratio + 
         audio_tokens * audio_ratio) * model_ratio * group_ratio
```

**特殊处理**:
- 缓存 Token 折扣（Claude 提示词缓存）
- OpenRouter 成本反推
- 模型价格优先模式（UsePrice）

---

## 2. 积分包/套餐包功能

### 2.1 预付费积分包（充值）
**数据模型**: `model/topup.go`

```go
type TopUp struct {
    Id            int     // 订单ID
    UserId        int     // 用户ID
    Amount        int64   // 充值数量（美元）
    Money         float64 // 支付金额
    TradeNo       string  // 交易单号
    PaymentMethod string  // 支付方式
    CreateTime    int64   // 创建时间
    CompleteTime  int64   // 完成时间
    Status        string  // 状态：pending/success/failed
}
```

**充值流程**:
1. 前端选择充值金额和支付方式
2. 后端创建 TopUp 订单（`controller/topup.go`）
3. 调用支付网关生成支付链接
4. 用户完成支付
5. 支付网关回调通知
6. 验证签名后更新订单状态
7. 增加用户 quota：`user.Quota += amount * QuotaPerUnit`

**支持的功能**:
- 多规格金额选项（`AmountOptions`）
- 分组折扣（`AmountDiscount`）
- 分组倍率（`TopupGroupRatio`）
- 最小充值限制（`MinTopUp`）
- 充值记录查询（分页、搜索）
- 管理员补单功能

### 2.2 套餐订阅系统（Plan）
**数据模型**: `model/plan.go`, `model/plan_assignment.go`, `model/usage_counter.go`

#### Plan（套餐定义）
```go
type Plan struct {
    Code                   string    // 套餐代码
    Name                   string    // 套餐名称
    CycleType              string    // 周期类型：daily/monthly/custom
    CycleDurationDays      int       // 自定义周期天数
    QuotaMetric            string    // 额度指标：requests/tokens
    QuotaAmount            int64     // 额度数量
    AllowCarryOver         bool      // 是否允许结转
    CarryLimitPercent      int       // 结转限制百分比
    UpstreamAliasWhitelist JSONValue // 上游白名单
    IsActive               bool      // 是否激活
    IsPublic               bool      // 是否公开
}
```

#### PlanAssignment（套餐分配）
```go
type PlanAssignment struct {
    SubjectType         string    // 主体类型：user/token
    SubjectId           int       // 主体ID
    PlanId              int       // 套餐ID
    BillingMode         string    // 计费模式：plan/balance/fallback
    ActivatedAt         time.Time // 激活时间
    DeactivatedAt       *time.Time // 停用时间
    RolloverAmount      int64     // 结转额度
    RolloverExpiresAt   *time.Time // 结转过期时间
    AutoFallbackEnabled bool      // 自动回退到余额
    FallbackPlanId      *int      // 回退套餐ID
}
```

#### UsageCounter（使用量计数器）
```go
type UsageCounter struct {
    PlanAssignmentId int       // 套餐分配ID
    Metric           string    // 指标类型
    CycleStart       time.Time // 周期开始
    CycleEnd         time.Time // 周期结束
    ConsumedAmount   int64     // 已消耗额度
}
```

### 2.3 优惠券系统
**文件位置**: `service/voucher.go`, `model/voucher.go`

**功能特性**:
- 批量生成优惠券码
- 支持两种类型：
  - `credit`: 额度型（直接增加用户 quota）
  - `plan`: 套餐型（绑定套餐到用户）
- 限制机制：
  - 每批次最大兑换次数
  - 每用户最大兑换次数
  - 有效期限制
- 兑换防重：基于 code 的唯一性约束

**关键函数**:
- `GenerateVoucherBatch()`: 生成批次
- `RedeemVoucher()`: 兑换优惠券

---

## 3. 支付集成

### 3.1 易支付（EPay）
**文件位置**: `controller/topup.go`

**配置项**:
- `operation_setting.PayAddress`: 支付网关地址
- `operation_setting.EpayId`: 商户号
- `operation_setting.EpayKey`: 商户密钥

**支付方式**: 支付宝、微信支付、QQ 钱包等（可配置）

**关键流程**:
1. `RequestEpay()`: 创建支付订单
   - 计算应付金额（考虑倍率、折扣）
   - 生成唯一订单号
   - 调用支付网关生成支付链接
   
2. `EpayNotify()`: 支付回调
   - 验证签名
   - 订单级互斥锁防止并发
   - 更新订单状态
   - 增加用户额度
   - 记录日志

### 3.2 Stripe
**文件位置**: `controller/topup_stripe.go`, `model/topup.go`

**配置项**:
- `setting.StripeApiSecret`: API 密钥
- `setting.StripeWebhookSecret`: Webhook 密钥
- `setting.StripePriceId`: 价格ID
- `setting.StripeMinTopUp`: 最小充值金额

**功能特性**:
- Checkout Session 创建
- Webhook 事件处理
- 客户ID绑定（`User.StripeCustomer`）
- 订单幂等性保证

**处理函数**:
- `Recharge()`: 通用充值处理（支持 Stripe 和 EPay）
- `ManualCompleteTopUp()`: 管理员手动补单

---

## 4. Billing Engine（新版计费引擎）

### 4.1 架构设计
**文件位置**: `service/billing_engine.go`

**核心概念**:
- **PrepareCharge**: 预授权，确定使用套餐还是余额
- **CommitCharge**: 提交扣费，原子性执行
- **RollbackCharge**: 回滚扣费（失败重试）

**计费模式** (`common/billing.go`):
- `plan`: 套餐优先
- `balance`: 余额优先
- `fallback`: 套餐不足时自动切换余额
- `auto`: 自动决策

### 4.2 决策逻辑

```go
func (be *BillingEngine) PrepareCharge(ctx, subjectKey, relayInfo) (*PreparedCharge, error) {
    // 1. 解析主体（用户/Token）
    // 2. 查询活跃套餐分配
    // 3. 确定计费模式
    // 4. 计算可用额度
    // 5. 返回预授权结果
}

func (be *BillingEngine) CommitCharge(ctx, params) (*RequestLog, error) {
    // 事务内执行：
    // 1. 创建 RequestLog（幂等性保障）
    // 2. 根据模式扣费：
    //    - plan 模式：递增 UsageCounter
    //    - balance 模式：扣减 User.Quota 和 Token.RemainQuota
    // 3. 自动回退逻辑
    // 4. 返回审计日志
}
```

### 4.3 特性
- **原子性**: 数据库事务保证
- **幂等性**: RequestLog 的 request_id 唯一约束
- **并发安全**: 行级锁（FOR UPDATE）
- **审计完整**: RequestLog 记录所有扣费细节
- **灵活策略**: 支持套餐结转、自动回退、白名单等

---

## 5. 数据模型总览

### 5.1 用户相关
| 表 | 关键字段 | 说明 |
|----|---------|------|
| `users` | `quota`, `used_quota`, `request_count` | 用户余额和统计 |
| `users` | `group`, `stripe_customer` | 用户分组和支付关联 |
| `users` | `aff_code`, `aff_quota` | 邀请码和邀请额度 |

### 5.2 Token 相关
| 表 | 关键字段 | 说明 |
|----|---------|------|
| `tokens` | `remain_quota`, `used_quota` | Token 额度限制 |
| `tokens` | `unlimited_quota` | 无限额度标记 |
| `tokens` | `user_id` | 所属用户 |

### 5.3 充值相关
| 表 | 关键字段 | 说明 |
|----|---------|------|
| `topups` | `amount`, `money`, `trade_no` | 充值金额和订单号 |
| `topups` | `payment_method`, `status` | 支付方式和状态 |

### 5.4 套餐相关
| 表 | 关键字段 | 说明 |
|----|---------|------|
| `plans` | `quota_metric`, `quota_amount` | 额度类型和数量 |
| `plans` | `cycle_type`, `cycle_duration_days` | 周期类型 |
| `plan_assignments` | `subject_type`, `subject_id` | 套餐分配对象 |
| `plan_assignments` | `billing_mode`, `auto_fallback_enabled` | 计费模式 |
| `usage_counters` | `consumed_amount`, `cycle_start` | 周期内消耗 |

### 5.5 优惠券相关
| 表 | 关键字段 | 说明 |
|----|---------|------|
| `voucher_batches` | `grant_type`, `credit_amount` | 券类型和额度 |
| `voucher_batches` | `plan_grant_id`, `max_redemptions` | 套餐授予和限制 |
| `voucher_redemptions` | `code`, `subject_id` | 兑换记录 |

### 5.6 审计相关
| 表 | 关键字段 | 说明 |
|----|---------|------|
| `request_logs` | `request_id`, `amount`, `mode` | 请求扣费日志 |
| `logs` | `user_id`, `type`, `content` | 用户操作日志 |

---

## 6. API 端点总结

### 6.1 充值相关
| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/topup/info` | GET | 获取充值配置 |
| `/api/topup/request/epay` | POST | 发起易支付 |
| `/api/topup/request/amount` | POST | 计算支付金额 |
| `/api/topup` | GET | 查询充值记录 |
| `/api/topup/admin/complete` | POST | 管理员补单 |
| `/api/user/epay/notify` | POST | 易支付回调 |

### 6.2 Stripe 相关
| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/stripe/checkout` | POST | 创建 Checkout |
| `/api/stripe/webhook` | POST | Stripe 回调 |

### 6.3 优惠券相关
| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/voucher/redeem` | POST | 兑换优惠券 |
| `/api/voucher/generate` | POST | 生成优惠券（管理员） |

### 6.4 计费查询
| 端点 | 方法 | 说明 |
|------|------|------|
| `/v1/dashboard/billing/subscription` | GET | OpenAI 兼容查询额度 |
| `/v1/dashboard/billing/usage` | GET | OpenAI 兼容查询消耗 |

---

## 7. 前端实现

### 7.1 充值页面
**路径**: `web/src/components/topup/`, `web/src/pages/TopUp/`

**组件结构**:
- `index.jsx`: 主充值页面
- `RechargeCard.jsx`: 充值卡片（金额选择、支付方式）
- `InvitationCard.jsx`: 邀请码卡片
- `modals/PaymentConfirmModal.jsx`: 支付确认弹窗
- `modals/TopupHistoryModal.jsx`: 充值记录弹窗
- `modals/TransferModal.jsx`: 邀请额度转移

**功能特性**:
- 预设金额选项（可配置折扣）
- 自定义金额输入
- 多支付方式选择（Stripe、支付宝、微信等）
- 实时金额计算（考虑分组倍率）
- 优惠券兑换
- 充值记录查看
- 邀请码分享和奖励

### 7.2 设置页面
**路径**: `web/src/pages/Setting/`

**配置项**:
- `Payment/SettingsPaymentGateway.jsx`: 支付网关配置（易支付）
- `Payment/SettingsPaymentGatewayStripe.jsx`: Stripe 配置
- `Operation/SettingsGeneral.jsx`: 运营设置（充值倍率、最小金额等）

---

## 8. 配置与常量

### 8.1 全局配置
**文件**: `common/features.go`

```go
// Billing Feature Gate
BillingFeatureEnabled     = true  // 启用新版计费
BillingDefaultMode        = "balance"  // 默认计费模式
BillingAutoFallbackEnabled = false  // 自动回退
```

### 8.2 Quota 换算
**文件**: `common/constants.go`

```go
QuotaPerUnit = 500000.0  // 1 USD = 500000 quota
```

**展示模式** (`operation_setting`):
- `USD`: 以美元展示
- `CNY`: 以人民币展示（需配置汇率）
- `Tokens`: 以 Token 数量展示

### 8.3 模型倍率
**文件**: `setting/ratio_setting/`

**倍率类型**:
- `ModelRatio`: 模型基础倍率
- `CompletionRatio`: 输出补全倍率（通常 > 1）
- `AudioRatio`: 音频倍率
- `AudioCompletionRatio`: 音频输出倍率
- `GroupRatio`: 用户分组倍率
- `TopupGroupRatio`: 充值分组倍率

---

## 9. 核心流程图

### 9.1 传统 Quota 扣费流程
```
用户请求
    ↓
PreConsumeQuota（预扣费）
    ├─ 检查用户余额
    ├─ 检查 Token 余额
    └─ 扣除预估额度
    ↓
执行 AI 请求
    ↓
统计实际 Token 消耗
    ↓
Post*ConsumeQuota（后扣费）
    ├─ 计算实际消耗
    ├─ 补扣或退还差额
    ├─ 更新 UsedQuota
    └─ 记录消费日志
    ↓
返回响应
```

### 9.2 新版 Billing 扣费流程
```
用户请求
    ↓
PrepareCharge（预授权）
    ├─ 查询活跃套餐
    ├─ 检查套餐余额
    ├─ 确定计费模式
    └─ 返回预授权结果
    ↓
执行 AI 请求
    ↓
统计实际 Token 消耗
    ↓
CommitCharge（提交扣费）
    ├─ 创建 RequestLog（幂等）
    ├─ 根据模式扣费：
    │   ├─ plan: 递增 UsageCounter
    │   └─ balance: 扣减 Quota
    ├─ 自动回退逻辑
    └─ 返回审计日志
    ↓
返回响应
```

### 9.3 充值流程
```
用户选择金额和支付方式
    ↓
后端创建 TopUp 订单
    ↓
调用支付网关
    ↓
用户完成支付
    ↓
支付网关异步回调
    ↓
验证签名和订单状态
    ↓
更新订单为 success
    ↓
增加用户 Quota
    ↓
记录操作日志
    ↓
通知用户（可选）
```

---

## 10. 关键代码位置索引

### 10.1 Service 层
- `service/token_counter.go`: Token 统计
- `service/quota.go`: 额度扣费
- `service/pre_consume_quota.go`: 预扣费
- `service/billing_engine.go`: 新版计费引擎
- `service/billing_gate.go`: 计费网关
- `service/voucher.go`: 优惠券服务

### 10.2 Model 层
- `model/user.go`: 用户模型
- `model/token.go`: Token 模型
- `model/topup.go`: 充值模型
- `model/plan.go`: 套餐模型
- `model/plan_assignment.go`: 套餐分配
- `model/usage_counter.go`: 使用量计数器
- `model/voucher.go`: 优惠券模型
- `model/request_log.go`: 请求日志

### 10.3 Controller 层
- `controller/topup.go`: 充值接口
- `controller/topup_stripe.go`: Stripe 接口
- `controller/billing.go`: 计费查询接口
- `controller/relay.go`: AI 请求转发（集成计费）

### 10.4 前端
- `web/src/components/topup/`: 充值组件
- `web/src/pages/TopUp/`: 充值页面
- `web/src/pages/Setting/Payment/`: 支付设置

---

## 11. 总结与建议

### 11.1 系统优势
✅ **灵活性高**: 支持传统按量付费和现代订阅制双模式  
✅ **扩展性强**: 模块化设计，易于添加新支付方式和计费规则  
✅ **安全可靠**: 事务保证、幂等性设计、行级锁防并发  
✅ **用户体验**: 前端完整，支持多币种展示和灵活配置  
✅ **审计完善**: 详细的日志记录，便于对账和问题排查  

### 11.2 系统特点
- 支持**预付费积分包**（通过充值/优惠券）
- 支持**订阅套餐**（周期性额度分配）
- 支持**混合计费**（套餐+余额自动回退）
- 支持**多支付渠道**（Stripe、易支付等）
- 支持**分组差异化定价**（不同用户群不同价格）
- 支持**模型差异化计费**（不同模型不同倍率）

### 11.3 适用场景
- ✅ **B2C SaaS**: 个人用户按需充值
- ✅ **B2B 订阅**: 企业客户购买套餐
- ✅ **混合模式**: 套餐用户超量按需付费
- ✅ **渠道分销**: 通过优惠券批量分发

---

## 附录：重要常量与配置

### 计费相关
- `QuotaPerUnit = 500000`: 1 USD = 50万积分
- `QuotaForNewUser`: 新用户赠送额度
- `QuotaForInviter`: 邀请者奖励额度
- `QuotaForInvitee`: 被邀请者奖励额度

### 支付相关
- `MinTopUp`: 最小充值金额
- `Price`: 充值倍率
- `TopupGroupRatio`: 分组充值倍率
- `AmountOptions`: 预设充值金额
- `AmountDiscount`: 充值折扣配置

### 套餐相关
- `PlanCycleDaily / Monthly / Custom`: 套餐周期
- `PlanQuotaMetricRequests / Tokens`: 额度指标
- `RolloverPolicyNone / CarryAll / Cap`: 结转策略

### Billing 相关
- `BillingFeatureEnabled`: 启用新计费
- `BillingDefaultMode`: 默认计费模式
- `BillingAutoFallbackEnabled`: 自动回退

---

**报告生成时间**: 2025-01-XX  
**系统版本**: new-api (Go 1.25)  
**分析范围**: 计费、支付、套餐、优惠券全模块
