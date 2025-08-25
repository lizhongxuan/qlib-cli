# 使用官方Go镜像作为构建环境
FROM golang:1.22.0-alpine AS builder

# 设置工作目录
WORKDIR /app

# 设置代理（可选，用于加速Go模块下载）
ENV GOPROXY=https://goproxy.cn,direct

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 使用轻量级镜像作为运行环境
FROM python:3.9-slim

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    curl \
    vim \
    && rm -rf /var/lib/apt/lists/*

# 设置工作目录
WORKDIR /app

# 安装Python依赖
RUN pip install --no-cache-dir \
    qlib \
    pandas \
    numpy \
    scikit-learn \
    lightgbm \
    xgboost \
    mlflow

# 从构建环境复制编译好的二进制文件
COPY --from=builder /app/main .

# 复制配置文件
COPY --from=builder /app/config ./config

# 创建必要的目录
RUN mkdir -p logs uploads output/qlib scripts/qlib

# 设置权限
RUN chmod +x main

# 暴露端口
EXPOSE 8000

# 设置环境变量
ENV GIN_MODE=release
ENV APP_PORT=8000

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# 启动应用
CMD ["./main"]