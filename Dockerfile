
# ==================== Stage 1: Web (Bun + Vite) ====================
FROM oven/bun:1.1.29-alpine AS webbuilder
WORKDIR /app/web
ENV BUN_ENABLE_TELEMETRY=0

# 一些老的 alpine 镜像默认仓库会过期，先把仓库指向稳定版本并更新索引
RUN set -eux; \
    echo "https://dl-cdn.alpinelinux.org/alpine/v3.18/main" > /etc/apk/repositories; \
    echo "https://dl-cdn.alpinelinux.org/alpine/v3.18/community" >> /etc/apk/repositories; \
    apk update || \
    (echo "https://mirrors.aliyun.com/alpine/v3.18/main" > /etc/apk/repositories && \
     echo "https://mirrors.aliyun.com/alpine/v3.18/community" >> /etc/apk/repositories && \
     apk update) || \
    (echo "https://mirrors.tuna.tsinghua.edu.cn/alpine/v3.18/main" > /etc/apk/repositories && \
     echo "https://mirrors.tuna.tsinghua.edu.cn/alpine/v3.18/community" >> /etc/apk/repositories && \
     apk update) || true

# 构建所需工具：
# - build-base: 包含 make/gcc/g++
# - libc6-compat: 兼容层，避免 esbuild/sharp 等预编译二进制在 musl 下崩溃
RUN apk add --no-cache git python3 build-base bash curl ca-certificates || \
    (echo "primary repository failed, switching to aliyun mirror" && \
     echo "https://mirrors.aliyun.com/alpine/v3.18/main" > /etc/apk/repositories && \
     echo "https://mirrors.aliyun.com/alpine/v3.18/community" >> /etc/apk/repositories && \
     apk add --no-cache git python3 build-base bash curl ca-certificates) || \
    (echo "aliyun mirror failed, switching to tsinghua mirror" && \
     echo "https://mirrors.tuna.tsinghua.edu.cn/alpine/v3.18/main" > /etc/apk/repositories && \
     echo "https://mirrors.tuna.tsinghua.edu.cn/alpine/v3.18/community" >> /etc/apk/repositories && \
     apk add --no-cache git python3 build-base bash curl ca-certificates)

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
FROM golang:1.25-alpine AS gobuilder
WORKDIR /build

# 修复 alpine 镜像仓库（避免部分区域旧仓库失效）
RUN set -eux; \
    echo "https://dl-cdn.alpinelinux.org/alpine/v3.21/main" > /etc/apk/repositories; \
    echo "https://dl-cdn.alpinelinux.org/alpine/v3.21/community" >> /etc/apk/repositories; \
    apk update || \
    (echo "https://mirrors.aliyun.com/alpine/v3.21/main" > /etc/apk/repositories && \
     echo "https://mirrors.aliyun.com/alpine/v3.21/community" >> /etc/apk/repositories && \
     apk update) || \
    (echo "https://mirrors.tuna.tsinghua.edu.cn/alpine/v3.21/main" > /etc/apk/repositories && \
     echo "https://mirrors.tuna.tsinghua.edu.cn/alpine/v3.21/community" >> /etc/apk/repositories && \
     apk update) || true

# 设置构建环境变量
ARG GOPROXY_MIRRORS="https://proxy.golang.org|https://goproxy.cn,direct"
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=${GOPROXY_MIRRORS}

# 先拉 go 依赖（利于缓存）
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

# 修复 alpine 镜像仓库并安装必要包
RUN set -eux; \
    echo "https://dl-cdn.alpinelinux.org/alpine/v3.20/main" > /etc/apk/repositories; \
    echo "https://dl-cdn.alpinelinux.org/alpine/v3.20/community" >> /etc/apk/repositories; \
    apk update || \
    (echo "https://mirrors.aliyun.com/alpine/v3.20/main" > /etc/apk/repositories && \
     echo "https://mirrors.aliyun.com/alpine/v3.20/community" >> /etc/apk/repositories && \
     apk update) || \
    (echo "https://mirrors.tuna.tsinghua.edu.cn/alpine/v3.20/main" > /etc/apk/repositories && \
     echo "https://mirrors.tuna.tsinghua.edu.cn/alpine/v3.20/community" >> /etc/apk/repositories && \
     apk update) || true; \
    apk add --no-cache ca-certificates tzdata || \
    apk add --no-cache --repository=https://dl-cdn.alpinelinux.org/alpine/v3.20/main \
        --repository=https://dl-cdn.alpinelinux.org/alpine/v3.20/community \
        ca-certificates tzdata || \
    apk add --no-cache --repository=https://mirrors.aliyun.com/alpine/v3.20/main \
        --repository=https://mirrors.aliyun.com/alpine/v3.20/community \
        ca-certificates tzdata || \
    apk add --no-cache --repository=https://mirrors.tuna.tsinghua.edu.cn/alpine/v3.20/main \
        --repository=https://mirrors.tuna.tsinghua.edu.cn/alpine/v3.20/community \
        ca-certificates tzdata

RUN update-ca-certificates || true

# 拷贝二进制与前端静态资源
COPY --from=gobuilder  /out/nebulagate /app/nebulagate
COPY --from=webbuilder /app/web/dist  /app/public

# 数据工作目录（挂卷）
WORKDIR /data
EXPOSE 3000
ENTRYPOINT ["/app/nebulagate"]
