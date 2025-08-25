#!/bin/bash

# Qlib后端部署脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 配置变量
DEPLOY_ENV=${DEPLOY_ENV:-production}
DEPLOY_USER=${DEPLOY_USER:-qlib}
DEPLOY_HOST=${DEPLOY_HOST:-localhost}
DEPLOY_PATH=${DEPLOY_PATH:-/opt/qlib-backend}
SERVICE_NAME=${SERVICE_NAME:-qlib-backend}
BACKUP_DIR=${BACKUP_DIR:-/opt/qlib-backend/backups}

# 检查必要的工具
check_tools() {
    print_step "Checking required tools..."
    
    local tools=("docker" "docker-compose" "git")
    for tool in "${tools[@]}"; do
        if ! command -v $tool &> /dev/null; then
            print_error "$tool is not installed"
            exit 1
        fi
    done
    
    print_message "All required tools are available"
}

# 备份当前部署
backup_current() {
    print_step "Creating backup..."
    
    local backup_name="backup_$(date +%Y%m%d_%H%M%S)"
    local backup_path="$BACKUP_DIR/$backup_name"
    
    if [ -d "$DEPLOY_PATH" ]; then
        mkdir -p "$BACKUP_DIR"
        cp -r "$DEPLOY_PATH" "$backup_path"
        print_message "Backup created: $backup_path"
    else
        print_warning "No existing deployment to backup"
    fi
}

# 停止服务
stop_services() {
    print_step "Stopping services..."
    
    # 停止Docker Compose服务
    if [ -f "$DEPLOY_PATH/docker-compose.yml" ]; then
        cd "$DEPLOY_PATH"
        docker-compose down
        print_message "Docker services stopped"
    fi
    
    # 停止系统服务（如果存在）
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        sudo systemctl stop "$SERVICE_NAME"
        print_message "System service stopped"
    fi
}

# 部署应用
deploy_app() {
    print_step "Deploying application..."
    
    # 创建部署目录
    mkdir -p "$DEPLOY_PATH"
    
    # 复制文件
    cp -r . "$DEPLOY_PATH/"
    
    # 设置权限
    sudo chown -R "$DEPLOY_USER:$DEPLOY_USER" "$DEPLOY_PATH"
    chmod +x "$DEPLOY_PATH/bin/qlib-backend"
    
    print_message "Application deployed to $DEPLOY_PATH"
}

# 配置环境
setup_environment() {
    print_step "Setting up environment..."
    
    # 创建环境文件
    cat > "$DEPLOY_PATH/.env" << EOF
# 环境配置
DEPLOY_ENV=$DEPLOY_ENV
GIN_MODE=release
APP_PORT=8000

# 数据库配置
DB_HOST=mysql
DB_PORT=3306
DB_USERNAME=qlib
DB_PASSWORD=qlib123456
DB_DATABASE=qlib

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379

# JWT配置
JWT_SECRET=$(openssl rand -base64 32)

# Qlib配置
QLIB_PYTHON_PATH=/usr/bin/python3
QLIB_DATA_PATH=/root/.qlib/qlib_data
QLIB_CACHE_PATH=/root/.qlib/cache
EOF
    
    print_message "Environment configured"
}

# 初始化数据库
init_database() {
    print_step "Initializing database..."
    
    # 等待MySQL启动
    print_message "Waiting for MySQL to be ready..."
    while ! docker-compose exec mysql mysqladmin ping -h"localhost" --silent; do
        sleep 1
    done
    
    # 运行数据库迁移
    docker-compose exec qlib-backend ./qlib-backend migrate
    
    print_message "Database initialized"
}

# 启动服务
start_services() {
    print_step "Starting services..."
    
    cd "$DEPLOY_PATH"
    
    # 启动Docker Compose服务
    docker-compose up -d
    
    # 等待服务启动
    print_message "Waiting for services to start..."
    sleep 10
    
    # 检查服务状态
    if docker-compose ps | grep -q "Up"; then
        print_message "Services started successfully"
    else
        print_error "Failed to start services"
        docker-compose logs
        exit 1
    fi
}

# 健康检查
health_check() {
    print_step "Performing health check..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f http://localhost:8000/health >/dev/null 2>&1; then
            print_message "Health check passed"
            return 0
        fi
        
        print_message "Attempt $attempt/$max_attempts: Health check failed, retrying..."
        sleep 2
        ((attempt++))
    done
    
    print_error "Health check failed after $max_attempts attempts"
    return 1
}

# 配置系统服务（可选）
setup_systemd() {
    print_step "Setting up systemd service..."
    
    cat > "/etc/systemd/system/$SERVICE_NAME.service" << EOF
[Unit]
Description=Qlib Backend Service
After=network.target

[Service]
Type=simple
User=$DEPLOY_USER
WorkingDirectory=$DEPLOY_PATH
ExecStart=$DEPLOY_PATH/bin/qlib-backend
Restart=always
RestartSec=5
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOF
    
    sudo systemctl daemon-reload
    sudo systemctl enable "$SERVICE_NAME"
    
    print_message "Systemd service configured"
}

# 配置Nginx（可选）
setup_nginx() {
    print_step "Setting up Nginx..."
    
    cat > "/etc/nginx/sites-available/$SERVICE_NAME" << EOF
server {
    listen 80;
    server_name qlib.example.com;
    
    location / {
        proxy_pass http://localhost:8000;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    location /ws/ {
        proxy_pass http://localhost:8000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF
    
    sudo ln -sf "/etc/nginx/sites-available/$SERVICE_NAME" "/etc/nginx/sites-enabled/"
    sudo nginx -t && sudo systemctl reload nginx
    
    print_message "Nginx configured"
}

# 清理旧的Docker镜像和容器
cleanup_docker() {
    print_step "Cleaning up Docker..."
    
    # 删除未使用的镜像
    docker image prune -f
    
    # 删除未使用的容器
    docker container prune -f
    
    print_message "Docker cleanup completed"
}

# 显示部署信息
show_deploy_info() {
    print_step "Deployment Information"
    
    echo "=================================="
    echo "Environment: $DEPLOY_ENV"
    echo "Deploy Path: $DEPLOY_PATH"
    echo "Service Name: $SERVICE_NAME"
    echo "Health Check: http://localhost:8000/health"
    echo "API Base URL: http://localhost:8000/api/v1"
    echo "=================================="
    
    if [ -f "$DEPLOY_PATH/docker-compose.yml" ]; then
        echo "Docker Services:"
        docker-compose ps
    fi
}

# 主函数
main() {
    print_message "Starting deployment process..."
    
    # 解析命令行参数
    SKIP_BACKUP=false
    SKIP_HEALTH_CHECK=false
    SETUP_NGINX=false
    SETUP_SYSTEMD=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-backup)
                SKIP_BACKUP=true
                shift
                ;;
            --skip-health-check)
                SKIP_HEALTH_CHECK=true
                shift
                ;;
            --nginx)
                SETUP_NGINX=true
                shift
                ;;
            --systemd)
                SETUP_SYSTEMD=true
                shift
                ;;
            --env)
                DEPLOY_ENV="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [options]"
                echo "Options:"
                echo "  --skip-backup      Skip backup creation"
                echo "  --skip-health-check Skip health check"
                echo "  --nginx            Setup Nginx configuration"
                echo "  --systemd          Setup systemd service"
                echo "  --env ENV          Set deployment environment"
                echo "  --help             Show this help"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # 检查工具
    check_tools
    
    # 备份当前部署（如果没有跳过）
    if [ "$SKIP_BACKUP" = false ]; then
        backup_current
    fi
    
    # 停止服务
    stop_services
    
    # 部署应用
    deploy_app
    
    # 配置环境
    setup_environment
    
    # 启动服务
    start_services
    
    # 初始化数据库
    init_database
    
    # 健康检查（如果没有跳过）
    if [ "$SKIP_HEALTH_CHECK" = false ]; then
        health_check
    fi
    
    # 配置可选服务
    if [ "$SETUP_SYSTEMD" = true ]; then
        setup_systemd
    fi
    
    if [ "$SETUP_NGINX" = true ]; then
        setup_nginx
    fi
    
    # 清理Docker
    cleanup_docker
    
    # 显示部署信息
    show_deploy_info
    
    print_message "Deployment completed successfully!"
}

# 执行主函数
main "$@"