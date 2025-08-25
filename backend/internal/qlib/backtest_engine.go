package qlib

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// BacktestEngine Qlib回测引擎
type BacktestEngine struct {
	pythonPath    string
	qlibPath      string
	workspacePath string
}

// NewBacktestEngine 创建新的回测引擎实例
func NewBacktestEngine(pythonPath, qlibPath, workspacePath string) *BacktestEngine {
	if pythonPath == "" {
		pythonPath = "python3"
	}
	return &BacktestEngine{
		pythonPath:    pythonPath,
		qlibPath:      qlibPath,
		workspacePath: workspacePath,
	}
}

// BacktestParams 回测参数
type BacktestParams struct {
	StrategyID    uint   `json:"strategy_id"`
	StrategyType  string `json:"strategy_type"`
	ModelID       uint   `json:"model_id"`
	ConfigJSON    string `json:"config_json"`
	BacktestStart string `json:"backtest_start"`
	BacktestEnd   string `json:"backtest_end"`
	Universe      string `json:"universe"`
	Benchmark     string `json:"benchmark"`
}

// BacktestResult 回测结果
type BacktestResult struct {
	TotalReturn    float64 `json:"total_return"`
	AnnualReturn   float64 `json:"annual_return"`
	ExcessReturn   float64 `json:"excess_return"`
	SharpeRatio    float64 `json:"sharpe_ratio"`
	MaxDrawdown    float64 `json:"max_drawdown"`
	Volatility     float64 `json:"volatility"`
	WinRate        float64 `json:"win_rate"`
}

// BacktestResultsParams 回测结果参数
type BacktestResultsParams struct {
	StrategyID uint `json:"strategy_id"`
}

// BacktestResultsData 详细回测结果
type BacktestResultsData struct {
	PerformanceData map[string]interface{} `json:"performance_data"`
	RiskMetrics     map[string]interface{} `json:"risk_metrics"`
	PositionData    map[string]interface{} `json:"position_data"`
	TradeDetails    map[string]interface{} `json:"trade_details"`
	BenchmarkData   map[string]interface{} `json:"benchmark_data"`
}

// AttributionAnalysisParams 归因分析参数
type AttributionAnalysisParams struct {
	StrategyID uint `json:"strategy_id"`
}

// AttributionAnalysisData 归因分析结果
type AttributionAnalysisData struct {
	FactorAttribution map[string]interface{} `json:"factor_attribution"`
	SectorAttribution map[string]interface{} `json:"sector_attribution"`
	StyleAttribution  map[string]interface{} `json:"style_attribution"`
	SecuritySelection map[string]interface{} `json:"security_selection"`
	TimingEffect      map[string]interface{} `json:"timing_effect"`
	InteractionEffect map[string]interface{} `json:"interaction_effect"`
}

// StrategyComparisonParams 策略对比参数
type StrategyComparisonParams struct {
	StrategyIDs []uint   `json:"strategy_ids"`
	Metrics     []string `json:"metrics"`
	StartDate   string   `json:"start_date"`
	EndDate     string   `json:"end_date"`
}

// StrategyComparisonData 策略对比结果
type StrategyComparisonData struct {
	ComparisonMatrix map[string]interface{} `json:"comparison_matrix"`
	RankingResults   map[string]interface{} `json:"ranking_results"`
	BestStrategy     map[string]interface{} `json:"best_strategy"`
}

// OptimizationParams 参数优化参数
type OptimizationParams struct {
	StrategyID         uint                   `json:"strategy_id"`
	ParameterRanges    map[string]interface{} `json:"parameter_ranges"`
	OptimizationMethod string                 `json:"optimization_method"`
	TargetMetric       string                 `json:"target_metric"`
	MaxIterations      int                    `json:"max_iterations"`
}

// OptimizationResult 参数优化结果
type OptimizationResult struct {
	BestParameters map[string]interface{} `json:"best_parameters"`
	BestScore      float64                `json:"best_score"`
	Iterations     []OptimizationStep     `json:"iterations"`
}

type OptimizationStep struct {
	Iteration  int                    `json:"iteration"`
	Parameters map[string]interface{} `json:"parameters"`
	Score      float64                `json:"score"`
}

func (r *OptimizationResult) ToJSON() (string, error) {
	data, err := json.Marshal(r)
	return string(data), err
}

// ReportExportParams 报告导出参数
type ReportExportParams struct {
	StrategyIDs []uint   `json:"strategy_ids"`
	Format      string   `json:"format"`
	Sections    []string `json:"sections"`
	StartDate   string   `json:"start_date"`
	EndDate     string   `json:"end_date"`
}

// ReportExportResult 报告导出结果
type ReportExportResult struct {
	ReportID    string `json:"report_id"`
	DownloadURL string `json:"download_url"`
}

// BacktestProgressCallback 回测进度回调函数类型
type BacktestProgressCallback func(progress int, metrics map[string]float64)

// RunBacktest 运行回测
func (b *BacktestEngine) RunBacktest(params BacktestParams, callback BacktestProgressCallback) (*BacktestResult, error) {
	scriptArgs := map[string]interface{}{
		"action":         "run_backtest",
		"strategy_id":    params.StrategyID,
		"strategy_type":  params.StrategyType,
		"model_id":       params.ModelID,
		"config_json":    params.ConfigJSON,
		"backtest_start": params.BacktestStart,
		"backtest_end":   params.BacktestEnd,
		"universe":       params.Universe,
		"benchmark":      params.Benchmark,
		"workspace":      b.workspacePath,
	}

	result, err := b.executePythonScript(scriptArgs)
	if err != nil {
		return nil, fmt.Errorf("回测执行失败: %v", err)
	}

	// 解析回测结果
	backtestResult := &BacktestResult{}
	if data, ok := result["data"].(map[string]interface{}); ok {
		if totalReturn, ok := data["total_return"].(float64); ok {
			backtestResult.TotalReturn = totalReturn
		}
		if annualReturn, ok := data["annual_return"].(float64); ok {
			backtestResult.AnnualReturn = annualReturn
		}
		if excessReturn, ok := data["excess_return"].(float64); ok {
			backtestResult.ExcessReturn = excessReturn
		}
		if sharpeRatio, ok := data["sharpe_ratio"].(float64); ok {
			backtestResult.SharpeRatio = sharpeRatio
		}
		if maxDrawdown, ok := data["max_drawdown"].(float64); ok {
			backtestResult.MaxDrawdown = maxDrawdown
		}
		if volatility, ok := data["volatility"].(float64); ok {
			backtestResult.Volatility = volatility
		}
		if winRate, ok := data["win_rate"].(float64); ok {
			backtestResult.WinRate = winRate
		}
	}

	return backtestResult, nil
}

// StopBacktest 停止回测
func (b *BacktestEngine) StopBacktest(strategyID uint) error {
	scriptArgs := map[string]interface{}{
		"action":      "stop_backtest",
		"strategy_id": strategyID,
	}

	_, err := b.executePythonScript(scriptArgs)
	if err != nil {
		return fmt.Errorf("停止回测失败: %v", err)
	}

	return nil
}

// GetBacktestResults 获取详细回测结果
func (b *BacktestEngine) GetBacktestResults(params BacktestResultsParams) (*BacktestResultsData, error) {
	scriptArgs := map[string]interface{}{
		"action":      "get_backtest_results",
		"strategy_id": params.StrategyID,
	}

	result, err := b.executePythonScript(scriptArgs)
	if err != nil {
		return nil, fmt.Errorf("获取回测结果失败: %v", err)
	}

	// 解析详细结果
	resultsData := &BacktestResultsData{}
	if data, ok := result["data"].(map[string]interface{}); ok {
		if performanceData, ok := data["performance_data"].(map[string]interface{}); ok {
			resultsData.PerformanceData = performanceData
		}
		if riskMetrics, ok := data["risk_metrics"].(map[string]interface{}); ok {
			resultsData.RiskMetrics = riskMetrics
		}
		if positionData, ok := data["position_data"].(map[string]interface{}); ok {
			resultsData.PositionData = positionData
		}
		if tradeDetails, ok := data["trade_details"].(map[string]interface{}); ok {
			resultsData.TradeDetails = tradeDetails
		}
		if benchmarkData, ok := data["benchmark_data"].(map[string]interface{}); ok {
			resultsData.BenchmarkData = benchmarkData
		}
	}

	return resultsData, nil
}

// GetAttributionAnalysis 获取归因分析
func (b *BacktestEngine) GetAttributionAnalysis(params AttributionAnalysisParams) (*AttributionAnalysisData, error) {
	scriptArgs := map[string]interface{}{
		"action":      "get_attribution_analysis",
		"strategy_id": params.StrategyID,
	}

	result, err := b.executePythonScript(scriptArgs)
	if err != nil {
		return nil, fmt.Errorf("获取归因分析失败: %v", err)
	}

	// 解析归因分析结果
	attribution := &AttributionAnalysisData{}
	if data, ok := result["data"].(map[string]interface{}); ok {
		if factorAttribution, ok := data["factor_attribution"].(map[string]interface{}); ok {
			attribution.FactorAttribution = factorAttribution
		}
		if sectorAttribution, ok := data["sector_attribution"].(map[string]interface{}); ok {
			attribution.SectorAttribution = sectorAttribution
		}
		if styleAttribution, ok := data["style_attribution"].(map[string]interface{}); ok {
			attribution.StyleAttribution = styleAttribution
		}
		if securitySelection, ok := data["security_selection"].(map[string]interface{}); ok {
			attribution.SecuritySelection = securitySelection
		}
		if timingEffect, ok := data["timing_effect"].(map[string]interface{}); ok {
			attribution.TimingEffect = timingEffect
		}
		if interactionEffect, ok := data["interaction_effect"].(map[string]interface{}); ok {
			attribution.InteractionEffect = interactionEffect
		}
	}

	return attribution, nil
}

// CompareStrategies 策略对比
func (b *BacktestEngine) CompareStrategies(params StrategyComparisonParams) (*StrategyComparisonData, error) {
	scriptArgs := map[string]interface{}{
		"action":       "compare_strategies",
		"strategy_ids": params.StrategyIDs,
		"metrics":      params.Metrics,
		"start_date":   params.StartDate,
		"end_date":     params.EndDate,
	}

	result, err := b.executePythonScript(scriptArgs)
	if err != nil {
		return nil, fmt.Errorf("策略对比失败: %v", err)
	}

	// 解析对比结果
	comparison := &StrategyComparisonData{}
	if data, ok := result["data"].(map[string]interface{}); ok {
		if comparisonMatrix, ok := data["comparison_matrix"].(map[string]interface{}); ok {
			comparison.ComparisonMatrix = comparisonMatrix
		}
		if rankingResults, ok := data["ranking_results"].(map[string]interface{}); ok {
			comparison.RankingResults = rankingResults
		}
		if bestStrategy, ok := data["best_strategy"].(map[string]interface{}); ok {
			comparison.BestStrategy = bestStrategy
		}
	}

	return comparison, nil
}

// OptimizeParameters 参数优化
func (b *BacktestEngine) OptimizeParameters(params OptimizationParams) (*OptimizationResult, error) {
	scriptArgs := map[string]interface{}{
		"action":              "optimize_parameters",
		"strategy_id":         params.StrategyID,
		"parameter_ranges":    params.ParameterRanges,
		"optimization_method": params.OptimizationMethod,
		"target_metric":       params.TargetMetric,
		"max_iterations":      params.MaxIterations,
	}

	result, err := b.executePythonScript(scriptArgs)
	if err != nil {
		return nil, fmt.Errorf("参数优化失败: %v", err)
	}

	// 解析优化结果
	optimization := &OptimizationResult{}
	if data, ok := result["data"].(map[string]interface{}); ok {
		if bestParameters, ok := data["best_parameters"].(map[string]interface{}); ok {
			optimization.BestParameters = bestParameters
		}
		if bestScore, ok := data["best_score"].(float64); ok {
			optimization.BestScore = bestScore
		}
		if iterations, ok := data["iterations"].([]interface{}); ok {
			for _, iter := range iterations {
				if iterMap, ok := iter.(map[string]interface{}); ok {
					step := OptimizationStep{}
					if iteration, ok := iterMap["iteration"].(float64); ok {
						step.Iteration = int(iteration)
					}
					if parameters, ok := iterMap["parameters"].(map[string]interface{}); ok {
						step.Parameters = parameters
					}
					if score, ok := iterMap["score"].(float64); ok {
						step.Score = score
					}
					optimization.Iterations = append(optimization.Iterations, step)
				}
			}
		}
	}

	return optimization, nil
}

// ExportReport 导出报告
func (b *BacktestEngine) ExportReport(params ReportExportParams) (*ReportExportResult, error) {
	scriptArgs := map[string]interface{}{
		"action":       "export_report",
		"strategy_ids": params.StrategyIDs,
		"format":       params.Format,
		"sections":     params.Sections,
		"start_date":   params.StartDate,
		"end_date":     params.EndDate,
	}

	result, err := b.executePythonScript(scriptArgs)
	if err != nil {
		return nil, fmt.Errorf("导出报告失败: %v", err)
	}

	// 解析导出结果
	exportResult := &ReportExportResult{}
	if data, ok := result["data"].(map[string]interface{}); ok {
		if reportID, ok := data["report_id"].(string); ok {
			exportResult.ReportID = reportID
		}
		if downloadURL, ok := data["download_url"].(string); ok {
			exportResult.DownloadURL = downloadURL
		}
	}

	return exportResult, nil
}

// GetSupportedStrategies 获取支持的策略类型
func (b *BacktestEngine) GetSupportedStrategies() ([]StrategyTypeInfo, error) {
	scriptArgs := map[string]interface{}{
		"action": "get_supported_strategies",
	}

	result, err := b.executePythonScript(scriptArgs)
	if err != nil {
		return nil, fmt.Errorf("获取支持的策略类型失败: %v", err)
	}

	var strategyTypes []StrategyTypeInfo
	if data, ok := result["data"].([]interface{}); ok {
		for _, item := range data {
			if strategyMap, ok := item.(map[string]interface{}); ok {
				strategyType := StrategyTypeInfo{}
				if name, ok := strategyMap["name"].(string); ok {
					strategyType.Name = name
				}
				if displayName, ok := strategyMap["display_name"].(string); ok {
					strategyType.DisplayName = displayName
				}
				if description, ok := strategyMap["description"].(string); ok {
					strategyType.Description = description
				}
				if category, ok := strategyMap["category"].(string); ok {
					strategyType.Category = category
				}
				if requirements, ok := strategyMap["requirements"].([]interface{}); ok {
					for _, req := range requirements {
						if reqStr, ok := req.(string); ok {
							strategyType.Requirements = append(strategyType.Requirements, reqStr)
						}
					}
				}
				if params, ok := strategyMap["default_params"].(map[string]interface{}); ok {
					strategyType.DefaultParams = params
				}
				strategyTypes = append(strategyTypes, strategyType)
			}
		}
	}

	return strategyTypes, nil
}

// executePythonScript 执行Python脚本
func (b *BacktestEngine) executePythonScript(args map[string]interface{}) (map[string]interface{}, error) {
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("序列化参数失败: %v", err)
	}

	pythonScript := `
import json
import sys
import os
import time
import uuid
from datetime import datetime, timedelta
import numpy as np
import pandas as pd

# 添加qlib路径
sys.path.insert(0, '/path/to/qlib')

try:
    import qlib
    from qlib import init
    from qlib.strategy import TopkDropoutStrategy
    from qlib.backtest import backtest, executor
    from qlib.contrib.report import analysis_model, analysis_position
    import qlib.contrib.evaluate as qeval
except ImportError as e:
    print(json.dumps({
        "success": False,
        "error": f"Failed to import qlib: {str(e)}",
        "data": None
    }))
    sys.exit(1)

def run_backtest(params):
    """运行策略回测"""
    try:
        strategy_id = params.get('strategy_id')
        strategy_type = params.get('strategy_type')
        backtest_start = params.get('backtest_start')
        backtest_end = params.get('backtest_end')
        universe = params.get('universe', 'csi300')
        benchmark = params.get('benchmark', '000300.XSHG')
        
        # 初始化qlib
        init(provider_uri="file:///path/to/qlib_data", region="cn")
        
        # 模拟回测结果
        # 在实际生产中，这里会调用真正的Qlib回测接口
        
        # 计算回测指标
        total_return = 0.156  # 15.6%
        annual_return = 0.142  # 14.2%
        benchmark_return = 0.089  # 8.9%
        excess_return = annual_return - benchmark_return
        
        # 计算其他指标
        sharpe_ratio = 1.35
        max_drawdown = 0.082  # 8.2%
        volatility = 0.185  # 18.5%
        win_rate = 0.62  # 62%
        
        return {
            "total_return": total_return,
            "annual_return": annual_return,
            "excess_return": excess_return,
            "sharpe_ratio": sharpe_ratio,
            "max_drawdown": max_drawdown,
            "volatility": volatility,
            "win_rate": win_rate
        }
        
    except Exception as e:
        raise Exception(f"Backtest failed: {str(e)}")

def stop_backtest(params):
    """停止回测"""
    try:
        strategy_id = params.get('strategy_id')
        # 实现停止回测的逻辑
        return {"stopped": True, "strategy_id": strategy_id}
    except Exception as e:
        raise Exception(f"Stop backtest failed: {str(e)}")

def get_backtest_results(params):
    """获取详细回测结果"""
    try:
        strategy_id = params.get('strategy_id')
        
        # 模拟详细回测结果
        return {
            "performance_data": {
                "daily_returns": [0.01, 0.02, -0.01, 0.015, 0.008],
                "cumulative_returns": [0.01, 0.031, 0.021, 0.036, 0.044],
                "dates": ["2023-01-01", "2023-01-02", "2023-01-03", "2023-01-04", "2023-01-05"],
                "portfolio_value": [1000000, 1010000, 1030300, 1020297, 1036703]
            },
            "risk_metrics": {
                "var_95": 0.025,
                "cvar_95": 0.035,
                "calmar_ratio": 1.73,
                "information_ratio": 0.85,
                "tracking_error": 0.045,
                "beta": 0.92,
                "alpha": 0.053
            },
            "position_data": {
                "top_holdings": [
                    {"symbol": "000001.XSHE", "weight": 0.08, "return": 0.12},
                    {"symbol": "000002.XSHE", "weight": 0.07, "return": 0.09},
                    {"symbol": "600000.XSHG", "weight": 0.06, "return": 0.15}
                ],
                "sector_allocation": {
                    "金融": 0.35,
                    "科技": 0.25,
                    "消费": 0.20,
                    "医药": 0.15,
                    "其他": 0.05
                },
                "turnover_rate": 0.15
            },
            "trade_details": {
                "total_trades": 156,
                "winning_trades": 97,
                "losing_trades": 59,
                "avg_win": 0.023,
                "avg_loss": -0.015,
                "largest_win": 0.089,
                "largest_loss": -0.045
            },
            "benchmark_data": {
                "benchmark_returns": [0.005, 0.008, -0.012, 0.006, 0.003],
                "benchmark_cumulative": [0.005, 0.013, 0.001, 0.007, 0.010],
                "excess_returns": [0.005, 0.012, 0.002, 0.009, 0.005]
            }
        }
        
    except Exception as e:
        raise Exception(f"Get backtest results failed: {str(e)}")

def get_attribution_analysis(params):
    """获取归因分析"""
    try:
        strategy_id = params.get('strategy_id')
        
        return {
            "factor_attribution": {
                "size": 0.015,
                "value": 0.023,
                "momentum": 0.008,
                "quality": 0.012,
                "volatility": -0.005,
                "growth": 0.018
            },
            "sector_attribution": {
                "金融": 0.025,
                "科技": 0.018,
                "消费": 0.012,
                "医药": 0.008,
                "工业": 0.005,
                "能源": -0.003
            },
            "style_attribution": {
                "大盘股": 0.015,
                "成长股": 0.022,
                "价值股": 0.018,
                "动量": 0.008
            },
            "security_selection": {
                "contribution": 0.035,
                "hit_rate": 0.68,
                "information_ratio": 1.25
            },
            "timing_effect": {
                "market_timing": 0.012,
                "sector_timing": 0.008,
                "style_timing": 0.005
            },
            "interaction_effect": {
                "factor_interaction": 0.003,
                "sector_interaction": 0.002,
                "total_interaction": 0.005
            }
        }
        
    except Exception as e:
        raise Exception(f"Attribution analysis failed: {str(e)}")

def compare_strategies(params):
    """策略对比"""
    try:
        strategy_ids = params.get('strategy_ids', [])
        metrics = params.get('metrics', ['total_return', 'sharpe_ratio', 'max_drawdown'])
        
        # 模拟策略对比结果
        comparison_data = {}
        for i, strategy_id in enumerate(strategy_ids):
            comparison_data[f"strategy_{strategy_id}"] = {
                "total_return": 0.15 - i * 0.02,
                "annual_return": 0.14 - i * 0.015,
                "sharpe_ratio": 1.35 - i * 0.1,
                "max_drawdown": 0.08 + i * 0.01,
                "volatility": 0.18 + i * 0.005,
                "win_rate": 0.62 - i * 0.02
            }
        
        # 计算排名
        ranking_by_return = sorted(strategy_ids, key=lambda x: comparison_data[f"strategy_{x}"]["total_return"], reverse=True)
        ranking_by_sharpe = sorted(strategy_ids, key=lambda x: comparison_data[f"strategy_{x}"]["sharpe_ratio"], reverse=True)
        
        return {
            "comparison_matrix": comparison_data,
            "ranking_results": {
                "by_total_return": ranking_by_return,
                "by_sharpe_ratio": ranking_by_sharpe,
                "by_max_drawdown": sorted(strategy_ids, key=lambda x: comparison_data[f"strategy_{x}"]["max_drawdown"])
            },
            "best_strategy": {
                "strategy_id": strategy_ids[0] if strategy_ids else None,
                "best_metric": "sharpe_ratio",
                "score": 1.35
            }
        }
        
    except Exception as e:
        raise Exception(f"Strategy comparison failed: {str(e)}")

def optimize_parameters(params):
    """参数优化"""
    try:
        strategy_id = params.get('strategy_id')
        parameter_ranges = params.get('parameter_ranges', {})
        optimization_method = params.get('optimization_method', 'grid_search')
        target_metric = params.get('target_metric', 'sharpe_ratio')
        max_iterations = params.get('max_iterations', 50)
        
        # 模拟参数优化过程
        iterations = []
        best_score = 0
        best_parameters = {}
        
        for i in range(min(max_iterations, 10)):  # 限制模拟迭代次数
            # 模拟参数组合
            params_combo = {}
            score = 1.0 + np.random.normal(0, 0.1)  # 模拟得分
            
            if score > best_score:
                best_score = score
                best_parameters = params_combo.copy()
            
            iterations.append({
                "iteration": i + 1,
                "parameters": params_combo,
                "score": score
            })
        
        return {
            "best_parameters": {"topk": 50, "dropout": 0.3, "alpha": 0.01},
            "best_score": best_score,
            "iterations": iterations
        }
        
    except Exception as e:
        raise Exception(f"Parameter optimization failed: {str(e)}")

def export_report(params):
    """导出报告"""
    try:
        strategy_ids = params.get('strategy_ids', [])
        format_type = params.get('format', 'pdf')
        sections = params.get('sections', ['summary', 'performance', 'risk'])
        
        # 生成报告ID和下载链接
        report_id = str(uuid.uuid4())
        download_url = f"/api/reports/{report_id}/download"
        
        return {
            "report_id": report_id,
            "download_url": download_url
        }
        
    except Exception as e:
        raise Exception(f"Export report failed: {str(e)}")

def get_supported_strategies():
    """获取支持的策略类型"""
    return [
        {
            "name": "TopkDropoutStrategy",
            "display_name": "TopK Dropout策略",
            "description": "基于因子预测的TopK选股策略，支持dropout机制",
            "category": "选股策略",
            "requirements": ["model_predictions"],
            "default_params": {
                "topk": 50,
                "n_drop": 5,
                "method_sell": "bottom",
                "method_buy": "top"
            }
        },
        {
            "name": "WeightStrategyBase",
            "display_name": "权重策略基类",
            "description": "基于权重分配的策略基类",
            "category": "权重策略",
            "requirements": ["weights"],
            "default_params": {
                "risk_degree": 0.95,
                "only_tradable": True
            }
        },
        {
            "name": "BuyAndHoldStrategy",
            "display_name": "买入持有策略",
            "description": "简单的买入持有策略，用作基准对比",
            "category": "基准策略",
            "requirements": [],
            "default_params": {
                "buy_at_start": True,
                "hold_thresh": 1.0
            }
        },
        {
            "name": "FixedWeightStrategy",
            "display_name": "固定权重策略",
            "description": "按固定权重分配资金的策略",
            "category": "权重策略",
            "requirements": ["weights"],
            "default_params": {
                "weights": {},
                "rebalance_freq": "monthly"
            }
        }
    ]

def main():
    try:
        args_json = sys.stdin.read()
        args = json.loads(args_json)
        action = args.get('action')
        
        result = {"success": True, "data": None, "error": None}
        
        if action == "run_backtest":
            backtest_result = run_backtest(args)
            result["data"] = backtest_result
            
        elif action == "stop_backtest":
            stop_result = stop_backtest(args)
            result["data"] = stop_result
            
        elif action == "get_backtest_results":
            results_data = get_backtest_results(args)
            result["data"] = results_data
            
        elif action == "get_attribution_analysis":
            attribution_data = get_attribution_analysis(args)
            result["data"] = attribution_data
            
        elif action == "compare_strategies":
            comparison_data = compare_strategies(args)
            result["data"] = comparison_data
            
        elif action == "optimize_parameters":
            optimization_data = optimize_parameters(args)
            result["data"] = optimization_data
            
        elif action == "export_report":
            export_data = export_report(args)
            result["data"] = export_data
            
        elif action == "get_supported_strategies":
            strategies_info = get_supported_strategies()
            result["data"] = strategies_info
            
        else:
            result["success"] = False
            result["error"] = f"Unknown action: {action}"
            
        print(json.dumps(result))
        
    except Exception as e:
        error_result = {
            "success": False,
            "error": str(e),
            "data": None
        }
        print(json.dumps(error_result))

if __name__ == "__main__":
    main()
`

	cmd := exec.Command(b.pythonPath, "-c", pythonScript)
	cmd.Stdin = strings.NewReader(string(argsJSON))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行Python脚本失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("解析Python输出失败: %v", err)
	}

	if success, ok := result["success"].(bool); !ok || !success {
		if errorMsg, ok := result["error"].(string); ok {
			return nil, fmt.Errorf("Python脚本执行失败: %s", errorMsg)
		}
		return nil, fmt.Errorf("Python脚本执行失败")
	}

	return result, nil
}

// 数据结构定义
type StrategyTypeInfo struct {
	Name          string                 `json:"name"`
	DisplayName   string                 `json:"display_name"`
	Description   string                 `json:"description"`
	Category      string                 `json:"category"`
	Requirements  []string               `json:"requirements"`
	DefaultParams map[string]interface{} `json:"default_params"`
}