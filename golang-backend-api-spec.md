# Qlib Golang后端API规范

## 项目结构

```
qlib-backend/
├── main.go                 # 主入口
├── config/                 # 配置管理
│   ├── config.go
│   └── qlib.yaml
├── internal/               # 内部包
│   ├── api/               # API路由和控制器
│   │   ├── handlers/      # 请求处理器
│   │   ├── middleware/    # 中间件
│   │   └── routes/        # 路由定义
│   ├── models/            # 数据模型
│   ├── services/          # 业务逻辑服务
│   ├── qlib/              # Qlib Python接口封装
│   └── utils/             # 工具函数
├── pkg/                   # 公共包
├── scripts/               # 脚本文件
└── docker/                # Docker配置
```

## API接口规范

### 1. 基础响应格式

```go
type Response struct {
    Code    int         `json:"code"`    // 0表示成功，非0表示错误
    Message string      `json:"message"` // 响应消息
    Data    interface{} `json:"data"`    // 响应数据
    Total   int64       `json:"total,omitempty"` // 总数（分页时使用）
}

type ErrorResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Error   string `json:"error,omitempty"`
}
```

### 2. 数据集管理API

#### 2.1 数据集模型

```go
type Dataset struct {
    ID          string    `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"not null"`
    Description string    `json:"description"`
    Universe    string    `json:"universe"`    // 股票池
    Fields      []string  `json:"fields" gorm:"serializer:json"`
    StartDate   string    `json:"start_date"`
    EndDate     string    `json:"end_date"`
    Status      string    `json:"status"`     // preparing, ready, error
    Samples     int64     `json:"samples"`
    Features    int       `json:"features"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type CreateDatasetRequest struct {
    Name        string   `json:"name" binding:"required"`
    Description string   `json:"description"`
    Universe    string   `json:"universe" binding:"required"`
    Fields      []string `json:"fields" binding:"required"`
    StartDate   string   `json:"start_date" binding:"required"`
    EndDate     string   `json:"end_date" binding:"required"`
}
```

#### 2.2 接口实现

```go
// GET /api/v1/datasets
func GetDatasets(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    status := c.Query("status")
    
    // 调用服务层逻辑
    datasets, total, err := services.GetDatasets(page, limit, status)
    if err != nil {
        c.JSON(500, ErrorResponse{Code: 500, Message: "获取数据集失败", Error: err.Error()})
        return
    }
    
    c.JSON(200, Response{Code: 0, Message: "success", Data: datasets, Total: total})
}

// POST /api/v1/datasets
func CreateDataset(c *gin.Context) {
    var req CreateDatasetRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, ErrorResponse{Code: 400, Message: "参数错误", Error: err.Error()})
        return
    }
    
    dataset, err := services.CreateDataset(&req)
    if err != nil {
        c.JSON(500, ErrorResponse{Code: 500, Message: "创建数据集失败", Error: err.Error()})
        return
    }
    
    c.JSON(201, Response{Code: 0, Message: "数据集创建成功", Data: dataset})
}
```

### 3. 因子管理API

#### 3.1 因子模型

```go
type Factor struct {
    ID          string    `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"not null"`
    Expression  string    `json:"expression" gorm:"not null"`
    Description string    `json:"description"`
    Category    string    `json:"category"`
    Status      string    `json:"status"`     // active, inactive
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type TestFactorRequest struct {
    Expression string `json:"expression" binding:"required"`
    DatasetID  string `json:"dataset_id" binding:"required"`
    StartDate  string `json:"start_date" binding:"required"`
    EndDate    string `json:"end_date" binding:"required"`
}

type FactorTestResult struct {
    IC          float64 `json:"ic"`
    ICIR        float64 `json:"ic_ir"`
    RankIC      float64 `json:"rank_ic"`
    RankICIR    float64 `json:"rank_ic_ir"`
    Turnover    float64 `json:"turnover"`
    Coverage    float64 `json:"coverage"`
    ValidPeriods int    `json:"valid_periods"`
}
```

#### 3.2 接口实现

```go
// POST /api/v1/factors/test
func TestFactor(c *gin.Context) {
    var req TestFactorRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, ErrorResponse{Code: 400, Message: "参数错误", Error: err.Error()})
        return
    }
    
    // 调用Qlib Python接口进行因子测试
    result, err := qlib.TestFactor(req.Expression, req.DatasetID, req.StartDate, req.EndDate)
    if err != nil {
        c.JSON(500, ErrorResponse{Code: 500, Message: "因子测试失败", Error: err.Error()})
        return
    }
    
    c.JSON(200, Response{Code: 0, Message: "因子测试完成", Data: result})
}

// POST /api/v1/factors/validate
func ValidateFactorExpression(c *gin.Context) {
    var req struct {
        Expression string `json:"expression" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, ErrorResponse{Code: 400, Message: "参数错误", Error: err.Error()})
        return
    }
    
    isValid, errors := qlib.ValidateExpression(req.Expression)
    
    c.JSON(200, Response{
        Code: 0, 
        Message: "验证完成", 
        Data: map[string]interface{}{
            "valid": isValid,
            "errors": errors,
        },
    })
}
```

### 4. 模型管理API

#### 4.1 模型相关结构

```go
type Model struct {
    ID          string    `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"not null"`
    Type        string    `json:"type"`       // XGBoost, LightGBM, etc.
    DatasetID   string    `json:"dataset_id"`
    FactorIDs   []string  `json:"factor_ids" gorm:"serializer:json"`
    Config      ModelConfig `json:"config" gorm:"serializer:json"`
    Status      string    `json:"status"`     // training, trained, error
    Metrics     ModelMetrics `json:"metrics" gorm:"serializer:json"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type ModelConfig struct {
    Objective      string                 `json:"objective"`
    NumRounds      int                    `json:"num_rounds"`
    LearningRate   float64               `json:"learning_rate"`
    MaxDepth       int                   `json:"max_depth"`
    SubSample      float64               `json:"subsample"`
    ColSampleByTree float64              `json:"colsample_bytree"`
    Params         map[string]interface{} `json:"params"`
}

type ModelMetrics struct {
    IC          float64 `json:"ic"`
    ICIR        float64 `json:"ic_ir"`
    RankIC      float64 `json:"rank_ic"`
    RankICIR    float64 `json:"rank_ic_ir"`
    Sharpe      float64 `json:"sharpe"`
    TrainTime   string  `json:"train_time"`
}

type TrainModelRequest struct {
    ModelID     string      `json:"model_id" binding:"required"`
    TrainConfig TrainConfig `json:"train_config"`
}

type TrainConfig struct {
    StartDate    string `json:"start_date"`
    EndDate      string `json:"end_date"`
    ValidationRatio float64 `json:"validation_ratio"`
    EarlyStop    int    `json:"early_stop"`
}
```

### 5. 回测管理API

#### 5.1 回测相关结构

```go
type Backtest struct {
    ID          string    `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"not null"`
    ModelID     string    `json:"model_id"`
    DatasetID   string    `json:"dataset_id"`
    Strategy    string    `json:"strategy"`    // topk, long_short, market_neutral
    Config      BacktestConfig `json:"config" gorm:"serializer:json"`
    Status      string    `json:"status"`      // pending, running, completed, error
    Results     *BacktestResults `json:"results,omitempty" gorm:"serializer:json"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type BacktestConfig struct {
    StartDate    string  `json:"start_date"`
    EndDate      string  `json:"end_date"`
    InitialCash  float64 `json:"initial_cash"`
    TopK         int     `json:"top_k"`
    RebalanceFreq string `json:"rebalance_freq"`
    Commission   float64 `json:"commission"`
    ImpactCost   float64 `json:"impact_cost"`
}

type BacktestResults struct {
    TotalReturn    float64 `json:"total_return"`
    AnnualReturn   float64 `json:"annual_return"`
    SharpeRatio    float64 `json:"sharpe_ratio"`
    CalmarRatio    float64 `json:"calmar_ratio"`
    MaxDrawdown    float64 `json:"max_drawdown"`
    Volatility     float64 `json:"volatility"`
    WinRate        float64 `json:"win_rate"`
    TotalTrades    int     `json:"total_trades"`
    ProfitFactor   float64 `json:"profit_factor"`
}
```

### 6. 任务管理API

#### 6.1 任务相关结构

```go
type Task struct {
    ID          string    `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"not null"`
    Type        string    `json:"type"`        // dataset, factor_test, model_train, backtest
    Status      string    `json:"status"`      // pending, running, completed, error, cancelled
    Progress    int       `json:"progress"`    // 0-100
    Message     string    `json:"message"`
    Error       string    `json:"error,omitempty"`
    Config      map[string]interface{} `json:"config" gorm:"serializer:json"`
    Results     map[string]interface{} `json:"results,omitempty" gorm:"serializer:json"`
    StartTime   *time.Time `json:"start_time,omitempty"`
    EndTime     *time.Time `json:"end_time,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type TaskLog struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    TaskID    string    `json:"task_id"`
    Level     string    `json:"level"`     // info, warning, error
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
}
```

### 7. Qlib Python接口封装

```go
package qlib

import (
    "encoding/json"
    "os/exec"
    "fmt"
)

type QlibService struct {
    PythonPath string
    ScriptPath string
}

func NewQlibService(pythonPath, scriptPath string) *QlibService {
    return &QlibService{
        PythonPath: pythonPath,
        ScriptPath: scriptPath,
    }
}

// 执行Python脚本
func (q *QlibService) ExecutePythonScript(scriptName string, args map[string]interface{}) (map[string]interface{}, error) {
    argsJSON, _ := json.Marshal(args)
    
    cmd := exec.Command(q.PythonPath, fmt.Sprintf("%s/%s", q.ScriptPath, scriptName), string(argsJSON))
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("执行Python脚本失败: %v", err)
    }
    
    var result map[string]interface{}
    if err := json.Unmarshal(output, &result); err != nil {
        return nil, fmt.Errorf("解析Python脚本输出失败: %v", err)
    }
    
    return result, nil
}

// 测试因子
func (q *QlibService) TestFactor(expression, datasetID, startDate, endDate string) (*FactorTestResult, error) {
    args := map[string]interface{}{
        "expression":  expression,
        "dataset_id":  datasetID,
        "start_date":  startDate,
        "end_date":    endDate,
    }
    
    result, err := q.ExecutePythonScript("test_factor.py", args)
    if err != nil {
        return nil, err
    }
    
    // 转换结果为Go结构体
    resultJSON, _ := json.Marshal(result)
    var factorResult FactorTestResult
    json.Unmarshal(resultJSON, &factorResult)
    
    return &factorResult, nil
}

// 训练模型
func (q *QlibService) TrainModel(modelConfig map[string]interface{}) (string, error) {
    result, err := q.ExecutePythonScript("train_model.py", modelConfig)
    if err != nil {
        return "", err
    }
    
    taskID, ok := result["task_id"].(string)
    if !ok {
        return "", fmt.Errorf("获取任务ID失败")
    }
    
    return taskID, nil
}

// 运行回测
func (q *QlibService) RunBacktest(backtestConfig map[string]interface{}) (string, error) {
    result, err := q.ExecutePythonScript("run_backtest.py", backtestConfig)
    if err != nil {
        return "", err
    }
    
    taskID, ok := result["task_id"].(string)
    if !ok {
        return "", fmt.Errorf("获取任务ID失败")
    }
    
    return taskID, nil
}

// 验证表达式
func (q *QlibService) ValidateExpression(expression string) (bool, []string) {
    args := map[string]interface{}{
        "expression": expression,
    }
    
    result, err := q.ExecutePythonScript("validate_expression.py", args)
    if err != nil {
        return false, []string{err.Error()}
    }
    
    valid, _ := result["valid"].(bool)
    errorsInterface, _ := result["errors"].([]interface{})
    
    errors := make([]string, len(errorsInterface))
    for i, e := range errorsInterface {
        errors[i] = e.(string)
    }
    
    return valid, errors
}
```

### 8. WebSocket支持

```go
package websocket

import (
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "net/http"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // 允许跨域
    },
}

// 任务进度WebSocket
func TaskProgressWS(c *gin.Context) {
    taskID := c.Param("taskId")
    
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer conn.Close()
    
    // 订阅任务进度更新
    progressChan := services.SubscribeTaskProgress(taskID)
    
    for progress := range progressChan {
        if err := conn.WriteJSON(progress); err != nil {
            break
        }
    }
}

// 系统状态WebSocket
func SystemStatusWS(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer conn.Close()
    
    statusChan := services.SubscribeSystemStatus()
    
    for status := range statusChan {
        if err := conn.WriteJSON(status); err != nil {
            break
        }
    }
}
```

### 9. 中间件

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "net/http"
    "time"
)

// CORS中间件
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusOK)
            return
        }
        
        c.Next()
    }
}

// 请求日志中间件
func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        
        duration := time.Since(start)
        statusCode := c.Writer.Status()
        
        // 记录请求日志
        log.Printf("[%s] %s %s %d %v",
            c.Request.Method,
            c.Request.RequestURI,
            c.ClientIP(),
            statusCode,
            duration,
        )
    }
}

// 认证中间件
func Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, ErrorResponse{Code: 401, Message: "未授权访问"})
            c.Abort()
            return
        }
        
        // 验证token
        userID, err := services.ValidateToken(token)
        if err != nil {
            c.JSON(401, ErrorResponse{Code: 401, Message: "无效token"})
            c.Abort()
            return
        }
        
        c.Set("user_id", userID)
        c.Next()
    }
}
```

### 10. 配置文件示例

```yaml
# config/qlib.yaml
server:
  port: 8080
  mode: debug

database:
  driver: sqlite
  dsn: qlib.db

qlib:
  python_path: /usr/bin/python3
  script_path: ./python_scripts
  data_path: ./qlib_data
  
cors:
  allowed_origins:
    - "http://localhost:3000"
    - "http://localhost:8080"
  
logging:
  level: info
  file: logs/qlib-backend.log

auth:
  jwt_secret: your-secret-key
  token_expire: 24h
```

这个API规范为您的Golang后端提供了完整的接口设计，包括数据结构、请求处理、Python脚本调用和WebSocket支持。您可以基于这个规范实现完整的qlib可视化平台后端服务。