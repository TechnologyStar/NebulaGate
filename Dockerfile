# ==================== Stage 1: Web (Bun + Vite) ====================
FROM oven/bun:1.1-alpine AS webbuilder
WORKDIR /app/web
ENV BUN_ENABLE_TELEMETRY=0

# 1) 先复制清单文件，安装依赖（利用缓存）
COPY web/package.json web/bun.lockb* ./
RUN if [ -f bun.lockb ]; then \
      bun install --frozen-lockfile; \
    else \
      rm -f bun.lock; \
      bun install; \
    fi

# 2) 再复制源码（避免可选文件 COPY 失败）
COPY web/ ./

# 3) 复制版本文件（可没有）
COPY VERSION /app/VERSION

# 4) 解析版本变量：--build-arg > /app/VERSION > dev
ARG VITE_REACT_APP_VERSION
RUN set -eux; \
  echo "Bun version: $(bun --version)"; \
  FINAL_VERSION="$(cat /app/VERSION 2>/dev/null || true)"; \
  if [ -n "${VITE_REACT_APP_VERSION:-}" ]; then FINAL_VERSION="$VITE_REACT_APP_VERSION"; fi; \
  if [ -z "${FINAL_VERSION:-}" ]; then FINAL_VERSION="dev"; fi; \
  export VITE_REACT_APP_VERSION="$FINAL_VERSION"; \
  echo "VITE_REACT_APP_VERSION=${VITE_REACT_APP_VERSION}"; \
  # 可选：打印 build 脚本，确认存在
  node -e "console.log(require('./package.json').scripts?.build || 'NO_BUILD_SCRIPT')" ; \
  # 5) 进行构建（这里依赖 package.json 的 \"build\" 脚本）
  bun run build
