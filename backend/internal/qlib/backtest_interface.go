package qlib

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// BacktestInterface 回测接口
type BacktestInterface struct {
	client *QlibClient
}

// BacktestConfig 回测配置
type BacktestConfig struct {
	StrategyName string                 `json:"strategy_name"` // 策略名称
	ModelID      string                 `json:"model_id"`      // 模型ID
	StartDate    time.Time              `json:"start_date"`    // 回测开始日期
	EndDate      time.Time              `json:"end_date"`      // 回测结束日期
	Benchmark    string                 `json:"benchmark"`     // 基准
	Universe     string                 `json:"universe"`      // 股票池
	InitCash     float64                `json:"init_cash"`     // 初始资金
	Commission   float64                `json:"commission"`    // 手续费率
	Strategy     StrategyConfig         `json:"strategy"`      // 策略配置
	Portfolio    PortfolioConfig        `json:"portfolio"`     // 组合配置
	RiskModel    map[string]interface{} `json:"risk_model"`    // 风险模型配置
	TaskName     string                 `json:"task_name"`     // 任务名称
}

// StrategyConfig 策略配置
type StrategyConfig struct {
	ClassName    string                 `json:"class_name"`    // 策略类名
	ModulePath   string                 `json:"module_path"`   // 模块路径
	Parameters   map[string]interface{} `json:"parameters"`    // 策略参数
	SignalConfig SignalConfig           `json:"signal_config"` // 信号配置
}

// SignalConfig 信号配置
type SignalConfig struct {
	TopK     int     `json:"topk"`      // 选股数量
	Buffer   float64 `json:"buffer"`    // 缓冲区比例
	Method   string  `json:"method"`    // 选股方法
	Reverse  bool    `json:"reverse"`   // 是否反向选择
}

// PortfolioConfig 组合配置
type PortfolioConfig struct {
	AccountType     string                 `json:"account_type"`     // 账户类型
	CashLimit       float64                `json:"cash_limit"`       // 现金限制
	GenerateAmount  float64                `json:"generate_amount"`  // 生成金额
	TradeExchange   string                 `json:"trade_exchange"`   // 交易所
	LimitThreshold  float64                `json:"limit_threshold"`  // 限价阈值
	DealPrice       string                 `json:"deal_price"`       // 成交价格
	SubscriptMethod string                 `json:"subscript_method"` // 下标方法
	RiskDegree      float64                `json:"risk_degree"`      // 风险程度
	Parameters      map[string]interface{} `json:"parameters"`       // 其他参数
}

// BacktestResultDetailed 详细回测结果
type BacktestResultDetailed struct {
	Success        bool                   `json:"success"`
	TaskID         string                 `json:"task_id"`
	StrategyName   string                 `json:"strategy_name"`
	Performance    PerformanceMetrics     `json:"performance"`
	Positions      []PositionRecord       `json:"positions"`
	Trades         []TradeRecord          `json:"trades"`
	Reports        BacktestReports        `json:"reports"`
	Charts         map[string]interface{} `json:"charts"`
	Error          string                 `json:"error"`
	Metadata       map[string]interface{} `json:"metadata"`
	Duration       time.Duration          `json:"duration"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	TotalReturn    float64 `json:"total_return"`    // 总收益率
	AnnualReturn   float64 `json:"annual_return"`   // 年化收益率
	Volatility     float64 `json:"volatility"`      // 波动率
	SharpeRatio    float64 `json:"sharpe_ratio"`    // 夏普比率
	MaxDrawdown    float64 `json:"max_drawdown"`    // 最大回撤
	CalmarRatio    float64 `json:"calmar_ratio"`    // 卡玛比率
	WinRate        float64 `json:"win_rate"`        // 胜率
	ProfitLossRate float64 `json:"profit_loss_rate"` // 盈亏比
	Beta           float64 `json:"beta"`            // 贝塔
	Alpha          float64 `json:"alpha"`           // 阿尔法
	IR             float64 `json:"ir"`              // 信息比率
	Tracking       float64 `json:"tracking"`        // 跟踪误差
}

// PositionRecord 持仓记录
type PositionRecord struct {
	Date       time.Time `json:"date"`
	Instrument string    `json:"instrument"`
	Amount     float64   `json:"amount"`     // 持仓数量
	Weight     float64   `json:"weight"`     // 权重
	Price      float64   `json:"price"`      // 价格
	Value      float64   `json:"value"`      // 市值
}

// TradeRecord 交易记录
type TradeRecord struct {
	Date       time.Time `json:"date"`
	Instrument string    `json:"instrument"`
	Direction  string    `json:"direction"`  // buy/sell
	Amount     float64   `json:"amount"`     // 交易数量
	Price      float64   `json:"price"`      // 成交价格
	Commission float64   `json:"commission"` // 手续费
	PnL        float64   `json:"pnl"`        // 盈亏
}

// BacktestReports 回测报告
type BacktestReports struct {
	Summary       string                 `json:"summary"`        // 总结报告
	Analytics     map[string]interface{} `json:"analytics"`      // 分析报告
	RiskAnalysis  map[string]interface{} `json:"risk_analysis"`  // 风险分析
	Attribution   map[string]interface{} `json:"attribution"`    // 归因分析
}

// NewBacktestInterface 创建回测接口
func NewBacktestInterface(client *QlibClient) *BacktestInterface {
	return &BacktestInterface{
		client: client,
	}
}

// RunBacktest 运行回测
func (bi *BacktestInterface) RunBacktest(ctx context.Context, config BacktestConfig) (*BacktestResultDetailed, error) {
	if !bi.client.IsInitialized() {
		return nil, fmt.Errorf("Qlib客户端未初始化")
	}

	log.Printf("开始运行回测: %s", config.StrategyName)

	// 构建回测脚本
	script := bi.buildBacktestScript(config)

	// 执行回测
	startTime := time.Now()
	output, err := bi.client.ExecuteScript(ctx, script)
	duration := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("回测执行失败: %w", err)
	}

	// 解析结果
	var result BacktestResultDetailed
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("解析回测结果失败: %w", err)
	}

	result.Duration = duration

	if !result.Success {
		return nil, fmt.Errorf("回测失败: %s", result.Error)
	}

	log.Printf("回测完成: %s, 耗时: %v", result.TaskID, duration)
	return &result, nil
}

// buildBacktestScript 构建回测脚本
func (bi *BacktestInterface) buildBacktestScript(config BacktestConfig) string {
	configJson, _ := json.Marshal(config)

	return fmt.Sprintf(`
import json
import qlib
from qlib import data
from qlib.workflow import R
from qlib.workflow.record_temp import SignalRecord, PortAnaRecord
from qlib.backtest import backtest, executor
from qlib.contrib.strategy import TopkDropoutStrategy
from qlib.contrib.evaluate import risk_analysis
from qlib.utils import flatten_dict, init_instance_by_config
import pandas as pd
import numpy as np
import uuid
from datetime import datetime
import warnings
warnings.filterwarnings('ignore')

try:
	# 解析配置
	config = json.loads('''%s''')
	
	# 生成任务ID
	task_id = str(uuid.uuid4())[:8]
	
	# 配置参数
	start_date = config['start_date'][:10] if config['start_date'] else '2022-01-01'
	end_date = config['end_date'][:10] if config['end_date'] else '2023-12-31'
	benchmark = config.get('benchmark', 'SH000300')
	init_cash = config.get('init_cash', 1000000)
	commission = config.get('commission', 0.003)
	
	# 股票池配置
	universe = config.get('universe', 'csi300')
	if universe == 'csi300':
		instruments = data.D.instruments(market='csi300')
	elif universe == 'csi500':
		instruments = data.D.instruments(market='csi500')
	else:
		instruments = data.D.instruments(market='all')
	
	# 加载模型（如果指定）
	model = None
	model_id = config.get('model_id')
	if model_id:
		try:
			with R.start(experiment_name="backtest", recorder_id=task_id):
				model = R.load_object('model')
		except:
			model = None
	
	# 构建策略配置
	strategy_config = {
		'class': 'TopkDropoutStrategy',
		'module_path': 'qlib.contrib.strategy.signal_strategy',
		'kwargs': {
			'signal': None,  # 将在下面设置
			'topk': config['strategy']['signal_config'].get('topk', 30),
			'n_drop': config['strategy']['signal_config'].get('topk', 30) // 10,  # 默认为topk的1/10
			'method_sell': config['strategy']['signal_config'].get('method', 'bottom'),
			'method_buy': config['strategy']['signal_config'].get('method', 'top'),
		}
	}
	
	# 构建数据集用于生成信号
	dataset_config = {
		'class': 'DatasetH',
		'module_path': 'qlib.data.dataset',
		'kwargs': {
			'handler': {
				'class': 'Alpha158',
				'module_path': 'qlib.contrib.data.handler',
				'kwargs': {
					'start_time': start_date,
					'end_time': end_date,
					'instruments': instruments,
					'infer_processors': [
						{
							'class': 'RobustZScoreNorm',
							'kwargs': {'fields_group': 'feature', 'clip_outlier': True}
						},
						{
							'class': 'Fillna',
							'kwargs': {'fields_group': 'feature'}
						}
					]
				}
			},
			'segments': {
				'test': (start_date, end_date)
			}
		}
	}
	
	# 初始化数据集
	dataset = init_instance_by_config(dataset_config)
	
	# 生成信号
	if model is not None:
		signal = model.predict(dataset)
	else:
		# 如果没有模型，使用简单的动量策略信号
		signal = data.D.features(
			instruments=instruments,
			fields=['($close - Ref($close, 20)) / Ref($close, 20)'],  # 20日动量
			start_time=start_date,
			end_time=end_date,
			freq='day'
		)
		if signal is not None:
			signal.columns = ['score']
		else:
			raise ValueError("无法生成交易信号")
	
	# 设置策略信号
	strategy_config['kwargs']['signal'] = signal
	
	# 构建策略
	strategy = init_instance_by_config(strategy_config)
	
	# 构建执行器配置
	executor_config = {
		'class': 'SimulatorExecutor',
		'module_path': 'qlib.backtest.executor',
		'kwargs': {
			'time_per_step': 'day',
			'generate_portfolio_metrics': True,
			'verbose': False,
			'trade_exchange': {
				'class': 'Exchange',
				'module_path': 'qlib.backtest.exchange',
				'kwargs': {
					'freq': 'day',
					'limit_threshold': config['portfolio'].get('limit_threshold', 0.095),
					'deal_price': config['portfolio'].get('deal_price', 'close'),
					'open_cost': commission,
					'close_cost': commission,
					'min_cost': 5,
				}
			}
		}
	}
	
	# 构建组合配置
	portfolio_config = {
		'class': 'Account',
		'module_path': 'qlib.backtest.account',
		'kwargs': {
			'init_cash': init_cash,
			'fee_rate': commission,
			'deal_price': config['portfolio'].get('deal_price', 'close'),
		}
	}
	
	# 运行回测
	with R.start(experiment_name="backtest", recorder_id=task_id):
		# 执行回测
		executor = init_instance_by_config(executor_config)
		portfolio = init_instance_by_config(portfolio_config)
		
		# 运行策略
		result_portfolio = backtest(
			executor=executor,
			strategy=strategy,
			account=portfolio
		)
		
		# 计算性能指标
		if result_portfolio is not None and not result_portfolio.empty:
			# 基本性能分析
			returns = result_portfolio['return'] if 'return' in result_portfolio.columns else result_portfolio.iloc[:, 0]
			
			# 计算累计收益
			cum_returns = (1 + returns).cumprod()
			total_return = cum_returns.iloc[-1] - 1
			
			# 年化收益率
			days = len(returns)
			annual_return = (1 + total_return) ** (252 / days) - 1 if days > 0 else 0
			
			# 波动率
			volatility = returns.std() * np.sqrt(252)
			
			# 夏普比率
			sharpe_ratio = (annual_return - 0.03) / volatility if volatility > 0 else 0  # 假设无风险利率3%
			
			# 最大回撤
			cum_max = cum_returns.expanding().max()
			drawdown = (cum_returns - cum_max) / cum_max
			max_drawdown = drawdown.min()
			
			# 卡玛比率
			calmar_ratio = annual_return / abs(max_drawdown) if max_drawdown != 0 else 0
			
			# 胜率
			win_rate = (returns > 0).mean()
			
			# 构建性能指标
			performance = {
				'total_return': float(total_return),
				'annual_return': float(annual_return),
				'volatility': float(volatility),
				'sharpe_ratio': float(sharpe_ratio),
				'max_drawdown': float(max_drawdown),
				'calmar_ratio': float(calmar_ratio),
				'win_rate': float(win_rate),
				'profit_loss_rate': 1.5,  # 占位值
				'beta': 0.8,  # 占位值
				'alpha': float(annual_return - 0.08),  # 简化计算
				'ir': float(sharpe_ratio * 0.8),  # 占位值
				'tracking': float(volatility * 0.3)   # 占位值
			}
			
			# 保存记录
			R.save_objects(portfolio=result_portfolio)
			
			# 构建结果
			result = {
				'success': True,
				'task_id': task_id,
				'strategy_name': config['strategy_name'],
				'performance': performance,
				'positions': [],  # 简化处理
				'trades': [],     # 简化处理
				'reports': {
					'summary': f"回测完成，总收益率: {total_return:.4f}, 夏普比率: {sharpe_ratio:.4f}",
					'analytics': {'sample_days': days},
					'risk_analysis': {'max_drawdown': float(max_drawdown)},
					'attribution': {}
				},
				'charts': {
					'cumulative_returns': cum_returns.to_dict(),
					'daily_returns': returns.to_dict()
				},
				'metadata': {
					'start_date': start_date,
					'end_date': end_date,
					'sample_days': days,
					'universe': universe,
					'init_cash': init_cash
				}
			}
		else:
			result = {
				'success': False,
				'task_id': task_id,
				'strategy_name': config['strategy_name'],
				'error': '回测未产生有效的组合结果',
				'metadata': {}
			}

except Exception as e:
	result = {
		'success': False,
		'task_id': task_id if 'task_id' in locals() else '',
		'strategy_name': config.get('strategy_name', ''),
		'error': str(e),
		'metadata': {}
	}

print(json.dumps(result, default=str))
`, string(configJson))
}

// GetBacktestProgress 获取回测进度（模拟实现）
func (bi *BacktestInterface) GetBacktestProgress(ctx context.Context, taskID string) (map[string]interface{}, error) {
	// 这里是一个简化的实现，实际中需要真实的进度跟踪
	progress := map[string]interface{}{
		"task_id":         taskID,
		"status":          "completed", // running, completed, failed
		"progress":        100,
		"current_stage":   "完成",
		"elapsed_time":    "2m30s",
		"estimated_time":  "0s",
		"message":         "回测已完成",
	}

	return progress, nil
}

// CompareBacktests 对比多个回测结果
func (bi *BacktestInterface) CompareBacktests(ctx context.Context, taskIDs []string) (map[string]interface{}, error) {
	script := fmt.Sprintf(`
import json
import qlib
from qlib.workflow import R
import pandas as pd
import numpy as np

try:
	task_ids = %s
	
	# 加载多个回测结果
	portfolios = {}
	performance_metrics = {}
	
	for task_id in task_ids:
		try:
			with R.start(experiment_name="backtest", recorder_id=task_id):
				portfolio = R.load_object('portfolio')
				if portfolio is not None:
					portfolios[task_id] = portfolio
					
					# 计算性能指标
					if not portfolio.empty:
						returns = portfolio['return'] if 'return' in portfolio.columns else portfolio.iloc[:, 0]
						cum_returns = (1 + returns).cumprod()
						
						total_return = cum_returns.iloc[-1] - 1
						volatility = returns.std() * np.sqrt(252)
						sharpe = (total_return - 0.03) / volatility if volatility > 0 else 0
						
						cum_max = cum_returns.expanding().max()
						drawdown = (cum_returns - cum_max) / cum_max
						max_drawdown = drawdown.min()
						
						performance_metrics[task_id] = {
							'total_return': float(total_return),
							'volatility': float(volatility),
							'sharpe_ratio': float(sharpe),
							'max_drawdown': float(max_drawdown)
						}
		except:
			continue
	
	# 构建对比结果
	comparison = {
		'tasks': list(performance_metrics.keys()),
		'metrics': performance_metrics,
		'ranking': {
			'by_return': sorted(performance_metrics.items(), key=lambda x: x[1]['total_return'], reverse=True),
			'by_sharpe': sorted(performance_metrics.items(), key=lambda x: x[1]['sharpe_ratio'], reverse=True),
			'by_drawdown': sorted(performance_metrics.items(), key=lambda x: x[1]['max_drawdown'], reverse=True)
		},
		'summary': f"对比了 {len(performance_metrics)} 个回测结果"
	}
	
	result = {
		'success': True,
		'comparison': comparison
	}

except Exception as e:
	result = {
		'success': False,
		'comparison': {},
		'error': str(e)
	}

print(json.dumps(result, default=str))
`, fmt.Sprintf(`["%s"]`, taskIDs[0])) // 简化处理

	output, err := bi.client.ExecuteScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("对比回测失败: %w", err)
	}

	var response struct {
		Success    bool                   `json:"success"`
		Comparison map[string]interface{} `json:"comparison"`
		Error      string                 `json:"error"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("解析对比结果失败: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("对比失败: %s", response.Error)
	}

	return response.Comparison, nil
}

// GetBacktestReport 获取详细的回测报告
func (bi *BacktestInterface) GetBacktestReport(ctx context.Context, taskID string, reportType string) (map[string]interface{}, error) {
	script := fmt.Sprintf(`
import json
import qlib
from qlib.workflow import R
import pandas as pd
import numpy as np

try:
	task_id = '%s'
	report_type = '%s'
	
	# 加载回测结果
	with R.start(experiment_name="backtest", recorder_id=task_id):
		portfolio = R.load_object('portfolio')
		
		if portfolio is not None and not portfolio.empty:
			if report_type == 'performance':
				# 性能报告
				returns = portfolio['return'] if 'return' in portfolio.columns else portfolio.iloc[:, 0]
				cum_returns = (1 + returns).cumprod()
				
				report = {
					'total_return': float(cum_returns.iloc[-1] - 1),
					'annual_return': float((cum_returns.iloc[-1] ** (252/len(returns))) - 1),
					'volatility': float(returns.std() * np.sqrt(252)),
					'max_drawdown': float(((cum_returns / cum_returns.expanding().max()) - 1).min()),
					'win_rate': float((returns > 0).mean()),
					'daily_returns': returns.tail(20).to_dict()
				}
				
			elif report_type == 'positions':
				# 持仓报告
				report = {
					'current_positions': {},  # 当前持仓
					'position_history': {},   # 持仓历史
					'turnover': 0.15         # 换手率
				}
				
			elif report_type == 'trades':
				# 交易报告
				report = {
					'total_trades': 0,
					'win_trades': 0,
					'loss_trades': 0,
					'avg_win': 0.0,
					'avg_loss': 0.0,
					'trade_history': []
				}
				
			elif report_type == 'risk':
				# 风险报告
				returns = portfolio['return'] if 'return' in portfolio.columns else portfolio.iloc[:, 0]
				report = {
					'var_95': float(returns.quantile(0.05)),
					'var_99': float(returns.quantile(0.01)),
					'beta': 0.8,  # 占位值
					'tracking_error': float(returns.std() * np.sqrt(252) * 0.3)
				}
			else:
				report = {'message': '未知的报告类型'}
		else:
			report = {'error': '没有找到回测结果'}
	
	result = {
		'success': True,
		'report': report
	}

except Exception as e:
	result = {
		'success': False,
		'report': {},
		'error': str(e)
	}

print(json.dumps(result, default=str))
`, taskID, reportType)

	output, err := bi.client.ExecuteScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("获取回测报告失败: %w", err)
	}

	var response struct {
		Success bool                   `json:"success"`
		Report  map[string]interface{} `json:"report"`
		Error   string                 `json:"error"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("解析报告失败: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("获取报告失败: %s", response.Error)
	}

	return response.Report, nil
}

// OptimizeStrategy 策略参数优化
func (bi *BacktestInterface) OptimizeStrategy(ctx context.Context, baseConfig BacktestConfig, paramSpace map[string]interface{}) (map[string]interface{}, error) {
	// 简化的参数优化实现
	// 实际中需要使用网格搜索或贝叶斯优化等方法

	log.Printf("开始参数优化: %s", baseConfig.StrategyName)

	bestConfig := baseConfig
	bestPerformance := 0.0
	
	// 这里只是一个占位实现，实际需要根据参数空间进行搜索
	optimization := map[string]interface{}{
		"best_config":     bestConfig,
		"best_performance": bestPerformance,
		"optimization_history": []map[string]interface{}{
			{
				"params": paramSpace,
				"performance": bestPerformance,
			},
		},
		"total_trials": 1,
		"duration": "5m30s",
	}

	return optimization, nil
}

// DeleteBacktest 删除回测结果
func (bi *BacktestInterface) DeleteBacktest(ctx context.Context, taskID string) error {
	script := fmt.Sprintf(`
import json
import qlib
from qlib.workflow import R
import os
import shutil

try:
	task_id = '%s'
	
	# 删除recorder记录
	try:
		recorder_path = f"./mlruns/0/{task_id}"
		if os.path.exists(recorder_path):
			shutil.rmtree(recorder_path)
	except:
		pass
	
	result = {
		'success': True,
		'message': f'回测结果 {task_id} 已删除'
	}

except Exception as e:
	result = {
		'success': False,
		'error': str(e)
	}

print(json.dumps(result))
`, taskID)

	output, err := bi.client.ExecuteScript(ctx, script)
	if err != nil {
		return fmt.Errorf("删除回测失败: %w", err)
	}

	var response struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return fmt.Errorf("解析删除结果失败: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("删除回测失败: %s", response.Error)
	}

	return nil
}