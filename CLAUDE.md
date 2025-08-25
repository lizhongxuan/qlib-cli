# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是一个基于Microsoft Qlib框架构建的端到端量化投资研究平台，提供可视化界面完成从数据准备、因子研究、模型训练到策略回测的完整量化投资流程。

## 技术架构

### 后端 (Go)
- **框架**: Gin Web框架 + GORM ORM + MySQL
- **模块路径**: `qlib-backend`
- **Go版本**: 1.22.0
- **主要依赖**: gin-gonic/gin, gorm.io/gorm, gorilla/websocket, golang-jwt/jwt

### 前端 (JavaScript)
- **技术栈**: React + JavaScript (无TypeScript)
- **位置**: `frontend/` 目录

### 量化引擎
- **核心**: Microsoft Qlib Python框架
- **集成**: 通过Go调用Python脚本实现

## 常用开发命令

### 后端开发
```bash
# 进入后端目录
cd backend

# 安装依赖
go mod download

# 运行开发服务器
go run main.go

# 构建项目
go build -o qlib-backend main.go

# 数据库迁移
# 迁移会在启动时自动执行
```

### 前端开发
```bash
# 进入前端目录
cd frontend

# 启动前端服务器（通过Python SimpleHTTPServer）
python -m http.server 3000
# 或者
./deploy.sh
```

### 环境配置
数据库和服务配置通过环境变量管理，默认值在 `backend/config/config.go` 中定义：

```bash
# 数据库配置
export DB_HOST=localhost
export DB_PORT=3306
export DB_USERNAME=root
export DB_PASSWORD=password
export DB_DATABASE=qlib

# 应用配置
export APP_PORT=8000
export GIN_MODE=debug

# Qlib配置
export QLIB_PYTHON_PATH=/usr/bin/python3
export QLIB_DATA_PATH=~/.qlib/qlib_data
export QLIB_CACHE_PATH=~/.qlib/cache

# JWT配置
export JWT_SECRET=qlib-secret-key
export JWT_EXPIRE=24
```

## 核心架构模式

### 后端分层架构
```
backend/
├── main.go                    # 应用入口点
├── config/                    # 配置管理
├── internal/
│   ├── api/
│   │   ├── handlers/          # HTTP处理器（业务接口层）
│   │   ├── middleware/        # 中间件（认证、日志、恢复）
│   │   ├── routes/           # 路由配置
│   │   └── websocket/        # WebSocket处理
│   ├── models/               # 数据模型定义
│   ├── services/             # 业务逻辑层
│   ├── qlib/                 # Qlib Python接口封装
│   └── utils/                # 工具函数
└── pkg/                      # 公共包
```

### API设计模式
- **RESTful API**: 所有接口遵循REST设计原则
- **统一响应格式**: 使用 `utils/response.go` 中定义的标准响应结构
- **模块化路由**: 按功能模块组织路由（dashboard、data、factors、models等）
- **WebSocket支持**: 实时推送任务进度和系统状态

### 业务模块架构
1. **数据管理**: 数据集CRUD、数据源管理、数据探索
2. **因子管理**: 因子CRUD、因子测试、批量测试、AI助手集成
3. **模型训练**: 模型训练、进度监控、模型评估、模型对比
4. **策略回测**: 策略配置、回测执行、结果分析、参数优化
5. **工作流管理**: Qlib工作流执行、模板管理、进度监控
6. **结果分析**: 性能对比、报告生成、统计分析

### Qlib集成架构
- **Python接口封装**: `internal/qlib/` 目录下的各种接口文件
- **异步任务处理**: 通过Go routine处理长时间运行的Qlib任务
- **实时进度推送**: 使用WebSocket推送任务执行进度
- **配置管理**: 支持动态生成Qlib YAML配置文件

## 数据库架构

### 自动迁移
- 应用启动时自动执行数据库迁移
- 模型定义在 `internal/models/` 目录
- 使用GORM的AutoMigrate功能

### 主要数据模型
- **Base**: 通用基础模型（ID、创建时间、更新时间、删除时间）
- **Workflow**: 工作流模型
- 其他模型分散在各service文件中定义

## WebSocket实时通信

### 支持的事件类型
- 工作流进度推送 (`/ws/workflow-progress/:task_id`)
- 因子测试进度 (`/ws/factor-test/:test_id`)
- 系统监控数据 (`/ws/system-monitor`)
- 系统通知 (`/ws/notifications`)
- 任务状态推送 (`/ws/task/:task_id`)
- 任务日志推送 (`/ws/logs/:task_id`)

### 广播服务
使用 `services/broadcast_service.go` 和 `services/websocket_service.go` 管理WebSocket连接和消息广播。

## 开发注意事项

### 代码规范
- Go代码遵循官方Go代码规范
- 中文注释和响应消息，因为这是面向中文用户的量化平台
- 使用GORM进行数据库操作
- 统一的错误处理和日志记录

### 项目状态
- 大部分API接口已完成实现（100个接口）
- WebSocket实时通信已实现（7个事件）
- 完整的项目结构和基础设施已搭建
- 需要实现的部分在 `docs/backend_api_implementation_todo.md` 中标记

### 测试
目前项目没有测试文件，建议为新功能添加单元测试和集成测试。

### 前端集成
前端使用简单的JavaScript + HTML，通过 `frontend/api/qlib-api.js` 与后端API通信。