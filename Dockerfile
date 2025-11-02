# ==================== Stage 1: Web (Bun + Vite) ====================
FROM oven/bun:1.1.29-alpine AS webbuilder
WORKDIR /app/web
ENV BUN_ENABLE_TELEMETRY=0

# 避免旧源过期：统一切到 latest-stable；失败时忽略（后续还有重试）
RUN set -eux; \
    sed -i -E 's#https?://.*/alpine/v[0-9.]+/#https://dl-cdn.alpinelinux.org/alpine/latest-stable/#g' /etc/apk/repositories; \
    apk update || true

# 构建工具
RUN apk add --no-cache git python3 build-base bash curl ca-certificates || \
    (echo "primary repository failed, retrying with latest-stable mirror" && \
     apk add --no-cache --repository=https://dl-cdn.alpinelinux.org/alpine/latest-stable/main \
         --repository=https://dl-cdn.alpinelinux.org/alpine/latest-stable/community \
         git python3 build-base bash curl ca-certificates)

# 先复制清单（利用缓存）
COPY web/package.json web/bun.lock* ./
RUN if [ -f bun.lock ] || [ -f bun.lockb ]; then bun install --frozen-lockfile; else bun install; fi

# 再复制源码
COPY web/ ./

# 版本号可由 --build-arg 覆盖
ARG VITE_REACT_APP_VERSION=dev
ENV VITE_REACT_APP_VERSION=${VITE_REACT_APP_VERSION}

# 打印 bun 版本与依赖树，便于排障
RUN echo "Bun version:" && bun --version
RUN bun pm ls || true

# 适当放大可用内存（按你的 CI 机器内存调整）
ENV NODE_OPTIONS="--max-old-space-size=4096"

# 构建前端
RUN (bun run build --verbose || bun run build || bun x vite build --logLevel info)

# 保障有产物（否则 go:embed 会在下一阶段编译时报错）
RUN test -f /app/web/dist/index.html || (echo "ERROR: web/dist/index.html not found. Frontend build failed." && ls -la /app/web/dist || true && exit 1)


# ==================== Stage 2: Go Builder（与 go.mod 对齐的稳定版 Go） ====================
# go.mod 建议为：go 1.22
FROM golang:1.22-alpine AS gobuilder
WORKDIR /build

# 可切换 CGO：默认 0 生成纯静态；若依赖需要 cgo，构建时传 --build-arg CGO_ENABLED=1
ARG CGO_ENABLED=0
ENV CGO_ENABLED=${CGO_ENABLED}
ENV GOOS=linux GOARCH=amd64
ENV GOPROXY=https://goproxy.cn,direct
ENV GOFLAGS="-v"

# 基础包：git（拉取模块），必要时装 C 工具链
RUN set -eux; \
    apk add --no-cache ca-certificates git; \
    if [ "${CGO_ENABLED}" = "1" ]; then apk add --no-cache build-base; fi

# 先拉 go 依赖（缓存友好）
COPY go.mod go.sum ./
RUN go mod download

# 拷贝后端源码 + 前端产物
COPY . .
COPY --from=webbuilder /app/web/dist ./web/dist

# 再次保证 dist 存在，减少“黑盒”失败
RUN test -f ./web/dist/index.html || (echo "ERROR: ./web/dist/index.html missing in gobuilder stage." && ls -la ./web || true && exit 1)

# 构建二进制（主程序在仓库根目录）
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
