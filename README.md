# NebulaGate（基于 new-api 的增强发行版）

> ⚠️ 本文档为你二开仓库的**新 README 模版**。为避免与上游重名/冲突，展示名统一为 **NebulaGate**（代号），仓库名仍可保留 `TechnologyStar/new-api`。如需完全避开冲突，可将镜像名/可执行文件名改为 `nebulagate`（示例已用可替换占位符）。

---

## 目录

* [背景与定位](#背景与定位)
* [NebulaGate 相比上游的差异与优势](#nebulagate-相比上游的差异与优势)
* [特性总览](#特性总览)
* [系统架构](#系统架构)
* [兼容性与支持矩阵](#兼容性与支持矩阵)
* [快速开始](#快速开始)

  * [方式 A：Docker 一条命令](#方式-aDocker-一条命令)
  * [方式 B：Docker Compose（推荐）](#方式-bdocker-compose推荐)
  * [方式 C：原生二进制 + systemd](#方式-c原生二进制--systemd)
* [配置与环境变量](#配置与环境变量)
* [管理后台与运营流程](#管理后台与运营流程)
* [计费域模型（Fork 新增）](#计费域模型fork-新增)
* [治理与分流（Fork 新增）](#治理与分流fork-新增)
* [升级/迁移/备份](#升级迁移备份)
* [安全加固建议](#安全加固建议)
* [性能调优指南](#性能调优指南)
* [常见问题 FAQ](#常见问题-faq)
* [Roadmap](#roadmap)
* [License](#license)

---

## 背景与定位

**NebulaGate** 是在上游 `new-api` 基础上增强的 **LLM 网关 & AI 资产管理系统**，面向个人到中小企业的**统一接入、计费结算、治理审计、可视化运营**场景。

它保留了上游的**多模型聚合、OpenAI/Claude/Gemini/Responses/Realtime/Rerank** 等能力，并在此之上新增：**企业计费引擎、套餐与代金券、用量生命周期调度、治理与分流链路、可视化运营后台**等模块，实现从“接入 → 使用 → 计费 → 审计”的闭环。

---

## NebulaGate 相比上游的差异与优势

> 下表聚焦你在 Fork 中新增/增强的关键能力。

| 能力                                              | 上游 new-api | NebulaGate（本 Fork）                                      |
| ----------------------------------------------- | ---------- | ------------------------------------------------------- |
| 多模型聚合 / OpenAI 兼容                               | ✅          | ✅（沿用）                                                   |
| Claude / Gemini / Responses / Realtime / Rerank | ✅          | ✅（沿用）                                                   |
| Token/渠道/权重/看板                                  | ✅          | ✅（沿用）                                                   |
| 充值（易支付/Stripe）                                  | ✅          | ✅（沿用）                                                   |
| **企业计费域模型（Plan / Balance / Ledger / 幂等）**       | ❌（基础或无）    | **✅ 新增：可审计的账本、计划周期与幂等保障**                               |
| **套餐（Plans）及用户关联**                              | ❌          | **✅ 新增：可将计划分配到用户/组织**                                   |
| **代金券（Vouchers）发放与兑换**                          | ❌          | **✅ 新增：促销/补偿/风控通道**                                     |
| **用量生命周期调度（Scheduler）**                         | ❌          | **✅ 新增：周期重置/结转/到期提醒**                                   |
| **治理（Governance）与分流链路**                         | ❌          | **✅ 新增：检测 → 分流/重路由 → 审计日志**                             |
| **运营后台增强（计费/治理 UI）**                            | ❌          | **✅ 新增：Plan/Voucher/治理配置页面**                            |
| **加密与多机一致性**                                    | 基础配置       | **增强建议**：`SESSION_SECRET`、共享 Redis 时 `CRYPTO_SECRET` 必配 |

> 结论：NebulaGate 将上游“聚合网关 + 基础运营”提升为“**商用计费 + 治理审计 + 生命周期运维**”的一体化解决方案。

---

## 特性总览

### 继承自上游

* OpenAI 兼容中继：Chat Completions、Responses、Streaming、Realtime
* Claude/Gemini 兼容、Rerank 支持
* 渠道管理、加权随机、故障切换
* Token/令牌管理与分组
* 易支付/Stripe 充值
* 数据看板与运营指标

### 本 Fork 新增/强化

* **Billing Engine**：Plan、Balance、Ledger（审计日志）、幂等保证
* **套餐与授权**：按用户/组织分配计划与额度
* **Voucher**：代金券批量发放、兑换码校验、可配置有效期/额度/适用范围
* **Scheduler**：计划周期重置（如月度）、结转策略、到期通知（预留 Hook）
* **Governance Pipeline**：请求治理检测（关键词/策略/地域等）→ 分流/重路由 → 审计记录
* **运营后台**：新增计费/代金券/治理的 GUI 管理页

---

## 系统架构

```
           +------------------+            +----------------+
Client --> |  NebulaGate API  |--Routing-->|  Provider SDKs |
(Web/SDK)  |  (Gateway Core)  |            |  (OpenAI/...)  |
           +---------+--------+            +-------+--------+
                     |                             |
                     | Billing & Governance        |
                     v                             v
               +-----+------+               +------+------+
               |  Billing   |<--Ledger----->|  Database  |
               |  (Plans/   |               | (SQLite/   |
               |  Vouchers) |               |  MySQL/PG) |
               +-----+------+               +------+------+
                     |                             ^
                     | Scheduler                    |
                     v                             |
               +-----+------+               +------+------+
               |  Redis     |<--Cache/Crypto Keys|  Admin UI  |
               +------------+               +-------------+
```

---

## 兼容性与支持矩阵

* **接口兼容**：默认兼容上游的主要路由与参数（OpenAI/Claude/Gemini/Responses/Realtime/Rerank）。
* **数据库**：SQLite（试用/单机）、MySQL ≥5.7.8、PostgreSQL ≥9.6。
* **缓存**：Redis（建议生产开启；如多机共享，需配 `CRYPTO_SECRET`）。
* **时区**：建议设置 `TZ=Asia/Shanghai` 或你的本地时区。

---

## 快速开始

> 下面给出 3 种方式，镜像名使用占位：`YOUR_DOCKER_IMAGE/nebulagate:latest`。如你沿用原镜像，可替换为原有镜像名；如需完全去重命名，请在 CI/CD 中推送到你的镜像仓库。

### 方式 A：Docker 一条命令

```bash
# 数据持久化目录（用于 SQLite 或本地文件）
mkdir -p /opt/nebulagate-data

# SQLite 直跑（试用/小流量）
docker run --name nebulagate -d --restart always -p 3000:3000 \
  -e TZ=Asia/Shanghai \
  -e SESSION_SECRET='请改成强随机值' \
  -v /opt/nebulagate-data:/data \
  YOUR_DOCKER_IMAGE/nebulagate:latest

# 如使用 MySQL/PG，追加 SQL_DSN
# -e SQL_DSN='user:pass@tcp(mysql:3306)/oneapi?charset=utf8mb4&parseTime=True&loc=Local'
```

### 方式 B：Docker Compose（推荐）

`docker-compose.yml` 示例（已去重命名，便于与你线上环境并存）：

```yaml
version: "3.9"
services:
  nebulagate:
    image: YOUR_DOCKER_IMAGE/nebulagate:latest  # 替换为你的镜像名
    container_name: nebulagate
    restart: always
    ports:
      - "3000:3000"
    env_file:
      - .env
    environment:
      - TZ=Asia/Shanghai
      - SESSION_SECRET=${SESSION_SECRET}
      - SQL_DSN=${SQL_DSN:-}
      - REDIS_CONN_STRING=${REDIS_CONN_STRING:-}
      - CRYPTO_SECRET=${CRYPTO_SECRET:-}
    volumes:
      - ./data:/data
    depends_on:
      - redis
      - mysql

  redis:
    image: redis:7-alpine
    command: ["redis-server", "--save", "", "--appendonly", "no"]
    volumes:
      - ./redis:/data
    restart: always

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: example
      MYSQL_DATABASE: oneapi
      TZ: Asia/Shanghai
    command: ["--character-set-server=utf8mb4", "--collation-server=utf8mb4_unicode_ci"]
    volumes:
      - ./mysql:/var/lib/mysql
    restart: always
```

`.env` 示例：

```dotenv
SESSION_SECRET=请改成强随机值
# MySQL:
SQL_DSN=root:example@tcp(mysql:3306)/oneapi?charset=utf8mb4&parseTime=True&loc=Local
# Redis（可选/建议生产开启）
REDIS_CONN_STRING=redis://@redis:6379/0
# 如启用共享 Redis，请务必设置：
CRYPTO_SECRET=请改成强随机值
```

启动：

```bash
docker compose up -d
```

### 方式 C：原生二进制 + systemd

```bash
# 编译
make build || go build -o nebulagate ./...

# 运行目录
sudo mkdir -p /opt/nebulagate /var/lib/nebulagate
sudo cp nebulagate /opt/nebulagate/

# 环境变量（/etc/default/nebulagate）
sudo tee /etc/default/nebulagate >/dev/null <<'EOF'
SESSION_SECRET=强随机
SQL_DSN=
REDIS_CONN_STRING=
CRYPTO_SECRET=
TZ=Asia/Shanghai
EOF

# systemd 单元（/etc/systemd/system/nebulagate.service）
[Unit]
Description=NebulaGate Service
After=network.target

[Service]
EnvironmentFile=/etc/default/nebulagate
ExecStart=/opt/nebulagate/nebulagate
User=nobody
Restart=always

[Install]
WantedBy=multi-user.target

# 启动
sudo systemctl daemon-reload
sudo systemctl enable --now nebulagate
```

---

## 配置与环境变量

> 仅列常用项；其余保持与上游一致。

| 变量                          | 说明            | 建议                      |
| --------------------------- | ------------- | ----------------------- |
| `SESSION_SECRET`            | 会话/签名密钥，多机必配  | 生产必填，强随机                |
| `CRYPTO_SECRET`             | 共享 Redis 加密密钥 | 启用共享 Redis 时必填          |
| `SQL_DSN`                   | MySQL/PG 连接串  | 生产建议使用 MySQL/PG         |
| `REDIS_CONN_STRING`         | Redis 连接串     | 生产建议开启缓存与治理存储           |
| `TZ`                        | 时区            | `Asia/Shanghai` 或你的本地时区 |
| `ERROR_LOG_ENABLED`         | 错误日志          | 生产建议开启                  |
| `STREAMING_TIMEOUT`         | 流式超时          | 视链路情况调整                 |
| `AZURE_DEFAULT_API_VERSION` | Azure 兼容      | 如使用 Azure 路由时配置         |

---

## 管理后台与运营流程

1. **渠道**：接入上游（OpenAI/Claude/Gemini/Azure/DeepSeek…），配置权重/重试策略。
2. **令牌**：为调用方生成 API Key，可按组/用户管理。
3. **套餐（本 Fork 新增）**：创建 Plan（额度/周期/结转策略），分配给用户/组织。
4. **代金券（本 Fork 新增）**：批量发放兑换码，设定有效期/额度/适用范围，发给用户自助兑换。
5. **治理（本 Fork 新增）**：启用关键词/地域/模型/路径等策略，违规或命中条件时**分流/重路由**并记录审计日志。
6. **看板**：观察用量、成功率、时延、失败原因；结合账本（Ledger）做对账与审计。

---

## 计费域模型（Fork 新增）

* **Plan**：定义额度/周期（如月度）与结转策略（可关/可开上限）。
* **Balance**：用户/组织的可用余额（额度/金额）。
* **Ledger**：不可篡改的审计账本，记录变更（充值、消耗、退款、结转）。
* **Idempotency**：通过幂等键避免重复记账；失败回滚与补偿流程可配置。
* **调度**：周期重置、结转与到期通知由 Scheduler 触发，可对接外部通知（WebHook/Email/SMS）。

> 目标：将“调用量 → 成本 → 收入/额度”端到端闭环，支持促销/补偿/风控等商业化运营。

---

## 治理与分流（Fork 新增）

* **策略**：关键词/地域/模型/路径/速率等；命中后执行 **阻断/降级/改路由**。
* **分流/重路由**：按策略将请求改发到其它渠道或模型，降低失败率与合规风险。
* **审计**：将命中记录写入 Ledger/审计日志，配合看板追踪。

---

## 升级/迁移/备份

* **Docker/Compose 升级**：

  ```bash
  docker pull YOUR_DOCKER_IMAGE/nebulagate:latest && docker restart nebulagate
  # 或
  docker compose pull && docker compose up -d
  ```
* **数据库迁移**：随版本提供自动迁移（如包含 `migrate` 步骤请在发布说明注明）。
* **备份**：

  * SQLite：备份挂载目录 `./data`（冷备期间停服或使用文件级快照）。
  * MySQL/PG：使用常规备份策略（mysqldump/pg_dump + binlog/WAL）。

---

## 安全加固建议

* 强制设置 `SESSION_SECRET`、共享 Redis 时 `CRYPTO_SECRET`。
* 通过反向代理（Nginx/Caddy）启用 **HTTPS/HTTP2**，仅暴露 80/443。
* 将网关端口置于内网，外网只暴露代理；开启 WAF/限速（如 Nginx `limit_req`）。
* 最小权限：容器与数据库账号分权、只读文件系统（可选）。
* 日志脱敏：对密钥/账号信息做掩码。

---

## 性能调优指南

* **连接池**：数据库与 Redis 连接池大小按核心数/负载调优。
* **重试与超时**：为各渠道设置合理超时与重试上限，避免雪崩。
* **权重与分流**：结合失败率/时延动态调整渠道权重；命中治理策略时优先路由到稳定模型。
* **缓存**：开启 Redis 缓存/元数据缓存，降低 DB 压力。

---

## 常见问题 FAQ

**Q: 与上游 new-api 有何不同？**
A: 本 Fork 在其基础上新增 **计费引擎、套餐/代金券、生命周期调度、治理与分流、运营后台增强**，面向商业化与合规运营。

**Q: 如何避免与线上已部署的 new-api 冲突？**
A: 使用本 README 的去重命名：可执行文件 `nebulagate`、容器名 `nebulagate`、镜像 `YOUR_DOCKER_IMAGE/nebulagate`、端口/卷路径独立。

**Q: 生产部署建议？**
A: 使用 Compose + MySQL/PG + Redis；务必配置 `SESSION_SECRET`/`CRYPTO_SECRET`，置于反代之后并启用 HTTPS；定期备份。

**Q: 升级是否破坏已有数据？**
A: 遵循语义化发布；涉及模型/表结构变更会在 Release Note 提供迁移指引。

---

## Roadmap

* [ ] Voucher 使用统计与分布式风控规则
* [ ] 更细粒度的配额策略（按模型/路径）
* [ ] 审计日志导出与 SIEM 对接
* [ ] 多租户组织视图与对账报表

---

## License

* 继承上游协议（在此处放置具体 LICENSE 信息与版权说明）。
