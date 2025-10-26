# ==================== Stage 1: Web (Bun + Vite) ====================
FROM oven/bun:1.1.29-alpine AS webbuilder
WORKDIR /app/web
ENV BUN_ENABLE_TELEMETRY=0

# Tooling for native addons + glibc compat for Alpine
RUN apk add --no-cache git python3 make g++ bash libc6-compat

# (Better cache) install deps first if lockfile exists
COPY web/package.json web/bun.lockb* ./

# If you always have a lock, keep only the frozen path; otherwise fall back.
RUN if [ -f bun.lockb ]; then bun install --frozen-lockfile; else bun install; fi

# Now copy the rest of the sources
COPY web/ ./

# Build args / envs — default to dev if not provided
ARG VITE_REACT_APP_VERSION=dev
ENV VITE_REACT_APP_VERSION=${VITE_REACT_APP_VERSION}

# Print versions and dependency tree explicitly
RUN echo "Bun version:" && bun --version
RUN bun pm ls || true

# Avoid OOMs in CI; tune as needed
ENV NODE_OPTIONS="--max-old-space-size=2048"

# Prefer your package.json script; otherwise use vite directly
# (split into separate steps so an error shows the real message)
RUN if bun x vite --version >/dev/null 2>&1; then \
      echo "vite found via bunx"; \
    else \
      echo "vite not found via bunx; will rely on package.json scripts"; \
    fi

RUN --mount=type=cache,target=/root/.bun \
    --mount=type=cache,target=/app/web/node_modules \
    (bun run build --verbose || bun run build || bun x vite build --logLevel info)

# ==================== Stage 2: Go Builder（稳定版 Golang） ====================
# go.mod 指定了 Go 1.25.1；如果基础镜像版本过低，`go mod download` 会直接报错
FROM golang:1.25.1-alpine AS gobuilder
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
