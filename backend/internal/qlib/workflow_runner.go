package qlib

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// WorkflowRunner Qlib工作流执行器
type WorkflowRunner struct {
	client     *QlibClient
	scriptDir  string
	outputDir  string
}

// WorkflowConfig 工作流配置
type WorkflowConfig struct {
	// 数据配置
	Market        string                 `json:"market"`         // csi300, csi500, all
	StartTime     string                 `json:"start_time"`     // 开始时间
	EndTime       string                 `json:"end_time"`       // 结束时间
	Features      []string               `json:"features"`       // 特征列表
	Label         string                 `json:"label"`          // 标签
	
	// 模型配置
	Model         WFModelConfig            `json:"model"`          // 模型配置
	
	// 训练配置
	TrainPeriod   []string               `json:"train_period"`   // 训练时间段
	ValidPeriod   []string               `json:"valid_period"`   // 验证时间段
	TestPeriod    []string               `json:"test_period"`    // 测试时间段
	
	// 策略配置
	Strategy      WFStrategyConfig         `json:"strategy"`       // 策略配置
	
	// 回测配置
	BacktestConfig WFBacktestConfig        `json:"backtest"`       // 回测配置
	
	// 其他配置
	TaskConfig    map[string]interface{} `json:"task_config"`    // 任务相关配置
	RecorderConfig RecorderConfig        `json:"recorder"`       // 记录器配置
}

// WFModelConfig 工作流模型配置
type WFModelConfig struct {
	Class       string                 `json:"class"`        // 模型类名
	Module      string                 `json:"module"`       // 模块路径
	Args        map[string]interface{} `json:"args"`         // 模型参数
	Kwargs      map[string]interface{} `json:"kwargs"`       // 关键字参数
}

// WFStrategyConfig 工作流策略配置
type WFStrategyConfig struct {
	Class       string                 `json:"class"`        // 策略类名
	Module      string                 `json:"module"`       // 模块路径
	Args        map[string]interface{} `json:"args"`         // 策略参数
	Kwargs      map[string]interface{} `json:"kwargs"`       // 关键字参数
}

// WFBacktestConfig 工作流回测配置
type WFBacktestConfig struct {
	StartTime     string                 `json:"start_time"`
	EndTime       string                 `json:"end_time"`
	Account       float64                `json:"account"`      // 初始资金
	Benchmark     string                 `json:"benchmark"`    // 基准
	Exchange      ExchangeConfig         `json:"exchange"`     // 交易所配置
}

// ExchangeConfig 交易所配置
type ExchangeConfig struct {
	Freq          string                 `json:"freq"`         // 交易频率
	Limit_threshold float64             `json:"limit_threshold"`
	Deal_price    string                 `json:"deal_price"`   // 成交价格
	Open_cost     float64                `json:"open_cost"`    // 开仓成本
	Close_cost    float64                `json:"close_cost"`   // 平仓成本
	Trade_unit    int                    `json:"trade_unit"`   // 交易单位
}

// RecorderConfig 记录器配置
type RecorderConfig struct {
	Class       string                 `json:"class"`
	Module      string                 `json:"module"`
	Kwargs      map[string]interface{} `json:"kwargs"`
}

// WFResult 工作流执行结果
type WFResult struct {
	TaskID        string                 `json:"task_id"`
	Status        string                 `json:"status"`
	Progress      int                    `json:"progress"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       *time.Time             `json:"end_time,omitempty"`
	Results       map[string]interface{} `json:"results"`
	ModelMetrics  ModelMetrics           `json:"model_metrics"`
	BacktestResults BacktestResults      `json:"backtest_results"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
}

// ModelMetrics 模型评估指标
type ModelMetrics struct {
	IC         float64 `json:"ic"`          // Information Coefficient
	RankIC     float64 `json:"rank_ic"`     // Rank IC
	ICIR       float64 `json:"icir"`        // IC Information Ratio
	RankICIR   float64 `json:"rank_icir"`   // Rank IC Information Ratio
	MSE        float64 `json:"mse"`         // Mean Squared Error
	MAE        float64 `json:"mae"`         // Mean Absolute Error
}

// BacktestResults 回测结果
type BacktestResults struct {
	TotalReturn     float64            `json:"total_return"`
	AnnualReturn    float64            `json:"annual_return"`
	BenchmarkReturn float64            `json:"benchmark_return"`
	ExcessReturn    float64            `json:"excess_return"`
	SharpeRatio     float64            `json:"sharpe_ratio"`
	MaxDrawdown     float64            `json:"max_drawdown"`
	Volatility      float64            `json:"volatility"`
	WinRate         float64            `json:"win_rate"`
	Turnover        float64            `json:"turnover"`
	PerformanceChart []PerformancePoint `json:"performance_chart"`
	Holdings        []HoldingRecord    `json:"holdings"`
}

// PerformancePoint 业绩图表数据点
type PerformancePoint struct {
	Date      string  `json:"date"`
	Return    float64 `json:"return"`
	Benchmark float64 `json:"benchmark"`
	Drawdown  float64 `json:"drawdown"`
}

// HoldingRecord 持仓记录
type HoldingRecord struct {
	Date   string             `json:"date"`
	Stocks map[string]float64 `json:"stocks"` // 股票代码 -> 权重
}

// NewWorkflowRunner 创建新的工作流执行器
func NewWorkflowRunner(client *QlibClient) *WorkflowRunner {
	scriptDir := os.Getenv("QLIB_SCRIPT_DIR")
	if scriptDir == "" {
		scriptDir = "./scripts/qlib"
	}
	
	outputDir := os.Getenv("QLIB_OUTPUT_DIR") 
	if outputDir == "" {
		outputDir = "./output/qlib"
	}

	return &WorkflowRunner{
		client:    client,
		scriptDir: scriptDir,
		outputDir: outputDir,
	}
}

// RunWorkflow 执行完整的Qlib工作流
func (wr *WorkflowRunner) RunWorkflow(ctx context.Context, config WorkflowConfig, taskID string) (*WFResult, error) {
	log.Printf("开始执行工作流任务: %s", taskID)
	
	// 创建输出目录
	taskOutputDir := filepath.Join(wr.outputDir, taskID)
	if err := os.MkdirAll(taskOutputDir, 0755); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %w", err)
	}

	result := &WFResult{
		TaskID:    taskID,
		Status:    "running",
		Progress:  0,
		StartTime: time.Now(),
		Results:   make(map[string]interface{}),
	}

	// 生成工作流脚本
	script, err := wr.generateWorkflowScript(config, taskOutputDir)
	if err != nil {
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("生成脚本失败: %v", err)
		return result, err
	}

	// 执行工作流脚本
	output, err := wr.client.ExecuteScript(ctx, script)
	if err != nil {
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("执行工作流失败: %v", err)
		return result, err
	}

	// 解析执行结果
	if err := wr.parseWorkflowResults(output, taskOutputDir, result); err != nil {
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("解析结果失败: %v", err)
		return result, err
	}

	endTime := time.Now()
	result.EndTime = &endTime
	result.Status = "completed"
	result.Progress = 100

	log.Printf("工作流任务完成: %s", taskID)
	return result, nil
}

// generateWorkflowScript 生成工作流执行脚本
func (wr *WorkflowRunner) generateWorkflowScript(config WorkflowConfig, outputDir string) (string, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("序列化配置失败: %w", err)
	}

	script := fmt.Sprintf(`
import json
import sys
import os
import pandas as pd
import numpy as np
from datetime import datetime
import qlib
from qlib.workflow import R
from qlib.workflow.record_temp import SignalRecord, PortAnaRecord
from qlib.data.dataset import DatasetH
from qlib.data.dataset.handler import DataHandlerLP
from qlib.model.trainer import task_train
from qlib.backtest import executor as qexecutor
from qlib.contrib.strategy.signal_strategy import TopkDropoutStrategy
from qlib.contrib.evaluate import backtest_analyze

# 设置输出目录
output_dir = '%s'
os.makedirs(output_dir, exist_ok=True)

# 加载配置
config = json.loads('''%s''')

try:
    # 步骤1: 数据准备
    print("步骤1: 准备数据集...")
    
    # 构建数据处理器配置
    data_handler_config = {
        "start_time": config["start_time"],
        "end_time": config["end_time"],
        "fit_start_time": config["train_period"][0],
        "fit_end_time": config["train_period"][1],
        "instruments": config["market"],
        "infer_processors": [
            {"class": "RobustZScoreNorm", "kwargs": {"fields_group": "feature", "clip_outlier": True}},
            {"class": "Fillna", "kwargs": {"fields_group": "feature"}}
        ],
        "learn_processors": [
            {"class": "DropnaLabel"},
            {"class": "CSRankNorm", "kwargs": {"fields_group": "label"}}
        ]
    }
    
    # 创建数据集
    dataset = DatasetH(
        handler={
            "class": "Alpha158",
            "module_path": "qlib.contrib.data.handler",
            "kwargs": data_handler_config
        }
    )
    
    print("数据集准备完成")
    
    # 步骤2: 模型训练
    print("步骤2: 训练模型...")
    
    # 构建模型配置
    model_config = {
        "class": config["model"]["class"],
        "module_path": config["model"]["module"],
        "kwargs": config["model"].get("kwargs", {})
    }
    
    # 训练模型
    with R.start(experiment_name="qlib_workflow"):
        model = task_train(
            dataset, 
            model=model_config
        )
        
        # 保存模型
        model_path = os.path.join(output_dir, "model.pkl")
        with open(model_path, 'wb') as f:
            import pickle
            pickle.dump(model, f)
        
        # 获取预测结果
        pred_score = model.predict(dataset)
        
        # 计算模型指标
        from qlib.contrib.evaluate import risk_analysis
        from qlib.contrib.strategy.signal_strategy import TopkDropoutStrategy
        
        # 信号分析
        signal_record = SignalRecord(model, dataset, recorder=R.get_recorder())
        signal_record.generate()
        
        # 获取IC等指标
        pred_label = dataset.prepare(["label"], col_set="label")
        ic_metrics = {}
        
        if len(pred_score) > 0 and len(pred_label) > 0:
            # 计算IC指标
            ic_data = []
            for date in pred_score.index.get_level_values(0).unique():
                if date in pred_label.index.get_level_values(0):
                    pred_day = pred_score.loc[date]
                    label_day = pred_label.loc[date]
                    merged = pd.concat([pred_day, label_day], axis=1, join='inner')
                    if len(merged) > 1:
                        ic = merged.iloc[:, 0].corr(merged.iloc[:, 1])
                        rank_ic = merged.iloc[:, 0].corr(merged.iloc[:, 1], method='spearman')
                        ic_data.append({'date': date, 'ic': ic, 'rank_ic': rank_ic})
            
            ic_df = pd.DataFrame(ic_data)
            if len(ic_df) > 0:
                ic_metrics = {
                    'ic': ic_df['ic'].mean(),
                    'rank_ic': ic_df['rank_ic'].mean(),
                    'icir': ic_df['ic'].mean() / ic_df['ic'].std() if ic_df['ic'].std() > 0 else 0,
                    'rank_icir': ic_df['rank_ic'].mean() / ic_df['rank_ic'].std() if ic_df['rank_ic'].std() > 0 else 0
                }
        
        print("模型训练完成")
        
        # 步骤3: 策略回测
        print("步骤3: 执行策略回测...")
        
        # 构建策略配置
        strategy_config = config.get("strategy", {
            "class": "TopkDropoutStrategy",
            "module_path": "qlib.contrib.strategy.signal_strategy", 
            "kwargs": {
                "signal": pred_score,
                "topk": 50,
                "n_drop": 5
            }
        })
        
        # 构建回测配置
        backtest_config = config.get("backtest", {
            "start_time": config["test_period"][0],
            "end_time": config["test_period"][1],
            "account": 100000000,
            "benchmark": "SH000300",
            "exchange_kwargs": {
                "freq": "day",
                "limit_threshold": 0.095,
                "deal_price": "close",
                "open_cost": 0.0005,
                "close_cost": 0.0015,
                "trade_unit": 100
            }
        })
        
        # 创建策略
        strategy = TopkDropoutStrategy(**strategy_config.get("kwargs", {}))
        
        # 执行回测
        executor_config = {
            "class": "SimulatorExecutor",
            "module_path": "qlib.backtest.executor",
            "kwargs": {
                "time_per_step": "day",
                "generate_portfolio_metrics": True
            }
        }
        
        portfolio_metric_dict, indicator_dict = qexecutor.backtest(
            start_time=backtest_config["start_time"],
            end_time=backtest_config["end_time"],
            strategy=strategy,
            executor=executor_config,
            benchmark=backtest_config["benchmark"],
            account=backtest_config["account"],
            exchange_kwargs=backtest_config["exchange_kwargs"]
        )
        
        # 计算回测指标
        backtest_results = {}
        if 'excess_return_wo_cost' in portfolio_metric_dict:
            returns = portfolio_metric_dict['excess_return_wo_cost'].dropna()
            if len(returns) > 0:
                total_return = (1 + returns).prod() - 1
                annual_return = (1 + returns).mean() * 252 - 1
                volatility = returns.std() * np.sqrt(252)
                sharpe_ratio = annual_return / volatility if volatility > 0 else 0
                max_drawdown = (returns.cumsum() - returns.cumsum().cummax()).min()
                
                backtest_results = {
                    'total_return': float(total_return),
                    'annual_return': float(annual_return),
                    'sharpe_ratio': float(sharpe_ratio),
                    'max_drawdown': float(max_drawdown),
                    'volatility': float(volatility),
                    'win_rate': float((returns > 0).mean())
                }
        
        print("策略回测完成")
        
        # 保存结果
        results = {
            'task_id': os.path.basename(output_dir),
            'status': 'completed',
            'progress': 100,
            'model_metrics': ic_metrics,
            'backtest_results': backtest_results,
            'output_dir': output_dir
        }
        
        # 保存结果到文件
        with open(os.path.join(output_dir, 'results.json'), 'w') as f:
            json.dump(results, f, indent=2, default=str)
        
        print("工作流执行完成")
        print(json.dumps(results, default=str))

except Exception as e:
    error_result = {
        'task_id': os.path.basename(output_dir),
        'status': 'failed',
        'progress': 0,
        'error_message': str(e)
    }
    print(json.dumps(error_result))
    sys.exit(1)
`, outputDir, string(configJSON))

	return script, nil
}

// parseWorkflowResults 解析工作流执行结果
func (wr *WorkflowRunner) parseWorkflowResults(output []byte, outputDir string, result *WFResult) error {
	// 尝试解析JSON输出
	var scriptResult map[string]interface{}
	if err := json.Unmarshal(output, &scriptResult); err != nil {
		// 如果无法解析JSON，尝试从文件读取结果
		resultFile := filepath.Join(outputDir, "results.json")
		if _, err := os.Stat(resultFile); err == nil {
			data, err := os.ReadFile(resultFile)
			if err != nil {
				return fmt.Errorf("读取结果文件失败: %w", err)
			}
			if err := json.Unmarshal(data, &scriptResult); err != nil {
				return fmt.Errorf("解析结果文件失败: %w", err)
			}
		} else {
			return fmt.Errorf("无法解析脚本输出: %w", err)
		}
	}

	// 解析模型指标
	if metrics, ok := scriptResult["model_metrics"].(map[string]interface{}); ok {
		result.ModelMetrics = ModelMetrics{
			IC:       getFloat64(metrics, "ic"),
			RankIC:   getFloat64(metrics, "rank_ic"),
			ICIR:     getFloat64(metrics, "icir"),
			RankICIR: getFloat64(metrics, "rank_icir"),
		}
	}

	// 解析回测结果
	if backtest, ok := scriptResult["backtest_results"].(map[string]interface{}); ok {
		result.BacktestResults = BacktestResults{
			TotalReturn:     getFloat64(backtest, "total_return"),
			AnnualReturn:    getFloat64(backtest, "annual_return"),
			SharpeRatio:     getFloat64(backtest, "sharpe_ratio"),
			MaxDrawdown:     getFloat64(backtest, "max_drawdown"),
			Volatility:      getFloat64(backtest, "volatility"),
			WinRate:         getFloat64(backtest, "win_rate"),
		}
	}

	// 保存所有结果
	result.Results = scriptResult

	return nil
}

// getFloat64 安全地从map中获取float64值
func getFloat64(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0.0
}

// GetWorkflowTemplates 获取工作流模板
func (wr *WorkflowRunner) GetWorkflowTemplates() []WFTemplate {
	return []WFTemplate{
		{
			Name:        "LightGBM Alpha158 完整流程",
			Description: "使用LightGBM模型和Alpha158因子的完整量化流程",
			Config: WorkflowConfig{
				Market:    "csi300",
				StartTime: "2018-01-01",
				EndTime:   "2023-12-31",
				Features:  []string{"Alpha158"},
				Label:     "Ref($close, -2) / Ref($close, -1) - 1",
				Model: WFModelConfig{
					Class:  "LGBModel",
					Module: "qlib.contrib.model.gbdt",
					Kwargs: map[string]interface{}{
						"loss":         "mse",
						"num_leaves":   31,
						"learning_rate": 0.05,
						"feature_fraction": 0.9,
						"bagging_fraction": 0.8,
						"bagging_freq":     5,
						"verbose":          -1,
					},
				},
				TrainPeriod: []string{"2018-01-01", "2020-12-31"},
				ValidPeriod: []string{"2021-01-01", "2021-12-31"},
				TestPeriod:  []string{"2022-01-01", "2023-12-31"},
				Strategy: WFStrategyConfig{
					Class:  "TopkDropoutStrategy",
					Module: "qlib.contrib.strategy.signal_strategy",
					Kwargs: map[string]interface{}{
						"topk":   50,
						"n_drop": 5,
					},
				},
				BacktestConfig: WFBacktestConfig{
					StartTime: "2022-01-01", 
					EndTime:   "2023-12-31",
					Account:   100000000,
					Benchmark: "SH000300",
					Exchange: ExchangeConfig{
						Freq:            "day",
						Limit_threshold: 0.095,
						Deal_price:      "close",
						Open_cost:       0.0005,
						Close_cost:      0.0015,
						Trade_unit:      100,
					},
				},
			},
		},
		{
			Name:        "XGBoost Alpha360 增强流程",
			Description: "使用XGBoost模型和Alpha360因子的增强量化流程",
			Config: WorkflowConfig{
				Market:    "csi500",
				StartTime: "2019-01-01",
				EndTime:   "2023-12-31",
				Features:  []string{"Alpha360"},
				Label:     "Ref($close, -2) / Ref($close, -1) - 1",
				Model: WFModelConfig{
					Class:  "XGBModel",
					Module: "qlib.contrib.model.xgboost",
					Kwargs: map[string]interface{}{
						"n_estimators":   100,
						"max_depth":      6,
						"learning_rate":  0.1,
						"subsample":      0.8,
						"colsample_bytree": 0.8,
					},
				},
				TrainPeriod: []string{"2019-01-01", "2021-12-31"},
				ValidPeriod: []string{"2022-01-01", "2022-06-30"},
				TestPeriod:  []string{"2022-07-01", "2023-12-31"},
				Strategy: WFStrategyConfig{
					Class:  "TopkDropoutStrategy",
					Module: "qlib.contrib.strategy.signal_strategy",
					Kwargs: map[string]interface{}{
						"topk":   30,
						"n_drop": 3,
					},
				},
				BacktestConfig: WFBacktestConfig{
					StartTime: "2022-07-01",
					EndTime:   "2023-12-31",
					Account:   50000000,
					Benchmark: "SH000905",
					Exchange: ExchangeConfig{
						Freq:            "day",
						Limit_threshold: 0.095,
						Deal_price:      "close",
						Open_cost:       0.0005,
						Close_cost:      0.0015,
						Trade_unit:      100,
					},
				},
			},
		},
	}
}

// WFTemplate 工作流模板
type WFTemplate struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Config      WorkflowConfig `json:"config"`
}

// ValidateWorkflowConfig 验证工作流配置
func (wr *WorkflowRunner) ValidateWorkflowConfig(config WorkflowConfig) error {
	// 验证时间范围
	if config.StartTime == "" || config.EndTime == "" {
		return fmt.Errorf("开始时间和结束时间不能为空")
	}

	// 验证市场
	if config.Market == "" {
		return fmt.Errorf("市场参数不能为空")
	}

	// 验证训练、验证、测试时间段
	if len(config.TrainPeriod) != 2 {
		return fmt.Errorf("训练时间段必须包含开始和结束时间")
	}
	if len(config.ValidPeriod) != 2 {
		return fmt.Errorf("验证时间段必须包含开始和结束时间")
	}
	if len(config.TestPeriod) != 2 {
		return fmt.Errorf("测试时间段必须包含开始和结束时间")
	}

	// 验证模型配置
	if config.Model.Class == "" {
		return fmt.Errorf("模型类名不能为空")
	}
	if config.Model.Module == "" {
		return fmt.Errorf("模型模块路径不能为空")
	}

	return nil
}

// GenerateWorkflowYAML 生成Qlib YAML配置文件
func (wr *WorkflowRunner) GenerateWorkflowYAML(config WorkflowConfig) (string, error) {
	yamlConfig := map[string]interface{}{
		"qlib_init": map[string]interface{}{
			"provider_uri": "~/.qlib/qlib_data/cn_data",
			"region":       "cn",
		},
		"market":    config.Market,
		"benchmark": config.BacktestConfig.Benchmark,
		"data_handler_config": map[string]interface{}{
			"start_time":     config.StartTime,
			"end_time":       config.EndTime,
			"fit_start_time": config.TrainPeriod[0],
			"fit_end_time":   config.TrainPeriod[1],
			"instruments":    config.Market,
		},
		"port_analysis_config": map[string]interface{}{
			"strategy": map[string]interface{}{
				"class":       config.Strategy.Class,
				"module_path": config.Strategy.Module,
				"kwargs":      config.Strategy.Kwargs,
			},
			"backtest": map[string]interface{}{
				"start_time": config.BacktestConfig.StartTime,
				"end_time":   config.BacktestConfig.EndTime,
				"account":    config.BacktestConfig.Account,
				"benchmark":  config.BacktestConfig.Benchmark,
			},
		},
		"task": map[string]interface{}{
			"model": map[string]interface{}{
				"class":       config.Model.Class,
				"module_path": config.Model.Module,
				"kwargs":      config.Model.Kwargs,
			},
			"dataset": map[string]interface{}{
				"class":       "DatasetH",
				"module_path": "qlib.data.dataset",
				"kwargs": map[string]interface{}{
					"handler": map[string]interface{}{
						"class":       "Alpha158",
						"module_path": "qlib.contrib.data.handler",
					},
					"segments": map[string]interface{}{
						"train": config.TrainPeriod,
						"valid": config.ValidPeriod,
						"test":  config.TestPeriod,
					},
				},
			},
		},
	}

	yamlBytes, err := json.Marshal(yamlConfig)
	if err != nil {
		return "", fmt.Errorf("生成YAML配置失败: %w", err)
	}

	return string(yamlBytes), nil
}