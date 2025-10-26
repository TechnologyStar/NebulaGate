# ==================== Stage 1: Web (Bun + Vite) ====================
FROM oven/bun:1.1-alpine AS webbuilder
WORKDIR /app/web
ENV BUN_ENABLE_TELEMETRY=0

# 先复制清单文件并安装依赖（最大化缓存命中）
COPY web/package.json web/bun.lockb* ./
RUN if [ -f bun.lockb ]; then \
      bun install --frozen-lockfile; \
    else \
      rm -f bun.lock; \
      bun install; \
    fi

# 再复制源码
COPY web/ ./

# 如仓库根目录存在 VERSION，则复制（若没有此文件，请删除下一行）
COPY VERSION /app/VERSION

# 版本注入：--build-arg > /app/VERSION > dev
ARG VITE_REACT_APP_VERSION
RUN set -eux; \
  echo "Bun version: $(bun --version)"; \
  FINAL_VERSION="$(cat /app/VERSION 2>/dev/null || true)"; \
  if [ -n "${VITE_REACT_APP_VERSION:-}" ]; then FINAL_VERSION="$VITE_REACT_APP_VERSION"; fi; \
  if [ -z "${FINAL_VERSION:-}" ]; then FINAL_VERSION="dev"; fi; \
  export VITE_REACT_APP_VERSION="$FINAL_VERSION"; \
  echo "VITE_REACT_APP_VERSION=${VITE_REACT_APP_VERSION}"; \
  bun run build

# ==================== Stage 2: Go Builder (Go 1.25) ====================
FROM golang:1.25-alpine AS gobuilder
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
