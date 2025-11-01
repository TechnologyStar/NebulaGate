# ==================== Stage 1: Web (Bun + Vite) ====================
FROM oven/bun:1.1.29 AS webbuilder
WORKDIR /app/web
ENV BUN_ENABLE_TELEMETRY=0

# Bun 镜像已经包含必要的构建工具

# 先复制 manifest（利于缓存），再装依赖
COPY web/package.json web/bun.lock* ./
RUN if [ -f bun.lock ] || [ -f bun.lockb ]; then bun install --frozen-lockfile; else bun install; fi

# 再复制剩余源码
COPY web/ ./

# 构建版本号：默认 dev，可被 --build-arg 覆盖
ARG VITE_REACT_APP_VERSION=dev
ENV VITE_REACT_APP_VERSION=${VITE_REACT_APP_VERSION}

# 打印 bun 版本与依赖树，便于定位问题
RUN echo "Bun version:" && bun --version
RUN bun pm ls || true

# 避免因内存不足导致构建中断（按 CI 机器内存大小调整）
ENV NODE_OPTIONS="--max-old-space-size=4096"

# 构建（优先使用 package.json 的 build 脚本，不行再直接用 vite）
RUN (bun run build --verbose || bun run build || bun x vite build --logLevel info)

# ==================== Stage 2: Go Builder（稳定版 Golang） ====================
# go.mod 指定了 Go 1.25.1；如果基础镜像版本过低，`go mod download` 会直接报错
FROM golang:1.25.1 AS gobuilder
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
FROM debian:12-slim
WORKDIR /app
ENV TZ=America/Chicago DEBIAN_FRONTEND=noninteractive
RUN set -eux; \
    apt-get update || (sed -i 's|deb.debian.org|mirrors.ustc.edu.cn|g' /etc/apt/sources.list && apt-get update); \
    apt-get install -y --no-install-recommends \
      ca-certificates \
      tzdata; \
    rm -rf /var/lib/apt/lists/*

# 拷贝二进制与前端静态资源
COPY --from=gobuilder  /out/nebulagate /app/nebulagate
COPY --from=webbuilder /app/web/dist  /app/public

# 数据工作目录（挂卷）
WORKDIR /data
EXPOSE 3000
ENTRYPOINT ["/app/nebulagate"]
