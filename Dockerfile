# ==================== Stage 1: Web (Bun + Vite) ====================
FROM oven/bun:1.1.29-alpine AS webbuilder
WORKDIR /app/web
ENV BUN_ENABLE_TELEMETRY=0

# 某些旧 alpine 源会过期，统一指向 latest-stable 并更新索引
RUN set -eux; \
    sed -i -E 's#https?://.*/alpine/v[0-9.]+/#https://dl-cdn.alpinelinux.org/alpine/latest-stable/#g' /etc/apk/repositories; \
    apk update || true

# 构建工具
RUN apk add --no-cache git python3 build-base bash curl ca-certificates || \
    (echo "primary repository failed, retrying with latest-stable mirror" && \
     apk add --no-cache --repository=https://dl-cdn.alpinelinux.org/alpine/latest-stable/main \
         --repository=https://dl-cdn.alpinelinux.org/alpine/latest-stable/community \
         git python3 build-base bash curl ca-certificates)

# 先复制 manifest（利于缓存）
COPY web/package.json web/bun.lock* ./
RUN if [ -f bun.lock ] || [ -f bun.lockb ]; then bun install --frozen-lockfile; else bun install; fi

# 再复制源码
COPY web/ ./

# 构建版本号：默认 dev，可被 --build-arg 覆盖
ARG VITE_REACT_APP_VERSION=dev
ENV VITE_REACT_APP_VERSION=${VITE_REACT_APP_VERSION}

# 打印 bun 版本与依赖树，便于排障
RUN echo "Bun version:" && bun --version
RUN bun pm ls || true

# 内存限制（按 CI 机器内存大小调整）
ENV NODE_OPTIONS="--max-old-space-size=4096"

# 构建前端
RUN (bun run build --verbose || bun run build || bun x vite build --logLevel info)

# ==================== Stage 2: Go Builder（稳定版 Golang） ====================
# 注：保持与 go.mod 的主/次版本一致。补丁版选用更高的补丁（兼容）。
FROM golang:1.25.3-alpine AS gobuilder
WORKDIR /build

# 可配置 CGO。默认 CGO=0 生成纯静态二进制；如依赖需要，可在构建时 --build-arg CGO_ENABLED=1
ARG CGO_ENABLED=0
ENV CGO_ENABLED=${CGO_ENABLED}
ENV GOOS=linux GOARCH=amd64
ENV GOPROXY=https://goproxy.cn,direct
# 更详细的编译输出，便于定位失败原因
ENV GOFLAGS="-v"

# 若启用 CGO，则安装 C 工具链
RUN if [ "${CGO_ENABLED}" = "1" ]; then \
      apk add --no-cache build-base ca-certificates; \
    else \
      apk add --no-cache ca-certificates; \
    fi

# 预拉依赖（缓存）
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# 拷贝后端源码 + 前端产物
COPY . .
COPY --from=webbuilder /app/web/dist ./web/dist

# 在 go:embed 编译前校验前端产物是否存在，给出清晰错误
RUN test -f ./web/dist/index.html || (echo "ERROR: web/dist/index.html 不存在，前端构建可能失败" && ls -la ./web || true && exit 1)

# 构建二进制（如 main 不在仓库根目录，替换 '.' 为实际主程序路径）
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags "-s -w" -o /out/nebulagate .

# ==================== Stage 3: Runtime ====================
FROM alpine:3.20
WORKDIR /app
ENV TZ=America/Chicago
RUN apk add --no-cache ca-certificates tzdata || \
    apk add --no-cache --repository=https://dl-cdn.alpinelinux.org/alpine/latest-stable/main ca-certificates tzdata
RUN update-ca-certificates || true

# 拷贝二进制与前端静态资源
COPY --from=gobuilder  /out/nebulagate /app/nebulagate
COPY --from=webbuilder /app/web/dist  /app/public

# 运行目录/端口
WORKDIR /data
EXPOSE 3000

ENTRYPOINT ["/app/nebulagate"]
