# 套餐与兑换码系统实现文档

## 概述

本文档描述了完整的套餐发行、兑换码管理和用户使用流程的实现，包括管理端和用户端的所有功能。

## 一、数据库模型

### 1. Package表（套餐定义）
- `id` - 主键
- `name` - 套餐名称
- `description` - 描述
- `token_quota` - token额度
- `model_scope` - 可用模型列表（JSON格式）
- `validity_days` - 有效期天数
- `price` - 价格
- `status` - 状态（1: 启用, 2: 禁用）
- `created_at`, `updated_at` - 时间戳

### 2. RedemptionCode表（兑换码）
- `id` - 主键
- `package_id` - 关联套餐ID
- `code` - 唯一兑换码字符串（16字符）
- `status` - 状态（1: 未使用, 2: 已使用, 3: 已作废）
- `used_by_user_id` - 使用者ID
- `used_at` - 使用时间
- `user_package_id` - 关联的用户积分包ID
- `created_at`, `updated_at` - 时间戳

### 3. UserPackage表（用户积分包）
- `id` - 主键
- `user_id` - 用户ID
- `package_id` - 来源套餐ID
- `token_quota` - 剩余token额度
- `initial_quota` - 初始额度（用于显示）
- `model_scope` - 可用模型列表（JSON格式）
- `expire_at` - 过期时间
- `status` - 状态（1: 有效, 2: 已用完, 3: 已过期）
- `created_at`, `updated_at` - 时间戳

## 二、后端实现

### 文件清单

#### 模型层 (model/)
- `package.go` - Package模型及CRUD操作
- `redemption_code.go` - RedemptionCode模型及CRUD操作
- `user_package.go` - UserPackage模型及CRUD操作
- `package_schema.go` - 数据库迁移schema注册
- `migrations/20250201_package_system.go` - 数据库迁移定义

#### 服务层 (service/)
- `package.go` - 套餐业务逻辑
- `redemption_code.go` - 兑换码生成、作废、兑换逻辑
- `user_package.go` - 用户积分包消费逻辑

#### 控制器层 (controller/)
- `package.go` - 套餐管理API端点
- `redemption_code.go` - 兑换码管理API端点
- `user_package.go` - 用户积分包API端点

### API端点

#### 管理端 - 套餐管理
- `GET /api/admin/packages` - 获取套餐列表
- `GET /api/admin/packages/:id` - 获取套餐详情
- `POST /api/admin/packages` - 创建套餐
- `PUT /api/admin/packages/:id` - 更新套餐
- `DELETE /api/admin/packages/:id` - 删除套餐

#### 管理端 - 兑换码管理
- `POST /api/admin/redemption-codes` - 批量生成兑换码
- `GET /api/admin/redemption-codes` - 获取兑换码列表
- `PUT /api/admin/redemption-codes/:id/revoke` - 作废兑换码
- `GET /api/admin/redemption-codes/export` - 导出兑换码（CSV）

#### 用户端 - 兑换和查看
- `POST /api/user/redeem` - 兑换码兑换
- `GET /api/user/packages` - 获取用户所有积分包
- `GET /api/user/packages/active` - 获取用户有效积分包

## 三、前端实现

### 文件清单

#### 页面组件 (pages/)
- `Package/index.jsx` - 套餐管理页面
- `RedemptionCode/index.jsx` - 兑换码管理页面
- `UserPackage/index.jsx` - 用户积分包列表页面

#### 业务组件 (components/package/)
- `PackageManagement.jsx` - 套餐管理组件
- `RedemptionCodeManagement.jsx` - 兑换码管理组件
- `UserPackageList.jsx` - 用户积分包列表组件

#### API服务 (services/)
- `packageService.js` - 套餐和兑换码相关API调用封装

### 路由配置

在 `App.jsx` 中添加了以下路由：

```javascript
/console/packages          # 套餐管理（管理员）
/console/redemption-codes  # 兑换码管理（管理员）
/console/user-packages     # 用户积分包（用户）
```

## 四、核心功能实现

### 1. 兑换码生成
- 使用16字符随机码（大写字母+数字）
- 确保唯一性
- 支持批量生成（最多500个）
- 事务保证

### 2. 用户兑换
- 验证兑换码有效性
- 检查套餐状态
- 创建用户积分包
- 设置过期时间
- 记录兑换日志
- 使用数据库事务保证原子性

### 3. 积分包过期处理
- 后台定时任务（每小时执行）
- 标记过期的积分包
- 仅在主节点执行

### 4. 模型权限控制
- 套餐可限制可用模型
- 用户积分包继承套餐的模型限制
- 空模型列表表示允许所有模型

### 5. 额度消费（预留接口）
`service/user_package.go` 中的 `ConsumePackageQuota` 函数提供了：
- 优先消费即将过期的积分包
- 验证模型权限
- 支持跨多个积分包扣费
- 事务保证

## 五、数据库迁移

系统使用GORM的AutoMigrate功能，在首次启动时自动创建表。

迁移版本：`20250201_package_system`

注册在：`model/package_schema.go`

## 六、常量定义

在 `common/constants.go` 中添加了：

```go
const (
    PackageStatusActive   = 1
    PackageStatusInactive = 2
)

const (
    RedemptionCodeStatusUnused   = 1
    RedemptionCodeStatusRedeemed = 2
    RedemptionCodeStatusRevoked  = 3
)

const (
    UserPackageStatusActive    = 1
    UserPackageStatusExhausted = 2
    UserPackageStatusExpired   = 3
)

const (
    KeyPaymentTypeBalance = "balance"
    KeyPaymentTypePackage = "package"
)
```

## 七、与现有系统集成点

### 预留集成功能

1. **Token创建扣费**
   - 需要在 `controller/token.go` 的 `AddToken` 中添加支付方式选择
   - 添加 `payment_type` 和 `user_package_id` 字段
   - 调用 `service/user_package.go` 的 `ConsumePackageQuota` 函数

2. **Relay层模型验证**
   - 在relay请求处理时验证Token对应的UserPackage的model_scope
   - 确保只有允许的模型可以被调用

3. **Billing系统集成**
   - 在billing dashboard中展示积分包使用情况
   - 记录积分包消费的详细日志

## 八、使用流程

### 管理员流程
1. 创建套餐（设置名称、额度、模型、有效期、价格）
2. 批量生成兑换码
3. 导出兑换码（CSV格式）
4. 分发兑换码给用户
5. 可以作废未使用的兑换码

### 用户流程
1. 获得兑换码
2. 在用户端输入兑换码兑换
3. 查看自己的积分包列表
4. 查看剩余额度和过期时间
5. （预留）创建Token时选择使用积分包

## 九、安全性考虑

1. **兑换码唯一性**：生成时验证，最多尝试10次
2. **事务保证**：兑换过程使用数据库事务，防止重复兑换
3. **行锁**：使用 `FOR UPDATE` 防止并发问题
4. **权限控制**：管理功能仅管理员可访问
5. **状态验证**：兑换时验证兑换码和套餐状态
6. **过期处理**：定时任务自动标记过期积分包

## 十、待完成功能

1. Token创建时的积分包支付方式选择
2. Relay层的模型权限验证
3. 积分包使用统计和分析
4. 多语言国际化支持
5. 更丰富的前端UI和数据展示
6. 积分包使用历史记录

## 十一、测试建议

1. 测试兑换码生成的唯一性
2. 测试并发兑换同一兑换码
3. 测试过期积分包的使用限制
4. 测试套餐删除的约束（有未使用兑换码时不能删除）
5. 测试CSV导出功能
6. 测试积分包消费的优先级（先到期先消费）
