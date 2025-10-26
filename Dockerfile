# ==================== Stage 1: Web (Bun + Vite) ====================
FROM oven/bun:1.1-alpine AS webbuilder
WORKDIR /app/web
ENV BUN_ENABLE_TELEMETRY=0

# 复制整个 web 目录，避免可选文件 COPY 失败
COPY web/ .

# 有 bun.lockb 就严格安装；否则删除误命名的 bun.lock 再普通安装
RUN if [ -f bun.lockb ]; then \
      bun install --frozen-lockfile; \
    else \
      rm -f bun.lock; \
      bun install; \
    fi

# 版本注入（可选；没有 VERSION 文件就回退为 dev）
COPY VERSION /app/VERSION
ARG VITE_REACT_APP_VERSION=dev
ENV VITE_REACT_APP_VERSION=${VITE_REACT_APP_VERSION}
RUN export VITE_REACT_APP_VERSION="$(cat /app/VERSION 2>/dev/null || echo ${VITE_REACT_APP_VERSION})" \
 && bun run build

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
