# syntax=docker/dockerfile:1.7
FROM golang:1.25 AS build
WORKDIR /src

# 拷贝所有源码
COPY . .

# 创建输出目录；默认禁用 CGO 避免编译器缺失；按需改 GOOS/GOARCH
# 使用多个代理源以提高成功率，设置超时和重试
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p /out && \
    GOPROXY=https://goproxy.cn,direct,https://proxy.golang.org,direct \
    GOSUMDB=off \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags "-s -w" -o /out/nebulagate .

# 极小运行时镜像（也可用 scratch）
FROM gcr.io/distroless/static-debian12 AS run
COPY --from=build /out/nebulagate /nebulagate
USER nonroot:nonroot
ENTRYPOINT ["/nebulagate"]
