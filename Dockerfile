# ==================== Stage 1: Web (Bun + Vite) ====================
FROM oven/bun:1.1.29-alpine AS webbuilder
WORKDIR /app/web
ENV BUN_ENABLE_TELEMETRY=0

# 安装常见原生依赖构建所需的工具（某些 npm 包在 Alpine/musl 下会需要）
RUN apk add --no-cache git python3 make g++ bash

# 先复制清单文件并安装依赖（最大化缓存命中）
# 注意：不要使用通配符复制 bun.lockb；如果没有 lock 文件，也能继续
COPY web/package.json ./
# 如果你确实有 bun.lockb，则取消下一行注释
# COPY web/bun.lockb ./
RUN set -eux; \
    if [ -f bun.lockb ]; then \
      bun install --frozen-lockfile; \
    else \
      bun install; \
    fi

# 再复制源码
COPY web/ ./

# 如仓库根目录存在 VERSION，则复制（若没有也不报错）
# 利用 BuildKit 的“软复制”手法：不存在就不复制
# （如果你的构建环境没有启用 BuildKit，可保持你原来的方式并确保有 VERSION）
# 这里用 try-copy 的方式替代；若你的构建器不支持，保留你原 COPY 行并确保文件存在
RUN [ -f /app/VERSION ] || true
# 改为：把根目录 VERSION“按需地”注入，构建时可不用此行
# （标准 Dockerfile 没有条件 COPY，只能以下面这行二选一）
# COPY VERSION /app/VERSION

# 版本注入：--build-arg > /app/VERSION > dev
ARG VITE_REACT_APP_VERSION
# 将值写入环境，确保 Vite 能读到（vite 只会暴露以 VITE_ 开头的变量给客户端）
ENV VITE_REACT_APP_VERSION=${VITE_REACT_APP_VERSION}
# 如果未通过 ARG 提供，则尝试从 /app/VERSION 读取，否则为 dev
# 同时提升日志级别，便于排错
RUN set -eux; \
  echo "Bun version: $(bun --version)"; \
  FINAL_VERSION="${VITE_REACT_APP_VERSION:-}"; \
  if [ -z "${FINAL_VERSION:-}" ] && [ -f /app/VERSION ]; then FINAL_VERSION="$(cat /app/VERSION || true)"; fi; \
  if [ -z "${FINAL_VERSION:-}" ]; then FINAL_VERSION="dev"; fi; \
  export VITE_REACT_APP_VERSION="$FINAL_VERSION"; \
  echo "VITE_REACT_APP_VERSION=${VITE_REACT_APP_VERSION}"; \
  # 避免内存不够导致构建中断（按需增减）
  export NODE_OPTIONS="--max-old-space-size=2048"; \
  # 打印依赖树，辅助定位依赖问题
  bun pm ls || true; \
  # 显式使用 vite build 并打开详细日志
  if [ -f package.json ] && bun x vite --version >/dev/null 2>&1; then \
    bun x vite build --logLevel info; \
  else \
    bun run build --silent || bun run build; \
  fi

# ==================== Stage 2: Go Builder (Go 1.23+) ====================
# 注意：Go 1.25 目前并非稳定标签，建议使用 1.23（如需特定版本请自行调整）
FROM golang:1.23-alpine AS gobuilder
WORKDIR /build
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# 先拉 go 依赖（利用缓存）
COPY go.mod go.sum ./
RUN go mod download

# 拷贝后端源码 + 前端产物
COPY . .
COPY --from=webbuilder /app/web/dist ./web/dist

# 构建二进制（如果 main 不在仓库根目录，把 "." 改成主程序路径）
RUN go build -trimpath -ldflags "-s -w" -o /out/nebulagate .

# ==================== Stage 3: Runtime ====================
FROM alpine:3.20
WORKDIR /app
ENV TZ=America/Chicago
RUN apk add --no-cache ca-certificates tzdata && update-ca-certificates

# 拷贝二进制与前端静态资源
COPY --from=gobuilder  /out/nebulagate /app/nebulagate
COPY --from=webbuilder /app/web/dist  /app/public

# 数据工作目录（挂卷）
WORKDIR /data
EXPOSE 3000
ENTRYPOINT ["/app/nebulagate"]
