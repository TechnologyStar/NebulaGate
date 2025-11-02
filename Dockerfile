# syntax=docker/dockerfile:1.7
FROM golang:1.22 AS build
WORKDIR /src

# 先拷贝依赖声明，充分利用缓存
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 再拷贝源码
COPY . .

# 创建输出目录；默认禁用 CGO 避免编译器缺失；按需改 GOOS/GOARCH
RUN --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p /out && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags "-s -w" -o /out/nebulagate ./cmd/nebulagate

# 极小运行时镜像（也可用 scratch）
FROM gcr.io/distroless/static-debian12 AS run
COPY --from=build /out/nebulagate /nebulagate
USER nonroot:nonroot
ENTRYPOINT ["/nebulagate"]
