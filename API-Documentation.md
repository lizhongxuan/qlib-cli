# Qlib可视化平台完整API文档

## 概述

基于对前端页面功能的深入分析，本文档详细定义了Qlib可视化平台的完整API接口规范。API采用RESTful设计风格，支持JSON数据格式，包含WebSocket实时通信支持。

## 基础信息

- **Base URL**: `http://localhost:8000/api/v1`
- **WebSocket URL**: `ws://localhost:8000/ws`
- **数据格式**: JSON
- **认证方式**: JWT Token
- **编码格式**: UTF-8

## 通用响应格式

```json
{
  "success": true,
  "code": 200,
  "message": "操作成功",
  "data": {},
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## 错误码说明

- `200`: 成功
- `400`: 请求参数错误
- `401`: 未授权
- `403`: 权限不足
- `404`: 资源不存在
- `500`: 服务器内部错误

---

 Qlib可视化平台API分类目录

  📊 1. 系统总览 API (Dashboard)

  GET /dashboard/overview                    # 获取总览统计数据
  GET /dashboard/market-overview             # 获取市场数据概览
  GET /dashboard/performance-chart           # 获取净值走势数据
  GET /dashboard/recent-activities           # 获取最近活动任务

  💾 2. 数据管理 API (Data Management)

  2.1 数据集管理

  GET    /data/datasets                      # 获取数据集列表
  POST   /data/datasets                      # 创建新数据集
  GET    /data/datasets/{dataset_id}         # 获取数据集详情
  DELETE /data/datasets/{dataset_id}         # 删除数据集
  GET    /data/datasets/{dataset_id}/statistics  # 
  获取数据集统计分析

  2.2 数据源管理

  GET  /data/sources                         # 获取数据源列表
  POST /data/sources                         # 添加数据源
  POST /data/sources/{source_id}/test        # 测试数据源连接

  🧮 3. 因子管理 API (Factor Management)

  3.1 因子工程

  GET  /factors/built-in                     # 获取内置因子库
  POST /factors/ai-generate                  # AI因子生成

  3.2 因子编辑器

  POST /factors/validate                     # 因子表达式语法检查
  POST /factors/test                         # 因子表达式测试
  GET  /factors/test/{task_id}/result        # 获取因子测试结果

  3.3 因子分析

  GET /factors/{factor_id}/analysis          # 获取因子分析报告

  3.4 因子库管理

  GET    /factors                            # 获取已保存因子列表
  POST   /factors                            # 保存因子
  PUT    /factors/{factor_id}                # 更新因子
  DELETE /factors/{factor_id}                # 删除因子

  🤖 4. 模型训练 API (Model Training)

  4.1 模型管理

  GET  /models                               # 获取模型列表
  POST /models/train                         # 创建训练任务
  GET  /models/{model_id}/performance        # 获取模型详细性能
  POST /models/{model_id}/deploy             # 部署模型

  4.2 训练监控

  GET  /models/train/{task_id}/status        # 获取训练任务状态
  GET  /models/train/{task_id}/metrics       # 获取实时训练指标
  POST /models/train/{task_id}/stop          # 停止训练任务

  4.3 模型对比

  POST /models/compare                       # 模型对比分析

  📈 5. 策略回测 API (Strategy Backtest)

  5.1 策略配置

  GET /strategies/types                      # 获取可用策略类型

  5.2 回测执行

  POST /backtest/create                      # 创建回测任务
  GET  /backtest/{task_id}/progress          # 获取回测进度

  5.3 回测结果

  GET /backtest/{backtest_id}/results        # 获取回测结果概览
  GET /backtest/{backtest_id}/nav-curve      # 获取净值曲线数据
  GET /backtest/{backtest_id}/positions      # 获取持仓明细
  GET /backtest/{backtest_id}/trades         # 获取交易记录

  5.4 回测分析

  GET /backtest/{backtest_id}/attribution    # 获取收益归因分析
  GET /backtest/{backtest_id}/risk-analysis  # 获取风险分析报告

  ⚙️ 6. Qlib工作流 API (Qlib Workflow)

  6.1 工作流配置

  GET  /qlib/presets                         # 获取qlib预设配置
  POST /qlib/validate-config                 # 验证qlib配置

  6.2 工作流执行

  POST /qlib/workflow/run                    # 启动qlib工作流
  GET  /qlib/workflow/{task_id}/progress     # 获取qlib工作流进度
  GET  /qlib/workflow/{task_id}/result       # 获取qlib工作流结果

  6.3 工作流管理

  GET  /qlib/workflow/history                # 获取工作流历史
  POST /qlib/workflow/{task_id}/stop         # 停止qlib工作流

  📊 7. 结果分析 API (Results Analysis)

  7.1 分析概览

  GET /analysis/overview                     # 获取分析结果概览

  7.2 模型分析

  POST /analysis/models/compare              # 获取模型性能对比
  GET  /analysis/models/{result_id}/factor-importance  # 
  获取因子重要性分析

  7.3 策略分析

  GET  /analysis/strategies/{result_id}/performance    # 
  获取策略绩效分析
  POST /analysis/strategies/compare          # 多策略对比分析

  7.4 报告生成

  POST /analysis/reports/generate            # 生成分析报告
  GET  /analysis/reports/{task_id}/status    # 获取报告生成状态
  GET  /analysis/reports/{report_id}/download # 下载分析报告

  🔌 8. WebSocket 实时通信

  ws://localhost:8000/ws/task/{task_id}      # 任务状态推送
  ws://localhost:8000/ws/system              # 系统状态推送
  ws://localhost:8000/ws/logs/{task_id}      # 实时日志推送

  🛠️ 9. 通用工具 API (Utilities)

  9.1 文件操作

  POST /files/upload                         # 文件上传
  GET  /files/{file_id}/download             # 文件下载

  9.2 任务管理

  GET  /tasks                                # 获取任务列表
  POST /tasks/{task_id}/cancel               # 取消任务

  ---
  API统计总览

  | 分类        | 接口数量 | 主要功能               |
  |-----------|------|--------------------|
  | 系统总览      | 4个   | 系统状态、市场数据、性能监控     |
  | 数据管理      | 8个   | 数据集CRUD、数据源管理、数据探索 |
  | 因子管理      | 9个   | 因子工程、编辑测试、分析管理     |
  | 模型训练      | 8个   | 模型训练、监控、评估、部署      |
  | 策略回测      | 9个   | 策略配置、回测执行、结果分析     |
  | Qlib工作流   | 7个   | 端到端量化研究流程          |
  | 结果分析      | 7个   | 结果对比、报告生成          |
  | WebSocket | 3个   | 实时通信推送             |
  | 通用工具      | 4个   | 文件管理、任务管理          |
  | 总计        | 59个  | 完整量化平台功能           |

---

## 1. 系统总览页面 (Dashboard) API

### 1.1 获取总览统计数据

**接口路径**: `GET /dashboard/overview`

**功能描述**: 获取系统总览页面的核心统计数据

**响应示例**:
```json
{
  "success": true,
  "data": {
    "statistics": {
      "total_datasets": 15,
      "ready_datasets": 12,
      "total_models": 8,
      "trained_models": 6,
      "running_tasks": 2,
      "completed_tasks": 45
    },
    "performance": {
      "total_return": "25.0%",
      "sharpe_ratio": "1.85",
      "max_drawdown": "-8.5%",
      "win_rate": "62.3%"
    },
    "system_resources": {
      "cpu_usage": 65,
      "memory_usage": 78,
      "disk_usage": 45,
      "gpu_usage": 23
    }
  }
}
```

### 1.2 获取市场数据概览

**接口路径**: `GET /dashboard/market-overview`

**功能描述**: 获取主要指数的实时行情数据

**响应示例**:
```json
{
  "success": true,
  "data": {
    "markets": [
      {
        "symbol": "SH000300",
        "name": "沪深300",
        "value": 3456.78,
        "change": "+1.23%",
        "trend": "up"
      },
      {
        "symbol": "SZ399905", 
        "name": "中证500",
        "value": 6789.12,
        "change": "-0.45%",
        "trend": "down"
      }
    ]
  }
}
```

### 1.3 获取净值走势数据

**接口路径**: `GET /dashboard/performance-chart`

**请求参数**:
- `time_range` (optional): 时间范围，可选值: 1D, 1W, 1M, 3M, 1Y

**响应示例**:
```json
{
  "success": true,
  "data": {
    "performance_data": [
      {
        "date": "2024-01-01",
        "value": 100,
        "benchmark": 100,
        "volume": 85000000
      }
    ]
  }
}
```

### 1.4 获取最近活动任务

**接口路径**: `GET /dashboard/recent-activities`

**请求参数**:
- `limit` (optional): 返回数量限制，默认5

**响应示例**:
```json
{
  "success": true,
  "data": {
    "activities": [
      {
        "id": "task_001",
        "name": "训练LightGBM模型",
        "type": "model_training",
        "status": "completed",
        "progress": 100,
        "start_time": "2024-01-15T09:00:00Z",
        "end_time": "2024-01-15T09:30:00Z"
      }
    ]
  }
}
```

---

## 2. 数据管理页面 (DataManagement) API

### 2.1 数据集管理接口

#### 2.1.1 获取数据集列表

**接口路径**: `GET /data/datasets`

**请求参数**:
- `page` (optional): 页码，默认1
- `limit` (optional): 每页数量，默认10
- `status` (optional): 状态筛选，可选值: ready, preparing, error
- `search` (optional): 搜索关键词

**响应示例**:
```json
{
  "success": true,
  "data": {
    "datasets": [
      {
        "id": "dataset_001",
        "name": "CSI300-2020-2023",
        "status": "ready",
        "samples": 245000,
        "features": 158,
        "date_range": "2020-01-01 至 2023-12-31",
        "create_time": "2024-01-15T08:00:00Z",
        "update_time": "2024-01-15T08:30:00Z",
        "market": "csi300",
        "benchmark": "SH000300"
      }
    ],
    "pagination": {
      "total": 15,
      "page": 1,
      "limit": 10,
      "total_pages": 2
    }
  }
}
```

#### 2.1.2 创建新数据集

**接口路径**: `POST /data/datasets`

**请求体**:
```json
{
  "name": "CSI500-2021-2023",
  "market": "csi500",
  "start_date": "2021-01-01",
  "end_date": "2023-12-31",
  "features": ["open", "close", "high", "low", "volume", "rsi", "macd"],
  "label": "Ref($close, -1) / $close - 1",
  "benchmark": "SZ399905"
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "dataset_id": "dataset_002",
    "task_id": "task_002",
    "message": "数据集创建任务已启动"
  }
}
```

#### 2.1.3 获取数据集详情

**接口路径**: `GET /data/datasets/{dataset_id}`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "id": "dataset_001",
    "name": "CSI300-2020-2023",
    "status": "ready",
    "samples": 245000,
    "features": 158,
    "date_range": "2020-01-01 至 2023-12-31",
    "market": "csi300",
    "benchmark": "SH000300",
    "feature_list": ["RESI5", "WVMA5", "RSQR10", "..."],
    "statistics": {
      "missing_ratio": 0.02,
      "correlation_with_label": 0.15,
      "feature_importance": [
        {"name": "RESI5", "importance": 0.125},
        {"name": "WVMA5", "importance": 0.098}
      ]
    }
  }
}
```

#### 2.1.4 删除数据集

**接口路径**: `DELETE /data/datasets/{dataset_id}`

**响应示例**:
```json
{
  "success": true,
  "message": "数据集删除成功"
}
```

### 2.2 数据源管理接口

#### 2.2.1 获取数据源列表

**接口路径**: `GET /data/sources`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "sources": [
      {
        "id": "yahoo",
        "name": "Yahoo Finance",
        "type": "API",
        "status": "在线",
        "last_update": "实时",
        "description": "免费股票数据"
      }
    ]
  }
}
```

#### 2.2.2 添加数据源

**接口路径**: `POST /data/sources`

**请求体**:
```json
{
  "name": "自定义数据源",
  "type": "api",
  "url": "https://api.example.com/data",
  "description": "自定义数据接口",
  "auth_type": "api_key",
  "credentials": {
    "api_key": "your_api_key"
  }
}
```

#### 2.2.3 测试数据源连接

**接口路径**: `POST /data/sources/{source_id}/test`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "status": "connected",
    "response_time": 150,
    "message": "连接测试成功"
  }
}
```

### 2.3 数据探索接口

#### 2.3.1 获取数据集统计分析

**接口路径**: `GET /data/datasets/{dataset_id}/statistics`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "basic_stats": {
      "total_samples": 245000,
      "feature_count": 158,
      "missing_values": 4900,
      "date_range": ["2020-01-01", "2023-12-31"]
    },
    "distribution": {
      "feature_distributions": [
        {
          "name": "close",
          "mean": 15.67,
          "std": 12.34,
          "min": 2.1,
          "max": 89.5
        }
      ]
    },
    "correlation_matrix": {
      "features": ["RESI5", "WVMA5", "RSQR10"],
      "matrix": [[1.0, 0.15, -0.23], [0.15, 1.0, 0.08], [-0.23, 0.08, 1.0]]
    }
  }
}
```

---

## 3. 因子管理页面 (FactorManagement) API

### 3.1 因子工程接口

#### 3.1.1 获取内置因子库

**接口路径**: `GET /factors/built-in`

**请求参数**:
- `category` (optional): 因子类别，可选值: technical, fundamental, volume, volatility, cross_sectional
- `search` (optional): 搜索关键词

**响应示例**:
```json
{
  "success": true,
  "data": {
    "categories": {
      "technical": {
        "name": "技术指标",
        "icon": "📈",
        "desc": "基于价格和成交量的技术分析因子",
        "count": 25
      }
    },
    "factors": [
      {
        "id": "rsi",
        "name": "RSI相对强弱指数",
        "expression": "(Sum(Max($close - Ref($close, 1), 0), 14) / Sum(Abs($close - Ref($close, 1)), 14)) * 100",
        "description": "衡量价格变动速度和幅度的技术指标",
        "category": "technical",
        "complexity": "medium",
        "return_period": "短期",
        "tags": ["动量", "技术分析", "超买超卖"]
      }
    ]
  }
}
```

#### 3.1.2 AI因子生成

**接口路径**: `POST /factors/ai-generate`

**请求体**:
```json
{
  "description": "我想要一个捕捉短期动量的因子",
  "context": {
    "market": "csi300",
    "timeframe": "daily",
    "style": "momentum"
  }
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "generated_factor": {
      "name": "AI动量因子",
      "expression": "Rank($close / Ref($close, 20) - 1)",
      "description": "基于20日价格变化的动量因子，适用于捕捉短期趋势",
      "confidence": 0.85,
      "suggested_parameters": {
        "lookback_period": 20,
        "rebalance_frequency": "daily"
      }
    }
  }
}
```

### 3.2 因子编辑器接口

#### 3.2.1 因子表达式语法检查

**接口路径**: `POST /factors/validate`

**请求体**:
```json
{
  "expression": "($close - Mean($close, 20)) / Std($close, 20)",
  "context": {
    "dataset_id": "dataset_001"
  }
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "is_valid": true,
    "syntax_errors": [],
    "warnings": ["建议使用Rank()函数进行横截面标准化"],
    "estimated_computation_time": 15.5
  }
}
```

#### 3.2.2 因子表达式测试

**接口路径**: `POST /factors/test`

**请求体**:
```json
{
  "expression": "($close - Mean($close, 20)) / Std($close, 20)",
  "dataset_id": "dataset_001",
  "test_period": {
    "start_date": "2023-01-01",
    "end_date": "2023-12-31"
  }
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "task_id": "test_task_001",
    "message": "因子测试任务已启动",
    "estimated_duration": 120
  }
}
```

#### 3.2.3 获取因子测试结果

**接口路径**: `GET /factors/test/{task_id}/result`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "status": "completed",
    "results": {
      "ic": 0.0367,
      "ic_ir": 1.85,
      "rank_ic": 0.0489,
      "rank_ic_ir": 2.15,
      "turnover": 0.234,
      "coverage": 0.892,
      "valid_periods": 248,
      "distribution": {
        "bins": [-0.2, -0.1, 0, 0.1, 0.2],
        "frequencies": [12, 45, 120, 58, 13]
      },
      "time_series": [
        {
          "date": "2023-01-01",
          "ic": 0.0423,
          "coverage": 0.89
        }
      ]
    }
  }
}
```

### 3.3 因子分析接口

#### 3.3.1 获取因子分析报告

**接口路径**: `GET /factors/{factor_id}/analysis`

**请求参数**:
- `dataset_id`: 数据集ID
- `period`: 分析周期，可选值: daily, weekly, monthly

**响应示例**:
```json
{
  "success": true,
  "data": {
    "factor_info": {
      "id": "factor_001",
      "name": "动量因子V1.0",
      "expression": "Rank($close / Ref($close, 20) - 1)"
    },
    "statistics": {
      "ic_mean": 0.0342,
      "ic_std": 0.0876,
      "ic_ir": 0.390,
      "rank_ic_mean": 0.0456,
      "rank_ic_ir": 0.406,
      "turnover_mean": 0.234,
      "coverage_mean": 0.892
    },
    "charts_data": {
      "ic_series": [
        {
          "date": "2023-01-01",
          "ic": 0.0423,
          "rank_ic": 0.0512
        }
      ],
      "quantile_analysis": [
        {
          "quantile": 1,
          "mean_return": 0.0234,
          "cumulative_return": 0.156,
          "sharpe": 1.23
        }
      ]
    }
  }
}
```

### 3.4 因子库管理接口

#### 3.4.1 保存因子

**接口路径**: `POST /factors`

**请求体**:
```json
{
  "name": "自定义动量因子",
  "expression": "Rank($close / Ref($close, 20) - 1)",
  "description": "基于20日收益率的排名因子",
  "category": "momentum",
  "tags": ["动量", "排名", "短期"]
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "factor_id": "factor_002",
    "message": "因子保存成功"
  }
}
```

#### 3.4.2 获取已保存因子列表

**接口路径**: `GET /factors`

**请求参数**:
- `page` (optional): 页码，默认1
- `limit` (optional): 每页数量，默认10
- `category` (optional): 类别筛选
- `status` (optional): 状态筛选

**响应示例**:
```json
{
  "success": true,
  "data": {
    "factors": [
      {
        "id": "factor_001",
        "name": "自定义动量因子",
        "expression": "Rank($close / Ref($close, 20) - 1)",
        "description": "基于20日收益率的排名因子",
        "category": "momentum",
        "status": "active",
        "create_time": "2024-01-15T10:00:00Z",
        "last_test_time": "2024-01-15T11:00:00Z",
        "performance": {
          "ic": 0.0367,
          "ic_ir": 1.85
        }
      }
    ],
    "pagination": {
      "total": 25,
      "page": 1,
      "limit": 10
    }
  }
}
```

#### 3.4.3 更新因子

**接口路径**: `PUT /factors/{factor_id}`

**请求体**:
```json
{
  "name": "优化的动量因子",
  "expression": "Rank($close / Ref($close, 15) - 1)",
  "description": "调整为15日回看期的动量因子"
}
```

#### 3.4.4 删除因子

**接口路径**: `DELETE /factors/{factor_id}`

**响应示例**:
```json
{
  "success": true,
  "message": "因子删除成功"
}
```

---

## 4. 模型训练页面 (ModelTraining) API

### 4.1 模型管理接口

#### 4.1.1 获取模型列表

**接口路径**: `GET /models`

**请求参数**:
- `page` (optional): 页码，默认1
- `limit` (optional): 每页数量，默认10
- `status` (optional): 状态筛选，可选值: training, trained, failed
- `model_type` (optional): 模型类型筛选

**响应示例**:
```json
{
  "success": true,
  "data": {
    "models": [
      {
        "id": "model_001",
        "name": "LightGBM-Alpha158",
        "type": "lightgbm",
        "status": "trained",
        "dataset_id": "dataset_001",
        "performance": {
          "ic": 0.0456,
          "ic_ir": 1.85,
          "sharpe": 1.23
        },
        "training_time": "15 分钟",
        "create_time": "2024-01-15T09:00:00Z",
        "train_end_time": "2024-01-15T09:15:00Z"
      }
    ],
    "pagination": {
      "total": 8,
      "page": 1,
      "limit": 10
    }
  }
}
```

#### 4.1.2 创建训练任务

**接口路径**: `POST /models/train`

**请求体**:
```json
{
  "name": "XGBoost-CustomFactors",
  "model_type": "xgboost",
  "dataset_id": "dataset_001",
  "parameters": {
    "learning_rate": 0.01,
    "n_estimators": 200,
    "max_depth": 6,
    "subsample": 0.8,
    "colsample_bytree": 0.8
  },
  "training_config": {
    "validation_split": 0.2,
    "early_stopping": true,
    "early_stopping_rounds": 50,
    "eval_metric": "ic"
  },
  "feature_selection": {
    "method": "importance",
    "top_k": 100
  }
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "model_id": "model_002",
    "task_id": "train_task_002",
    "message": "模型训练任务已启动",
    "estimated_duration": 1800
  }
}
```

### 4.2 训练监控接口

#### 4.2.1 获取训练任务状态

**接口路径**: `GET /models/train/{task_id}/status`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "task_id": "train_task_002",
    "status": "running",
    "progress": 65,
    "current_epoch": 130,
    "total_epochs": 200,
    "metrics": {
      "train_loss": 0.234,
      "valid_loss": 0.267,
      "train_ic": 0.0456,
      "valid_ic": 0.0398
    },
    "estimated_remaining_time": 540,
    "logs": [
      "[2024-01-15 10:30:00] Epoch 130/200 - train_loss: 0.234, valid_loss: 0.267"
    ]
  }
}
```

#### 4.2.2 获取实时训练指标

**接口路径**: `GET /models/train/{task_id}/metrics`

**请求参数**:
- `metric_type` (optional): 指标类型，可选值: loss, ic, sharpe
- `last_n` (optional): 返回最近N个数据点，默认100

**响应示例**:
```json
{
  "success": true,
  "data": {
    "metrics_history": [
      {
        "epoch": 125,
        "timestamp": "2024-01-15T10:25:00Z",
        "train_loss": 0.245,
        "valid_loss": 0.278,
        "train_ic": 0.0441,
        "valid_ic": 0.0387
      }
    ]
  }
}
```

#### 4.2.3 停止训练任务

**接口路径**: `POST /models/train/{task_id}/stop`

**响应示例**:
```json
{
  "success": true,
  "message": "训练任务停止请求已提交"
}
```

### 4.3 模型评估接口

#### 4.3.1 获取模型详细性能

**接口路径**: `GET /models/{model_id}/performance`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "model_info": {
      "id": "model_001",
      "name": "LightGBM-Alpha158",
      "type": "lightgbm"
    },
    "performance_metrics": {
      "training": {
        "ic": 0.0456,
        "ic_ir": 1.95,
        "rank_ic": 0.0612,
        "rank_ic_ir": 2.15
      },
      "validation": {
        "ic": 0.0398,
        "ic_ir": 1.75,
        "rank_ic": 0.0534,
        "rank_ic_ir": 1.89
      },
      "test": {
        "ic": 0.0367,
        "ic_ir": 1.65,
        "rank_ic": 0.0489,
        "rank_ic_ir": 1.78
      }
    },
    "feature_importance": [
      {
        "feature_name": "RESI5",
        "importance": 0.125,
        "rank": 1
      },
      {
        "feature_name": "WVMA5", 
        "importance": 0.098,
        "rank": 2
      }
    ]
  }
}
```

### 4.4 模型对比接口

#### 4.4.1 模型对比分析

**接口路径**: `POST /models/compare`

**请求体**:
```json
{
  "model_ids": ["model_001", "model_002", "model_003"],
  "metrics": ["ic", "ic_ir", "sharpe", "max_drawdown"],
  "comparison_period": {
    "start_date": "2023-01-01",
    "end_date": "2023-12-31"
  }
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "comparison_table": [
      {
        "model_id": "model_001",
        "model_name": "LightGBM-Alpha158",
        "ic": 0.0367,
        "ic_ir": 1.65,
        "sharpe": 1.23,
        "max_drawdown": -0.085
      }
    ],
    "ranking": [
      {
        "metric": "ic",
        "ranking": [
          {"model_id": "model_001", "value": 0.0367, "rank": 1}
        ]
      }
    ]
  }
}
```

### 4.5 模型部署接口

#### 4.5.1 部署模型

**接口路径**: `POST /models/{model_id}/deploy`

**请求体**:
```json
{
  "deployment_name": "prod-model-v1",
  "environment": "production",
  "config": {
    "instance_type": "gpu.medium",
    "auto_scaling": true,
    "min_instances": 1,
    "max_instances": 5
  }
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "deployment_id": "deploy_001",
    "endpoint_url": "https://api.qlib.com/predict/prod-model-v1",
    "status": "deploying",
    "estimated_ready_time": "2024-01-15T11:15:00Z"
  }
}
```

---

## 5. 策略回测页面 (StrategyBacktest) API

### 5.1 策略配置接口

#### 5.1.1 获取可用策略类型

**接口路径**: `GET /strategies/types`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "strategy_types": [
      {
        "type": "topk",
        "name": "TopK策略",
        "description": "选择评分最高的K只股票",
        "parameters": [
          {
            "name": "topk",
            "type": "integer",
            "default": 30,
            "min": 5,
            "max": 100,
            "description": "选股数量"
          }
        ]
      },
      {
        "type": "long_short",
        "name": "多空策略",
        "description": "同时做多和做空",
        "parameters": [
          {
            "name": "long_ratio",
            "type": "float",
            "default": 0.5,
            "min": 0.1,
            "max": 0.9,
            "description": "做多比例"
          }
        ]
      }
    ]
  }
}
```

### 5.2 回测执行接口

#### 5.2.1 创建回测任务

**接口路径**: `POST /backtest/create`

**请求体**:
```json
{
  "name": "TopK-Strategy-CSI300",
  "model_id": "model_001",
  "dataset_id": "dataset_001",
  "strategy": {
    "type": "topk",
    "parameters": {
      "topk": 30,
      "rebalance_frequency": "daily"
    }
  },
  "backtest_config": {
    "start_date": "2023-01-01",
    "end_date": "2023-12-31",
    "initial_cash": 1000000,
    "benchmark": "SH000300",
    "trading_costs": {
      "commission": 0.0003,
      "impact_cost": 0.0005,
      "min_commission": 5
    }
  },
  "risk_management": {
    "max_position_weight": 0.05,
    "stop_loss": -0.1,
    "max_turnover": 0.5
  }
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "backtest_id": "backtest_001",
    "task_id": "bt_task_001",
    "message": "回测任务已创建",
    "estimated_duration": 600
  }
}
```

#### 5.2.2 获取回测进度

**接口路径**: `GET /backtest/{task_id}/progress`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "task_id": "bt_task_001",
    "status": "running",
    "progress": 75,
    "current_date": "2023-09-15",
    "total_days": 365,
    "processed_days": 274,
    "estimated_remaining_time": 150,
    "current_metrics": {
      "total_return": 0.156,
      "max_drawdown": -0.045,
      "current_positions": 28
    }
  }
}
```

### 5.3 回测结果接口

#### 5.3.1 获取回测结果概览

**接口路径**: `GET /backtest/{backtest_id}/results`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "backtest_info": {
      "id": "backtest_001",
      "name": "TopK-Strategy-CSI300",
      "status": "completed",
      "period": "2023-01-01 to 2023-12-31"
    },
    "performance_summary": {
      "total_return": 0.235,
      "annualized_return": 0.182,
      "benchmark_return": 0.096,
      "excess_return": 0.139,
      "volatility": 0.124,
      "sharpe_ratio": 1.85,
      "information_ratio": 1.12,
      "max_drawdown": -0.085,
      "calmar_ratio": 2.14,
      "win_rate": 0.623
    },
    "risk_metrics": {
      "var_95": -0.025,
      "cvar_95": -0.041,
      "beta": 0.95,
      "alpha": 0.089,
      "tracking_error": 0.124
    },
    "trading_statistics": {
      "total_trades": 2341,
      "avg_holding_period": 8.7,
      "turnover_rate": 0.234,
      "trading_cost": 0.0156,
      "profit_factor": 2.1
    }
  }
}
```

#### 5.3.2 获取净值曲线数据

**接口路径**: `GET /backtest/{backtest_id}/nav-curve`

**请求参数**:
- `frequency` (optional): 数据频率，可选值: daily, weekly, monthly，默认daily

**响应示例**:
```json
{
  "success": true,
  "data": {
    "nav_data": [
      {
        "date": "2023-01-01",
        "strategy_nav": 1.0,
        "benchmark_nav": 1.0,
        "excess_nav": 1.0,
        "positions": 30,
        "turnover": 0.0
      },
      {
        "date": "2023-01-02",
        "strategy_nav": 1.012,
        "benchmark_nav": 1.008,
        "excess_nav": 1.004,
        "positions": 30,
        "turnover": 0.05
      }
    ]
  }
}
```

#### 5.3.3 获取持仓明细

**接口路径**: `GET /backtest/{backtest_id}/positions`

**请求参数**:
- `date` (optional): 指定日期，格式YYYY-MM-DD
- `page` (optional): 页码，默认1
- `limit` (optional): 每页数量，默认20

**响应示例**:
```json
{
  "success": true,
  "data": {
    "date": "2023-06-15",
    "positions": [
      {
        "symbol": "000001.SZ",
        "name": "平安银行",
        "weight": 0.045,
        "shares": 15000,
        "market_value": 450000,
        "cost_basis": 29.85,
        "current_price": 30.12,
        "unrealized_pnl": 4050,
        "holding_days": 5
      }
    ],
    "portfolio_summary": {
      "total_positions": 30,
      "total_market_value": 9850000,
      "cash": 150000,
      "leverage": 0.985
    }
  }
}
```

#### 5.3.4 获取交易记录

**接口路径**: `GET /backtest/{backtest_id}/trades`

**请求参数**:
- `start_date` (optional): 开始日期
- `end_date` (optional): 结束日期
- `symbol` (optional): 股票代码筛选
- `action` (optional): 交易类型筛选，可选值: buy, sell
- `page` (optional): 页码
- `limit` (optional): 每页数量

**响应示例**:
```json
{
  "success": true,
  "data": {
    "trades": [
      {
        "trade_id": "trade_001",
        "date": "2023-06-15",
        "symbol": "000001.SZ",
        "name": "平安银行",
        "action": "buy",
        "quantity": 5000,
        "price": 30.05,
        "amount": 150250,
        "commission": 45.08,
        "reason": "new_position"
      }
    ],
    "pagination": {
      "total": 2341,
      "page": 1,
      "limit": 20
    }
  }
}
```

### 5.4 回测分析接口

#### 5.4.1 获取收益归因分析

**接口路径**: `GET /backtest/{backtest_id}/attribution`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "time_period": "2023-01-01 to 2023-12-31",
    "attribution_analysis": {
      "factor_attribution": [
        {
          "factor": "momentum",
          "contribution": 0.085,
          "percentage": 36.5
        },
        {
          "factor": "value",
          "contribution": 0.042,
          "percentage": 18.1
        }
      ],
      "sector_attribution": [
        {
          "sector": "金融",
          "weight": 0.25,
          "benchmark_weight": 0.32,
          "active_weight": -0.07,
          "contribution": -0.012
        }
      ],
      "timing_effect": 0.023,
      "selection_effect": 0.089,
      "interaction_effect": 0.005
    }
  }
}
```

#### 5.4.2 获取风险分析报告

**接口路径**: `GET /backtest/{backtest_id}/risk-analysis`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "risk_metrics": {
      "market_risk": {
        "beta": 0.95,
        "correlation": 0.89,
        "r_squared": 0.79
      },
      "concentration_risk": {
        "top_10_weight": 0.45,
        "herfindahl_index": 0.045,
        "effective_positions": 22.3
      },
      "liquidity_risk": {
        "avg_turnover_days": 2.3,
        "liquidity_score": 8.5,
        "impact_cost": 0.0008
      }
    },
    "drawdown_analysis": {
      "max_drawdown": -0.085,
      "max_drawdown_duration": 23,
      "recovery_time": 15,
      "underwater_periods": [
        {
          "start_date": "2023-03-15",
          "end_date": "2023-04-07",
          "duration": 23,
          "max_drawdown": -0.085
        }
      ]
    }
  }
}
```

---

## 6. qlib工作流页面 (QlibWorkflow) API

### 6.1 qlib工作流配置接口

#### 6.1.1 获取qlib预设配置

**接口路径**: `GET /qlib/presets`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "datasets": [
      {
        "value": "csi300",
        "label": "CSI300 - 沪深300成分股",
        "description": "qlib内置中国A股主要指数"
      }
    ],
    "handlers": [
      {
        "value": "Alpha158",
        "label": "Alpha158 - qlib经典158因子",
        "description": "包含价格、成交量、技术指标等158个因子"
      }
    ],
    "models": [
      {
        "value": "LightGBM",
        "label": "LightGBM - 梯度提升树",
        "description": "qlib优化的LightGBM实现，适合表格数据"
      }
    ],
    "strategies": [
      {
        "value": "TopkDropoutStrategy",
        "label": "TopK选股策略",
        "description": "qlib经典的TopK选股+Dropout策略"
      }
    ]
  }
}
```

#### 6.1.2 验证qlib配置

**接口路径**: `POST /qlib/validate-config`

**请求体**:
```json
{
  "data": {
    "provider_uri": "~/.qlib/qlib_data/cn_data",
    "region": "cn",
    "market": "csi300",
    "start_time": "2020-01-01",
    "end_time": "2023-12-31"
  },
  "features": {
    "handler": "Alpha158",
    "factors": [],
    "label": "Ref($close, -1) / $close - 1"
  },
  "model": {
    "class": "LightGBM",
    "params": {
      "n_estimators": 200,
      "learning_rate": 0.1,
      "max_depth": 6
    }
  }
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "is_valid": true,
    "warnings": [
      "建议增加验证集的时间跨度以提高模型泛化能力"
    ],
    "errors": [],
    "estimated_execution_time": 1800
  }
}
```

### 6.2 qlib工作流执行接口

#### 6.2.1 启动qlib工作流

**接口路径**: `POST /qlib/workflow/run`

**请求体**:
```json
{
  "workflow_name": "LightGBM-Alpha158-CSI300",
  "config": {
    "qlib_init": {
      "provider_uri": "~/.qlib/qlib_data/cn_data",
      "region": "cn"
    },
    "market": "csi300",
    "benchmark": "SH000300",
    "data_handler_config": {
      "start_time": "2018-01-01",
      "end_time": "2023-12-31",
      "instruments": "market",
      "label": ["Ref($close, -1) / $close - 1"]
    },
    "task": {
      "model": {
        "class": "LightGBM",
        "kwargs": {
          "n_estimators": 200,
          "learning_rate": 0.1,
          "max_depth": 6,
          "seed": 2024
        }
      },
      "dataset": {
        "class": "DatasetH",
        "kwargs": {
          "handler": {
            "class": "Alpha158"
          },
          "segments": {
            "train": ["2018-01-01", "2021-12-31"],
            "valid": ["2022-01-01", "2022-12-31"],
            "test": ["2023-01-01", "2023-12-31"]
          }
        }
      }
    },
    "port_analysis_config": {
      "strategy": {
        "class": "TopkDropoutStrategy",
        "kwargs": {
          "topk": 50,
          "n_drop": 5
        }
      },
      "backtest": {
        "start_time": "2023-01-01",
        "end_time": "2023-12-31",
        "account": 100000000,
        "benchmark": "SH000300"
      }
    }
  }
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "workflow_id": "workflow_001",
    "task_id": "qlib_task_001",
    "message": "qlib工作流已启动",
    "estimated_duration": 1800,
    "config_file": "/tmp/qlib_config_20240115.yaml"
  }
}
```

#### 6.2.2 获取qlib工作流进度

**接口路径**: `GET /qlib/workflow/{task_id}/progress`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "task_id": "qlib_task_001",
    "status": "running",
    "progress": 60,
    "current_stage": "model_training",
    "stages": [
      {
        "name": "data_loading",
        "status": "completed",
        "progress": 100,
        "start_time": "2024-01-15T10:00:00Z",
        "end_time": "2024-01-15T10:05:00Z"
      },
      {
        "name": "feature_engineering",
        "status": "completed", 
        "progress": 100,
        "start_time": "2024-01-15T10:05:00Z",
        "end_time": "2024-01-15T10:15:00Z"
      },
      {
        "name": "model_training",
        "status": "running",
        "progress": 60,
        "start_time": "2024-01-15T10:15:00Z",
        "estimated_end_time": "2024-01-15T10:45:00Z"
      }
    ],
    "logs": [
      "[2024-01-15 10:30:00] 正在训练模型... 进度: 60%",
      "[2024-01-15 10:25:00] 特征工程完成，共计算158个因子"
    ]
  }
}
```

#### 6.2.3 获取qlib工作流结果

**接口路径**: `GET /qlib/workflow/{task_id}/result`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "workflow_id": "workflow_001",
    "status": "completed",
    "execution_time": 1650,
    "results": {
      "model_performance": {
        "train_ic": 0.0456,
        "valid_ic": 0.0398,
        "test_ic": 0.0367,
        "train_rank_ic": 0.0612,
        "valid_rank_ic": 0.0534,
        "test_rank_ic": 0.0489,
        "model_path": "/qlib/models/LightGBM_20240115.pkl"
      },
      "strategy_performance": {
        "annual_return": 0.1847,
        "benchmark_return": 0.0956,
        "excess_return": 0.0891,
        "volatility": 0.1623,
        "sharpe_ratio": 1.138,
        "information_ratio": 0.549,
        "max_drawdown": -0.0847,
        "win_rate": 0.574,
        "calmar_ratio": 2.18
      },
      "backtest_details": {
        "total_trades": 2341,
        "avg_holding_days": 8.7,
        "turnover_rate": 0.234,
        "trading_cost": 0.0156,
        "net_return": 0.1691
      },
      "factor_analysis": {
        "top_factors": [
          {
            "name": "RESI5",
            "ic": 0.0423,
            "weight": 0.125
          },
          {
            "name": "WVMA5",
            "ic": 0.0389,
            "weight": 0.098
          }
        ]
      }
    },
    "output_files": {
      "model_file": "/qlib/models/LightGBM_20240115.pkl",
      "predictions": "/qlib/outputs/predictions.csv",
      "portfolio": "/qlib/outputs/portfolio.csv",
      "analysis_report": "/qlib/outputs/analysis_report.html"
    }
  }
}
```

### 6.3 qlib工作流管理接口

#### 6.3.1 获取工作流历史

**接口路径**: `GET /qlib/workflow/history`

**请求参数**:
- `page` (optional): 页码，默认1
- `limit` (optional): 每页数量，默认10
- `status` (optional): 状态筛选

**响应示例**:
```json
{
  "success": true,
  "data": {
    "workflows": [
      {
        "workflow_id": "workflow_001",
        "name": "LightGBM-Alpha158-CSI300",
        "status": "completed",
        "create_time": "2024-01-15T10:00:00Z",
        "execution_time": 1650,
        "config_summary": {
          "market": "csi300",
          "model": "LightGBM",
          "handler": "Alpha158"
        },
        "performance_summary": {
          "annual_return": 0.1847,
          "sharpe_ratio": 1.138,
          "test_ic": 0.0367
        }
      }
    ],
    "pagination": {
      "total": 25,
      "page": 1,
      "limit": 10
    }
  }
}
```

#### 6.3.2 停止qlib工作流

**接口路径**: `POST /qlib/workflow/{task_id}/stop`

**响应示例**:
```json
{
  "success": true,
  "message": "工作流停止请求已提交"
}
```

---

## 7. 结果分析页面 (ResultsAnalysis) API

### 7.1 分析概览接口

#### 7.1.1 获取分析结果概览

**接口路径**: `GET /analysis/overview`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "summary_statistics": {
      "total_results": 25,
      "avg_annual_return": 0.1654,
      "avg_sharpe_ratio": 1.423,
      "avg_test_ic": 0.0389,
      "best_return": 0.2134,
      "worst_drawdown": -0.1023
    },
    "recent_results": [
      {
        "id": "result_001",
        "name": "LightGBM-Alpha158-CSI300",
        "type": "qlib_workflow",
        "date": "2024-01-15",
        "annual_return": 0.1847,
        "sharpe_ratio": 1.138,
        "test_ic": 0.0367,
        "max_drawdown": -0.0847
      }
    ]
  }
}
```

### 7.2 模型分析接口

#### 7.2.1 获取模型性能对比

**接口路径**: `POST /analysis/models/compare`

**请求体**:
```json
{
  "result_ids": ["result_001", "result_002", "result_003"],
  "metrics": ["test_ic", "valid_ic", "train_ic", "rank_ic"]
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "comparison_data": [
      {
        "result_id": "result_001",
        "name": "LightGBM-Alpha158-CSI300",
        "model_type": "LightGBM",
        "test_ic": 0.0367,
        "valid_ic": 0.0398,
        "train_ic": 0.0456,
        "rank_ic": 0.0489
      }
    ],
    "ranking": {
      "test_ic": [
        {"result_id": "result_001", "rank": 1, "value": 0.0367}
      ]
    }
  }
}
```

#### 7.2.2 获取因子重要性分析

**接口路径**: `GET /analysis/models/{result_id}/factor-importance`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "model_info": {
      "result_id": "result_001",
      "model_type": "LightGBM",
      "feature_handler": "Alpha158"
    },
    "factor_importance": [
      {
        "factor_name": "RESI5",
        "importance": 0.125,
        "rank": 1,
        "ic": 0.0423,
        "description": "残差相关因子"
      }
    ],
    "importance_chart_data": {
      "labels": ["RESI5", "WVMA5", "RSQR10"],
      "values": [0.125, 0.098, 0.087]
    }
  }
}
```

### 7.3 策略分析接口

#### 7.3.1 获取策略绩效分析

**接口路径**: `GET /analysis/strategies/{result_id}/performance`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "strategy_info": {
      "result_id": "result_001",
      "strategy_type": "TopkDropoutStrategy",
      "parameters": {
        "topk": 50,
        "n_drop": 5
      }
    },
    "performance_metrics": {
      "return_metrics": {
        "annual_return": 0.1847,
        "benchmark_return": 0.0956,
        "excess_return": 0.0891,
        "total_return": 0.2341
      },
      "risk_metrics": {
        "volatility": 0.1623,
        "sharpe_ratio": 1.138,
        "information_ratio": 0.549,
        "max_drawdown": -0.0847,
        "calmar_ratio": 2.18
      },
      "trading_metrics": {
        "win_rate": 0.574,
        "profit_factor": 2.1,
        "avg_holding_period": 8.7,
        "turnover_rate": 0.234
      }
    },
    "performance_attribution": {
      "alpha": 0.089,
      "beta": 0.95,
      "market_timing": 0.023,
      "stock_selection": 0.089
    }
  }
}
```

### 7.4 对比分析接口

#### 7.4.1 多策略对比分析

**接口路径**: `POST /analysis/strategies/compare`

**请求体**:
```json
{
  "result_ids": ["result_001", "result_002", "result_003"],
  "comparison_metrics": [
    "annual_return",
    "sharpe_ratio",
    "max_drawdown",
    "information_ratio"
  ],
  "benchmark": "SH000300"
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "comparison_table": [
      {
        "result_id": "result_001",
        "name": "LightGBM-Alpha158-CSI300",
        "annual_return": 0.1847,
        "sharpe_ratio": 1.138,
        "max_drawdown": -0.0847,
        "information_ratio": 0.549
      }
    ],
    "performance_ranking": {
      "annual_return": [
        {"result_id": "result_001", "rank": 1, "value": 0.1847}
      ]
    },
    "risk_return_scatter": [
      {
        "result_id": "result_001",
        "x": 0.1623,
        "y": 0.1847,
        "label": "LightGBM-Alpha158"
      }
    ]
  }
}
```

### 7.5 报告生成接口

#### 7.5.1 生成分析报告

**接口路径**: `POST /analysis/reports/generate`

**请求体**:
```json
{
  "report_type": "comprehensive",
  "result_ids": ["result_001", "result_002"],
  "sections": [
    "executive_summary",
    "model_analysis",
    "strategy_performance",
    "risk_analysis",
    "recommendations"
  ],
  "format": "pdf",
  "language": "zh"
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "report_id": "report_001",
    "task_id": "report_task_001",
    "message": "报告生成任务已启动",
    "estimated_completion_time": "2024-01-15T11:15:00Z"
  }
}
```

#### 7.5.2 获取报告生成状态

**接口路径**: `GET /analysis/reports/{task_id}/status`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "task_id": "report_task_001",
    "status": "completed",
    "progress": 100,
    "report_url": "/api/v1/analysis/reports/report_001/download",
    "file_size": 2048576,
    "page_count": 35
  }
}
```

#### 7.5.3 下载分析报告

**接口路径**: `GET /analysis/reports/{report_id}/download`

**响应**: 直接返回文件流

---

## 8. WebSocket 实时通信接口

### 8.1 任务状态推送

**连接地址**: `ws://localhost:8000/ws/task/{task_id}`

**消息格式**:
```json
{
  "type": "task_status",
  "task_id": "task_001",
  "status": "running",
  "progress": 65,
  "message": "正在训练模型...",
  "data": {
    "current_epoch": 130,
    "total_epochs": 200,
    "metrics": {
      "loss": 0.234,
      "ic": 0.0456
    }
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### 8.2 系统状态推送

**连接地址**: `ws://localhost:8000/ws/system`

**消息格式**:
```json
{
  "type": "system_status",
  "data": {
    "cpu_usage": 65,
    "memory_usage": 78,
    "active_tasks": 3,
    "queue_size": 2
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### 8.3 实时日志推送

**连接地址**: `ws://localhost:8000/ws/logs/{task_id}`

**消息格式**:
```json
{
  "type": "log",
  "task_id": "task_001",
  "level": "info",
  "message": "[2024-01-15 10:30:00] 训练完成，开始验证...",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

---

## 9. 通用工具接口

### 9.1 文件操作接口

#### 9.1.1 文件上传

**接口路径**: `POST /files/upload`

**请求**: multipart/form-data
- `file`: 文件
- `type` (optional): 文件类型，可选值: dataset, config, model

**响应示例**:
```json
{
  "success": true,
  "data": {
    "file_id": "file_001",
    "filename": "custom_dataset.csv",
    "size": 1024576,
    "url": "/api/v1/files/file_001/download"
  }
}
```

#### 9.1.2 文件下载

**接口路径**: `GET /files/{file_id}/download`

**响应**: 直接返回文件流

### 9.2 任务管理接口

#### 9.2.1 获取任务列表

**接口路径**: `GET /tasks`

**请求参数**:
- `status` (optional): 状态筛选
- `type` (optional): 类型筛选
- `page` (optional): 页码
- `limit` (optional): 每页数量

**响应示例**:
```json
{
  "success": true,
  "data": {
    "tasks": [
      {
        "task_id": "task_001",
        "name": "训练LightGBM模型",
        "type": "model_training",
        "status": "running",
        "progress": 65,
        "create_time": "2024-01-15T10:00:00Z",
        "estimated_end_time": "2024-01-15T10:45:00Z"
      }
    ],
    "pagination": {
      "total": 45,
      "page": 1,
      "limit": 10
    }
  }
}
```

#### 9.2.2 取消任务

**接口路径**: `POST /tasks/{task_id}/cancel`

**响应示例**:
```json
{
  "success": true,
  "message": "任务取消请求已提交"
}
```

---

## 总结

本API文档涵盖了Qlib可视化平台的完整功能接口，包括：

1. **系统总览**: 统计数据、市场概览、性能图表
2. **数据管理**: 数据集管理、数据源配置、数据探索
3. **因子管理**: 因子工程、编辑器、分析、库管理
4. **模型训练**: 训练任务、监控、评估、对比、部署
5. **策略回测**: 配置、执行、结果分析、归因分析
6. **qlib工作流**: 端到端量化研究流程
7. **结果分析**: 概览、对比、报告生成
8. **WebSocket**: 实时通信和状态推送

每个接口都提供了详细的请求参数和响应格式，支持完整的量化投资研究工作流程。API设计遵循RESTful风格，支持分页、筛选、排序等通用功能。

**修改原因概述**:
- 完全重写了API文档，基于前端页面功能分析设计了对应的API接口规范
- 按照页面功能模块进行API分组，每个模块包含相应的CRUD操作和专业功能接口
- 增加了WebSocket实时通信接口，支持任务状态、系统监控和日志推送
- 提供了详细的请求参数、响应格式和功能描述
- 涵盖了从数据管理到结果分析的完整量化投资研究流程API