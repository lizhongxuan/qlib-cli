#!/bin/bash

# Qlib后端构建脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Go环境
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        exit 1
    fi
    print_message "Go version: $(go version)"
}

# 检查Python环境
check_python() {
    if ! command -v python3 &> /dev/null; then
        print_error "Python3 is not installed"
        exit 1
    fi
    print_message "Python version: $(python3 --version)"
}

# 安装Go依赖
install_go_deps() {
    print_message "Installing Go dependencies..."
    go mod download
    go mod tidy
}

# 安装Python依赖
install_python_deps() {
    print_message "Installing Python dependencies..."
    pip3 install -r requirements.txt || true
}

# 运行测试
run_tests() {
    print_message "Running tests..."
    go test ./... -v
}

# 构建应用
build_app() {
    print_message "Building application..."
    
    # 设置构建变量
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "unknown")
    BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
    COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
    
    # 构建
    go build -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.Commit=${COMMIT}" -o bin/qlib-backend main.go
    
    print_message "Build completed: bin/qlib-backend"
}

# 构建Docker镜像
build_docker() {
    print_message "Building Docker image..."
    docker build -t qlib-backend:latest -f docker/Dockerfile .
    print_message "Docker image built: qlib-backend:latest"
}

# 清理构建文件
clean() {
    print_message "Cleaning build files..."
    rm -rf bin/
    rm -rf logs/
    rm -rf uploads/
    rm -rf output/
    print_message "Cleanup completed"
}

# 创建必要的目录
create_dirs() {
    print_message "Creating necessary directories..."
    mkdir -p bin logs uploads output/qlib scripts/qlib
}

# 复制配置文件
copy_configs() {
    print_message "Copying configuration files..."
    cp -r config/ bin/ 2>/dev/null || true
}

# 主函数
main() {
    print_message "Starting build process..."
    
    # 解析命令行参数
    SKIP_TESTS=false
    BUILD_DOCKER=false
    CLEAN_FIRST=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-tests)
                SKIP_TESTS=true
                shift
                ;;
            --docker)
                BUILD_DOCKER=true
                shift
                ;;
            --clean)
                CLEAN_FIRST=true
                shift
                ;;
            --help)
                echo "Usage: $0 [options]"
                echo "Options:"
                echo "  --skip-tests  Skip running tests"
                echo "  --docker      Build Docker image"
                echo "  --clean       Clean before build"
                echo "  --help        Show this help"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # 如果指定清理，先执行清理
    if [ "$CLEAN_FIRST" = true ]; then
        clean
    fi
    
    # 检查环境
    check_go
    check_python
    
    # 创建目录
    create_dirs
    
    # 安装依赖
    install_go_deps
    install_python_deps
    
    # 运行测试（如果没有跳过）
    if [ "$SKIP_TESTS" = false ]; then
        run_tests
    fi
    
    # 构建应用
    build_app
    
    # 复制配置文件
    copy_configs
    
    # 构建Docker镜像（如果指定）
    if [ "$BUILD_DOCKER" = true ]; then
        build_docker
    fi
    
    print_message "Build process completed successfully!"
    print_message "Binary: bin/qlib-backend"
    
    if [ "$BUILD_DOCKER" = true ]; then
        print_message "Docker image: qlib-backend:latest"
    fi
}

# 执行主函数
main "$@"