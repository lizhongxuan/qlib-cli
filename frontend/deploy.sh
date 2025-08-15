#!/bin/bash

# Qlib量化平台部署脚本
# 使用方法: ./deploy.sh [方案] [选项]

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 项目信息
PROJECT_NAME="qlib-frontend"
VERSION="1.0.0"
PORT=8080

# 打印信息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    echo "Qlib量化平台部署脚本"
    echo ""
    echo "使用方法:"
    echo "  ./deploy.sh [方案] [选项]"
    echo ""
    echo "部署方案:"
    echo "  simple       - 简单HTTP服务器部署"
    echo "  nginx        - Nginx部署（需要root权限）"
    echo "  docker       - Docker容器部署"
    echo "  package      - 打包文件准备部署"
    echo ""
    echo "选项:"
    echo "  --port PORT  - 指定端口（默认: 8080）"
    echo "  --host HOST  - 指定主机（默认: 0.0.0.0）"
    echo "  --help       - 显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  ./deploy.sh simple --port 3000"
    echo "  ./deploy.sh nginx"
    echo "  ./deploy.sh docker"
}

# 检查Python3是否安装
check_python() {
    if ! command -v python3 &> /dev/null; then
        print_error "Python3 未安装，请先安装 Python3"
        exit 1
    fi
    print_success "Python3 已安装: $(python3 --version)"
}

# 简单HTTP服务器部署
deploy_simple() {
    local port=${1:-$PORT}
    local host=${2:-"0.0.0.0"}
    
    print_info "启动简单HTTP服务器..."
    print_info "端口: $port"
    print_info "主机: $host"
    
    check_python
    
    # 检查端口是否被占用
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        print_warning "端口 $port 已被占用"
        read -p "是否要杀死占用进程并继续？(y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            lsof -Pi :$port -sTCP:LISTEN -t | xargs kill -9
            print_info "已清理端口 $port"
        else
            exit 1
        fi
    fi
    
    # 启动服务器
    print_info "启动中..."
    nohup python3 -m http.server $port --bind $host > deploy.log 2>&1 &
    SERVER_PID=$!
    
    sleep 2
    
    # 检查是否启动成功
    if ps -p $SERVER_PID > /dev/null; then
        print_success "服务器启动成功!"
        print_info "PID: $SERVER_PID"
        print_info "访问地址: http://localhost:$port"
        print_info "停止服务: kill $SERVER_PID"
        echo $SERVER_PID > .server.pid
    else
        print_error "服务器启动失败，请检查 deploy.log"
        exit 1
    fi
}

# Nginx部署
deploy_nginx() {
    print_info "开始Nginx部署..."
    
    # 检查Nginx是否安装
    if ! command -v nginx &> /dev/null; then
        print_error "Nginx 未安装，请先安装 Nginx"
        print_info "Ubuntu/Debian: sudo apt install nginx"
        print_info "CentOS/RHEL: sudo yum install nginx"
        print_info "macOS: brew install nginx"
        exit 1
    fi
    
    # 创建Nginx配置
    local config_file="/tmp/${PROJECT_NAME}.nginx.conf"
    cat > $config_file << EOF
server {
    listen 80;
    server_name localhost;
    root $(pwd);
    index index.html single-page.html;
    
    # 静态资源缓存
    location ~* \.(css|js|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
    
    # Gzip压缩
    gzip on;
    gzip_types text/css application/javascript text/javascript;
    
    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
}
EOF
    
    print_success "Nginx配置已创建: $config_file"
    print_info "请手动复制配置到Nginx配置目录并重启Nginx:"
    echo "  sudo cp $config_file /etc/nginx/sites-available/${PROJECT_NAME}"
    echo "  sudo ln -s /etc/nginx/sites-available/${PROJECT_NAME} /etc/nginx/sites-enabled/"
    echo "  sudo nginx -t && sudo systemctl reload nginx"
}

# Docker部署
deploy_docker() {
    print_info "开始Docker部署..."
    
    # 检查Docker是否安装
    if ! command -v docker &> /dev/null; then
        print_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    # 创建Dockerfile
    cat > Dockerfile << EOF
FROM nginx:alpine

# 复制项目文件
COPY . /usr/share/nginx/html/

# 创建Nginx配置
RUN echo 'server { \
    listen 80; \
    server_name localhost; \
    root /usr/share/nginx/html; \
    index index.html single-page.html; \
    \
    location ~* \.(css|js|png|jpg|jpeg|gif|ico|svg)$ { \
        expires 1y; \
        add_header Cache-Control "public, immutable"; \
    } \
    \
    gzip on; \
    gzip_types text/css application/javascript text/javascript; \
}' > /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
EOF
    
    # 创建.dockerignore
    cat > .dockerignore << EOF
.git
*.log
.DS_Store
Dockerfile
.dockerignore
deploy.sh
*.md
.server.pid
EOF
    
    print_info "构建Docker镜像..."
    docker build -t ${PROJECT_NAME}:${VERSION} .
    
    print_info "启动Docker容器..."
    docker run -d \
        --name ${PROJECT_NAME} \
        -p ${PORT}:80 \
        --restart unless-stopped \
        ${PROJECT_NAME}:${VERSION}
    
    print_success "Docker部署完成!"
    print_info "容器名称: ${PROJECT_NAME}"
    print_info "访问地址: http://localhost:${PORT}"
    print_info "停止容器: docker stop ${PROJECT_NAME}"
    print_info "删除容器: docker rm -f ${PROJECT_NAME}"
}

# 打包部署文件
deploy_package() {
    print_info "打包部署文件..."
    
    local package_name="${PROJECT_NAME}-${VERSION}.tar.gz"
    
    # 创建临时目录
    local temp_dir=$(mktemp -d)
    local project_dir="$temp_dir/$PROJECT_NAME"
    
    mkdir -p "$project_dir"
    
    # 复制必要文件
    cp *.html *.css *.js "$project_dir/" 2>/dev/null || true
    cp -r components "$project_dir/" 2>/dev/null || true
    
    # 创建启动脚本
    cat > "$project_dir/start.sh" << 'EOF'
#!/bin/bash
PORT=${1:-8080}
echo "启动Qlib量化平台..."
echo "端口: $PORT"
python3 -m http.server $PORT
EOF
    chmod +x "$project_dir/start.sh"
    
    # 创建README
    cat > "$project_dir/README.md" << EOF
# Qlib量化投资平台

## 快速开始

1. 确保已安装Python3
2. 运行启动脚本:
   \`\`\`bash
   ./start.sh [端口号]
   \`\`\`
3. 打开浏览器访问: http://localhost:8080

## 文件说明

- index.html - 主页面（需要HTTP服务器）
- single-page.html - 单页版本（推荐）
- styles.css - 样式文件
- components/ - React组件文件
- app.js - 主应用文件

## 部署选项

- 简单部署: ./start.sh
- Nginx部署: 将文件复制到Nginx根目录
- Docker部署: 使用提供的Dockerfile

访问: http://localhost:端口号
EOF
    
    # 打包
    cd "$temp_dir"
    tar -czf "$package_name" "$PROJECT_NAME"
    
    # 移动到当前目录
    mv "$package_name" "$(pwd)/../"
    cd - > /dev/null
    
    # 清理临时文件
    rm -rf "$temp_dir"
    
    print_success "打包完成: $package_name"
    print_info "部署说明:"
    echo "  1. 将 $package_name 上传到目标服务器"
    echo "  2. 解压: tar -xzf $package_name"
    echo "  3. 进入目录: cd $PROJECT_NAME"
    echo "  4. 启动服务: ./start.sh"
}

# 主函数
main() {
    local method=""
    local port=$PORT
    local host="0.0.0.0"
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            simple|nginx|docker|package)
                method="$1"
                shift
                ;;
            --port)
                port="$2"
                shift 2
                ;;
            --host)
                host="$2"
                shift 2
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                print_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 显示标题
    echo "=================================="
    echo "  Qlib量化平台部署脚本 v${VERSION}"
    echo "=================================="
    echo ""
    
    # 如果没有指定方法，显示帮助
    if [[ -z "$method" ]]; then
        show_help
        exit 1
    fi
    
    # 执行部署
    case $method in
        simple)
            deploy_simple "$port" "$host"
            ;;
        nginx)
            deploy_nginx
            ;;
        docker)
            deploy_docker
            ;;
        package)
            deploy_package
            ;;
        *)
            print_error "不支持的部署方案: $method"
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"