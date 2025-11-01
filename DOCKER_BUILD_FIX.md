# Docker Go 编译失败修复说明

## 问题原因

Docker buildx 在 Go 编译阶段失败的根本原因是：

1. **缺少 git 工具**：golang:1.25.1-alpine 镜像中没有包含 git，而 Go 模块下载需要 git 来克隆某些依赖包
2. **网络问题**：部分环境下访问 Go 模块代理和 Alpine 镜像仓库存在网络问题

## 修复措施

### 核心修复
在 Go Builder 阶段添加了 git 安装：

```dockerfile
# 安装 git 以支持 Go 模块下载
RUN apk add --no-cache git
```

### 其他改进
1. 保留了 Alpine 仓库优化，提高依赖安装成功率
2. 保持了 CGO_ENABLED=0 设置，因为 glebarez/sqlite 是纯 Go 实现，不需要 CGO
3. 保持了原有的多阶段构建结构，确保最终镜像体积小

## 验证方法

构建 Docker 镜像：
```bash
docker build -t nebulagate .
```

或者只构建 Go Builder 阶段进行测试：
```bash
docker build --target gobuilder -t nebulagate-builder .
```

## 技术细节

- Go 版本：1.25.1
- 基础镜像：golang:1.25.1-alpine
- 构建环境：CGO_ENABLED=0 (静态编译)
- 目标平台：linux/amd64

修复后，Go 编译阶段能够成功完成：
1. 使用 git 下载所有 Go 模块依赖
2. 编译生成静态二进制文件 nebulagate
3. 输出到 /out/nebulagate
