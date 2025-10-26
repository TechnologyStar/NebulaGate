# ==================== Stage 1: Web (Bun + Vite) ====================
FROM oven/bun:1.1.29-alpine AS webbuilder
WORKDIR /app/web
ENV BUN_ENABLE_TELEMETRY=0

# 许多前端依赖在 Alpine/musl 需要原生编译工具，否则 bun/vite 构建会直接失败
RUN apk add --no-cache git python3 make g++ bash

# 为了稳妥，直接复制整个 web 目录（牺牲一点缓存命中，换取对缺失 lock 文件的兼容）
COPY web/ ./

# 有 bun.lockb 就用 --frozen-lockfile；没有就普通安装
RUN set -eux; \
    if [ -f bun.lockb ]; then \
      bun install --frozen-lockfile; \
    else \
      bun install; \
    fi

# 注入版本：优先 --build-arg，再尝试 /app/VERSION，最后回落 dev
# 不强制 COPY VERSION（避免流水线里没有该文件导致构建直接失败）
ARG VITE_REACT_APP_VERSION
ENV VITE_REACT_APP_VERSION=${VITE_REACT_APP_VERSION}

RUN set -eux; \
  echo "Bun version: $(bun --version)"; \
  FINAL="${VITE_REACT_APP_VERSION:-}"; \
  if [ -z "${FINAL:-}" ] && [ -f /app/VERSION ]; then FINAL="$(cat /app/VERSION || true)"; fi; \
  : "${FINAL:=dev}"; \
  export VITE_REACT_APP_VERSION="$FINAL"; \
  echo "VITE_REACT_APP_VERSION=${VITE_REACT_APP_VERSION}"; \
  # 避免因内存不足引起的构建中断（按 CI 机器内存调整）
  export NODE_OPTIONS="--max-old-space-size=2048"; \
  # 打印依赖树帮助排错（不影响构建）
  bun pm ls || true; \
  # 直接调用 vite，打开详细日志，遇到真实错误能第一时间看到
  if bun x vite --version >/dev/null 2>&1; then \
    bun x vite build --logLevel info; \
  else \
    # 如果你的脚本里有 "build": "vite build"，这两行也能跑
    bun run build --verbose || bun run build; \
  fi

# ==================== Stage 2: Go Builder（稳定版 Golang） ====================
# 你之前用的是 golang:1.25-alpine，目前更稳的是 1.23 系列（除非你确实需要 1.25）
FROM golang:1.23-alpine AS gobuilder
WORKDIR /build
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# 先拉 go 依赖（利用缓存）
COPY go.mod go.sum ./
RUN go mod download

# 拷贝后端源码 + 前端产物
COPY . .
COPY --from=webbuilder /app/web/dist ./web/dist

# 构建二进制（如果 main 不在仓库根目录，把 "." 改成你的主程序路径）
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
