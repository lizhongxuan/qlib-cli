package qlib

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

// WorkflowEngine Qlib工作流引擎
type WorkflowEngine struct {
	db          *gorm.DB
	pythonPath  string
	scriptDir   string
	workspaceDir string
}

// WorkflowTemplate 工作流模板
type WorkflowTemplate struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Config      map[string]interface{} `json:"config"`
	Steps       []WorkflowStep         `json:"steps"`
}

// WorkflowStep 工作流步骤
type WorkflowStep struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Config       map[string]interface{} `json:"config"`
	Dependencies []string               `json:"dependencies"`
	Required     bool                   `json:"required"`
}

// WorkflowProgressCallback 工作流进度回调函数
type WorkflowProgressCallback func(step string, progress int, message string)

// WorkflowResult 工作流执行结果
type WorkflowResult struct {
	Success      bool                   `json:"success"`
	Steps        []StepResult           `json:"steps"`
	Duration     time.Duration          `json:"duration"`
	OutputFiles  []string               `json:"output_files"`
	Metrics      map[string]interface{} `json:"metrics"`
	Error        string                 `json:"error,omitempty"`
}

// StepResult 步骤执行结果
type StepResult struct {
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Success   bool                   `json:"success"`
	Duration  time.Duration          `json:"duration"`
	Output    map[string]interface{} `json:"output"`
	Error     string                 `json:"error,omitempty"`
}

// NewWorkflowEngine 创建新的工作流引擎
func NewWorkflowEngine(db *gorm.DB) *WorkflowEngine {
	pythonPath := os.Getenv("PYTHON_PATH")
	if pythonPath == "" {
		pythonPath = "python3"
	}
	
	scriptDir := os.Getenv("QLIB_SCRIPT_DIR")
	if scriptDir == "" {
		scriptDir = "/tmp/qlib_scripts"
	}
	
	workspaceDir := os.Getenv("QLIB_WORKSPACE_DIR")
	if workspaceDir == "" {
		workspaceDir = "/tmp/qlib_workspace"
	}
	
	// 确保目录存在
	os.MkdirAll(scriptDir, 0755)
	os.MkdirAll(workspaceDir, 0755)
	
	return &WorkflowEngine{
		db:          db,
		pythonPath:  pythonPath,
		scriptDir:   scriptDir,
		workspaceDir: workspaceDir,
	}
}

// Execute 执行工作流
func (we *WorkflowEngine) Execute(ctx context.Context, template *WorkflowTemplate, config map[string]interface{}, callback WorkflowProgressCallback) (map[string]interface{}, error) {
	startTime := time.Now()
	
	// 创建工作流工作空间
	workflowID := fmt.Sprintf("workflow_%d", time.Now().Unix())
	workflowDir := filepath.Join(we.workspaceDir, workflowID)
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return nil, fmt.Errorf("创建工作流目录失败: %v", err)
	}
	defer os.RemoveAll(workflowDir) // 清理临时目录
	
	// 合并配置
	workflowConfig := we.mergeConfig(template.Config, config)
	
	// 执行结果
	result := &WorkflowResult{
		Success:     true,
		Steps:       make([]StepResult, 0),
		OutputFiles: make([]string, 0),
		Metrics:     make(map[string]interface{}),
	}
	
	// 步骤执行上下文
	stepContext := map[string]interface{}{
		"workspace_dir": workflowDir,
		"config":        workflowConfig,
		"results":       make(map[string]interface{}),
	}
	
	// 执行步骤
	totalSteps := len(template.Steps)
	for i, step := range template.Steps {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("工作流执行被取消")
		default:
		}
		
		stepProgress := int(float64(i) / float64(totalSteps) * 100)
		callback(step.Name, stepProgress, fmt.Sprintf("正在执行步骤: %s", step.Description))
		
		stepResult, err := we.executeStep(ctx, step, stepContext, workflowDir)
		result.Steps = append(result.Steps, *stepResult)
		
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("步骤 %s 执行失败: %v", step.Name, err)
			callback(step.Name, stepProgress, result.Error)
			break
		}
		
		// 更新步骤上下文
		stepContext["results"].(map[string]interface{})[step.Name] = stepResult.Output
	}
	
	result.Duration = time.Since(startTime)
	
	if result.Success {
		callback("完成", 100, "工作流执行完成")
	}
	
	// 转换结果为map格式
	resultMap := map[string]interface{}{
		"success":      result.Success,
		"duration":     result.Duration.Seconds(),
		"steps":        result.Steps,
		"output_files": result.OutputFiles,
		"metrics":      result.Metrics,
	}
	
	if result.Error != "" {
		resultMap["error"] = result.Error
	}
	
	return resultMap, nil
}

// executeStep 执行单个步骤
func (we *WorkflowEngine) executeStep(ctx context.Context, step WorkflowStep, stepContext map[string]interface{}, workflowDir string) (*StepResult, error) {
	startTime := time.Now()
	
	result := &StepResult{
		Name:    step.Name,
		Type:    step.Type,
		Success: false,
		Output:  make(map[string]interface{}),
	}
	
	var err error
	
	switch step.Type {
	case "data_preparation":
		err = we.executeDataPreparation(ctx, step, stepContext, workflowDir, result)
	case "factor_generation":
		err = we.executeFactorGeneration(ctx, step, stepContext, workflowDir, result)
	case "model_training":
		err = we.executeModelTraining(ctx, step, stepContext, workflowDir, result)
	case "strategy_backtest":
		err = we.executeStrategyBacktest(ctx, step, stepContext, workflowDir, result)
	case "result_analysis":
		err = we.executeResultAnalysis(ctx, step, stepContext, workflowDir, result)
	case "report_generation":
		err = we.executeReportGeneration(ctx, step, stepContext, workflowDir, result)
	default:
		err = fmt.Errorf("不支持的步骤类型: %s", step.Type)
	}
	
	result.Duration = time.Since(startTime)
	
	if err != nil {
		result.Error = err.Error()
		return result, err
	}
	
	result.Success = true
	return result, nil
}

// executeDataPreparation 执行数据准备步骤
func (we *WorkflowEngine) executeDataPreparation(ctx context.Context, step WorkflowStep, stepContext map[string]interface{}, workflowDir string, result *StepResult) error {
	script := `
import qlib
import pandas as pd
import json
import sys
from pathlib import Path

def prepare_data(config):
    try:
        # 初始化qlib
        qlib.init(provider_uri=config.get('provider_uri', 'yahoo'), region=config.get('region', 'us'))
        
        # 数据参数
        instruments = config.get('instruments', ['AAPL', 'MSFT', 'GOOGL'])
        start_time = config.get('start_time', '2020-01-01')
        end_time = config.get('end_time', '2023-12-31')
        fields = config.get('fields', ['$close', '$volume', '$high', '$low', '$open'])
        
        # 获取数据
        from qlib.data import D
        data = D.features(instruments, fields, start_time=start_time, end_time=end_time)
        
        # 保存数据
        output_file = Path(config['workspace_dir']) / 'prepared_data.pkl'
        data.to_pickle(str(output_file))
        
        # 统计信息
        stats = {
            'instruments_count': len(instruments),
            'date_range': f"{start_time} to {end_time}",
            'data_shape': data.shape,
            'missing_ratio': data.isnull().sum().sum() / (data.shape[0] * data.shape[1]),
            'output_file': str(output_file)
        }
        
        return {
            'success': True,
            'stats': stats,
            'output_file': str(output_file)
        }
        
    except Exception as e:
        return {
            'success': False,
            'error': str(e)
        }

if __name__ == "__main__":
    config = json.loads(sys.argv[1])
    result = prepare_data(config)
    print(json.dumps(result))
`
	
	return we.executeScript(ctx, script, step.Config, stepContext, result)
}

// executeFactorGeneration 执行因子生成步骤
func (we *WorkflowEngine) executeFactorGeneration(ctx context.Context, step WorkflowStep, stepContext map[string]interface{}, workflowDir string, result *StepResult) error {
	script := `
import qlib
import pandas as pd
import json
import sys
from pathlib import Path

def generate_factors(config):
    try:
        # 加载数据
        data_file = Path(config['workspace_dir']) / 'prepared_data.pkl'
        if not data_file.exists():
            raise ValueError("数据文件不存在，请先执行数据准备步骤")
        
        data = pd.read_pickle(str(data_file))
        
        # 因子表达式
        factor_expressions = config.get('factor_expressions', [
            'Ref($close, 1) / $close - 1',
            'Mean($close, 5) / $close - 1',
            '($high + $low + $close) / 3',
            'Std($close, 20)',
            'Corr($close, $volume, 10)'
        ])
        
        # 计算因子
        from qlib.data.ops import Ref, Mean, Std, Corr
        factors = {}
        
        for i, expr in enumerate(factor_expressions):
            try:
                # 简化的因子计算（实际应使用qlib的表达式引擎）
                if 'Ref' in expr:
                    factors[f'factor_{i+1}'] = data.groupby(level=0).apply(lambda x: x.shift(1))['$close'] / data['$close'] - 1
                elif 'Mean' in expr:
                    factors[f'factor_{i+1}'] = data.groupby(level=0)['$close'].rolling(5).mean() / data['$close'] - 1
                elif 'Std' in expr:
                    factors[f'factor_{i+1}'] = data.groupby(level=0)['$close'].rolling(20).std()
                else:
                    factors[f'factor_{i+1}'] = data['$close']  # 默认值
            except Exception as e:
                print(f"计算因子 {expr} 失败: {e}")
                continue
        
        # 组合因子数据
        factor_data = pd.DataFrame(factors)
        
        # 保存因子数据
        output_file = Path(config['workspace_dir']) / 'factors.pkl'
        factor_data.to_pickle(str(output_file))
        
        # 统计信息
        stats = {
            'factors_count': len(factors),
            'factor_data_shape': factor_data.shape,
            'missing_ratio': factor_data.isnull().sum().sum() / (factor_data.shape[0] * factor_data.shape[1]),
            'output_file': str(output_file)
        }
        
        return {
            'success': True,
            'stats': stats,
            'output_file': str(output_file)
        }
        
    except Exception as e:
        return {
            'success': False,
            'error': str(e)
        }

if __name__ == "__main__":
    config = json.loads(sys.argv[1])
    result = generate_factors(config)
    print(json.dumps(result))
`
	
	return we.executeScript(ctx, script, step.Config, stepContext, result)
}

// executeModelTraining 执行模型训练步骤
func (we *WorkflowEngine) executeModelTraining(ctx context.Context, step WorkflowStep, stepContext map[string]interface{}, workflowDir string, result *StepResult) error {
	script := `
import qlib
import pandas as pd
import json
import sys
from pathlib import Path
import joblib

def train_model(config):
    try:
        # 加载因子数据
        factor_file = Path(config['workspace_dir']) / 'factors.pkl'
        if not factor_file.exists():
            raise ValueError("因子文件不存在，请先执行因子生成步骤")
        
        factor_data = pd.read_pickle(str(factor_file))
        
        # 加载价格数据作为标签
        data_file = Path(config['workspace_dir']) / 'prepared_data.pkl'
        price_data = pd.read_pickle(str(data_file))
        
        # 生成标签（下一期收益率）
        labels = price_data.groupby(level=0)['$close'].apply(lambda x: x.shift(-1) / x - 1)
        
        # 合并数据
        train_data = pd.concat([factor_data, labels.rename('label')], axis=1).dropna()
        
        # 分割训练测试集
        split_date = config.get('split_date', '2022-01-01')
        train_set = train_data[train_data.index.get_level_values(1) < split_date]
        test_set = train_data[train_data.index.get_level_values(1) >= split_date]
        
        # 特征和标签
        feature_cols = [col for col in train_set.columns if col != 'label']
        X_train, y_train = train_set[feature_cols], train_set['label']
        X_test, y_test = test_set[feature_cols], test_set['label']
        
        # 训练模型
        model_type = config.get('model_type', 'lightgbm')
        
        if model_type == 'lightgbm':
            import lightgbm as lgb
            model = lgb.LGBMRegressor(
                n_estimators=config.get('n_estimators', 100),
                learning_rate=config.get('learning_rate', 0.1),
                random_state=42
            )
        else:
            from sklearn.linear_model import LinearRegression
            model = LinearRegression()
        
        model.fit(X_train, y_train)
        
        # 预测和评估
        train_pred = model.predict(X_train)
        test_pred = model.predict(X_test)
        
        from sklearn.metrics import mean_squared_error, r2_score
        
        train_mse = mean_squared_error(y_train, train_pred)
        test_mse = mean_squared_error(y_test, test_pred)
        train_r2 = r2_score(y_train, train_pred)
        test_r2 = r2_score(y_test, test_pred)
        
        # 保存模型
        model_file = Path(config['workspace_dir']) / 'trained_model.pkl'
        joblib.dump(model, str(model_file))
        
        # 保存预测结果
        predictions = pd.DataFrame({
            'actual': y_test,
            'predicted': test_pred
        })
        pred_file = Path(config['workspace_dir']) / 'predictions.pkl'
        predictions.to_pickle(str(pred_file))
        
        # 模型性能
        metrics = {
            'train_mse': train_mse,
            'test_mse': test_mse,
            'train_r2': train_r2,
            'test_r2': test_r2,
            'train_samples': len(X_train),
            'test_samples': len(X_test),
            'features_count': len(feature_cols)
        }
        
        return {
            'success': True,
            'metrics': metrics,
            'model_file': str(model_file),
            'predictions_file': str(pred_file)
        }
        
    except Exception as e:
        return {
            'success': False,
            'error': str(e)
        }

if __name__ == "__main__":
    config = json.loads(sys.argv[1])
    result = train_model(config)
    print(json.dumps(result))
`
	
	return we.executeScript(ctx, script, step.Config, stepContext, result)
}

// executeStrategyBacktest 执行策略回测步骤
func (we *WorkflowEngine) executeStrategyBacktest(ctx context.Context, step WorkflowStep, stepContext map[string]interface{}, workflowDir string, result *StepResult) error {
	script := `
import qlib
import pandas as pd
import numpy as np
import json
import sys
from pathlib import Path

def run_backtest(config):
    try:
        # 加载预测结果
        pred_file = Path(config['workspace_dir']) / 'predictions.pkl'
        if not pred_file.exists():
            raise ValueError("预测文件不存在，请先执行模型训练步骤")
        
        predictions = pd.read_pickle(str(pred_file))
        
        # 策略参数
        top_k = config.get('top_k', 50)
        rebalance_freq = config.get('rebalance_freq', 'monthly')
        
        # 生成交易信号
        signals = predictions['predicted'].unstack()
        
        # 选股策略：选择预测收益最高的前top_k只股票
        positions = signals.apply(lambda x: pd.Series(0, index=x.index), axis=1)
        
        for date in signals.index:
            top_stocks = signals.loc[date].nlargest(top_k).index
            positions.loc[date, top_stocks] = 1 / top_k
        
        # 计算收益
        returns = predictions['actual'].unstack()
        portfolio_returns = (positions.shift(1) * returns).sum(axis=1)
        
        # 计算累积收益
        cumulative_returns = (1 + portfolio_returns).cumprod()
        
        # 基准收益（等权重）
        benchmark_returns = returns.mean(axis=1)
        benchmark_cumulative = (1 + benchmark_returns).cumprod()
        
        # 性能指标
        total_return = cumulative_returns.iloc[-1] - 1
        benchmark_total_return = benchmark_cumulative.iloc[-1] - 1
        
        annual_return = (1 + total_return) ** (252 / len(portfolio_returns)) - 1
        volatility = portfolio_returns.std() * np.sqrt(252)
        sharpe_ratio = annual_return / volatility if volatility > 0 else 0
        
        max_drawdown = ((cumulative_returns.cummax() - cumulative_returns) / cumulative_returns.cummax()).max()
        
        # 保存回测结果
        backtest_results = pd.DataFrame({
            'portfolio_returns': portfolio_returns,
            'benchmark_returns': benchmark_returns,
            'portfolio_cumulative': cumulative_returns,
            'benchmark_cumulative': benchmark_cumulative
        })
        
        results_file = Path(config['workspace_dir']) / 'backtest_results.pkl'
        backtest_results.to_pickle(str(results_file))
        
        # 性能汇总
        performance = {
            'total_return': total_return,
            'benchmark_total_return': benchmark_total_return,
            'annual_return': annual_return,
            'volatility': volatility,
            'sharpe_ratio': sharpe_ratio,
            'max_drawdown': max_drawdown,
            'excess_return': total_return - benchmark_total_return
        }
        
        return {
            'success': True,
            'performance': performance,
            'results_file': str(results_file)
        }
        
    except Exception as e:
        return {
            'success': False,
            'error': str(e)
        }

if __name__ == "__main__":
    config = json.loads(sys.argv[1])
    result = run_backtest(config)
    print(json.dumps(result))
`
	
	return we.executeScript(ctx, script, step.Config, stepContext, result)
}

// executeResultAnalysis 执行结果分析步骤
func (we *WorkflowEngine) executeResultAnalysis(ctx context.Context, step WorkflowStep, stepContext map[string]interface{}, workflowDir string, result *StepResult) error {
	script := `
import pandas as pd
import numpy as np
import json
import sys
from pathlib import Path
import matplotlib.pyplot as plt

def analyze_results(config):
    try:
        # 加载回测结果
        results_file = Path(config['workspace_dir']) / 'backtest_results.pkl'
        if not results_file.exists():
            raise ValueError("回测结果文件不存在，请先执行策略回测步骤")
        
        results = pd.read_pickle(str(results_file))
        
        # 详细分析
        portfolio_returns = results['portfolio_returns']
        benchmark_returns = results['benchmark_returns']
        
        # 年化指标计算
        trading_days = 252
        
        # 夏普比率
        excess_returns = portfolio_returns - benchmark_returns
        sharpe_ratio = excess_returns.mean() / excess_returns.std() * np.sqrt(trading_days)
        
        # 信息比率
        tracking_error = excess_returns.std() * np.sqrt(trading_days)
        information_ratio = excess_returns.mean() * trading_days / tracking_error
        
        # 最大回撤分析
        cumulative = results['portfolio_cumulative']
        rolling_max = cumulative.expanding().max()
        drawdowns = (cumulative - rolling_max) / rolling_max
        max_drawdown = drawdowns.min()
        
        # 计算回撤持续时间
        drawdown_periods = []
        in_drawdown = False
        start_date = None
        
        for date, dd in drawdowns.items():
            if dd < -0.01 and not in_drawdown:  # 进入回撤期
                in_drawdown = True
                start_date = date
            elif dd >= -0.01 and in_drawdown:  # 退出回撤期
                in_drawdown = False
                if start_date:
                    drawdown_periods.append((date - start_date).days)
        
        # 月度收益分析
        monthly_returns = portfolio_returns.resample('M').apply(lambda x: (1 + x).prod() - 1)
        monthly_volatility = monthly_returns.std()
        win_rate = (monthly_returns > 0).mean()
        
        # 生成分析图表
        plt.figure(figsize=(12, 8))
        
        # 累积收益图
        plt.subplot(2, 2, 1)
        plt.plot(results.index, results['portfolio_cumulative'], label='Portfolio')
        plt.plot(results.index, results['benchmark_cumulative'], label='Benchmark')
        plt.title('Cumulative Returns')
        plt.legend()
        
        # 回撤图
        plt.subplot(2, 2, 2)
        plt.fill_between(drawdowns.index, drawdowns, 0, color='red', alpha=0.3)
        plt.title('Drawdowns')
        plt.ylabel('Drawdown %')
        
        # 月度收益分布
        plt.subplot(2, 2, 3)
        plt.hist(monthly_returns, bins=20, alpha=0.7)
        plt.title('Monthly Returns Distribution')
        plt.xlabel('Returns')
        
        # 滚动夏普比率
        plt.subplot(2, 2, 4)
        rolling_sharpe = excess_returns.rolling(60).mean() / excess_returns.rolling(60).std() * np.sqrt(trading_days)
        plt.plot(rolling_sharpe.index, rolling_sharpe)
        plt.title('Rolling Sharpe Ratio (60 days)')
        
        plt.tight_layout()
        chart_file = Path(config['workspace_dir']) / 'analysis_charts.png'
        plt.savefig(str(chart_file))
        plt.close()
        
        # 分析汇总
        analysis_summary = {
            'sharpe_ratio': sharpe_ratio,
            'information_ratio': information_ratio,
            'max_drawdown': max_drawdown,
            'tracking_error': tracking_error,
            'monthly_volatility': monthly_volatility,
            'win_rate': win_rate,
            'avg_drawdown_duration': np.mean(drawdown_periods) if drawdown_periods else 0,
            'max_drawdown_duration': max(drawdown_periods) if drawdown_periods else 0,
        }
        
        # 保存分析结果
        analysis_file = Path(config['workspace_dir']) / 'analysis_summary.json'
        with open(str(analysis_file), 'w') as f:
            json.dump(analysis_summary, f, indent=2, default=str)
        
        return {
            'success': True,
            'analysis': analysis_summary,
            'chart_file': str(chart_file),
            'analysis_file': str(analysis_file)
        }
        
    except Exception as e:
        return {
            'success': False,
            'error': str(e)
        }

if __name__ == "__main__":
    config = json.loads(sys.argv[1])
    result = analyze_results(config)
    print(json.dumps(result))
`
	
	return we.executeScript(ctx, script, step.Config, stepContext, result)
}

// executeReportGeneration 执行报告生成步骤
func (we *WorkflowEngine) executeReportGeneration(ctx context.Context, step WorkflowStep, stepContext map[string]interface{}, workflowDir string, result *StepResult) error {
	script := `
import json
import sys
from pathlib import Path
from datetime import datetime
import pandas as pd

def generate_report(config):
    try:
        workspace = Path(config['workspace_dir'])
        
        # 生成HTML报告
        html_content = generate_html_report(workspace, config)
        
        # 保存报告
        report_file = workspace / 'workflow_report.html'
        with open(str(report_file), 'w', encoding='utf-8') as f:
            f.write(html_content)
        
        return {
            'success': True,
            'report_file': str(report_file)
        }
        
    except Exception as e:
        return {
            'success': False,
            'error': str(e)
        }

def generate_html_report(workspace, config):
    # 加载分析结果
    analysis_file = workspace / 'analysis_summary.json'
    if analysis_file.exists():
        with open(str(analysis_file), 'r') as f:
            analysis = json.load(f)
    else:
        analysis = {}
    
    html = f"""
    <!DOCTYPE html>
    <html>
    <head>
        <title>Qlib工作流报告</title>
        <meta charset="utf-8">
        <style>
            body {{ font-family: Arial, sans-serif; margin: 20px; }}
            .header {{ background-color: #f8f9fa; padding: 20px; border-radius: 5px; }}
            .section {{ margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }}
            .metric {{ display: inline-block; margin: 10px; padding: 10px; background-color: #e9ecef; border-radius: 3px; }}
            .metric-value {{ font-weight: bold; font-size: 1.2em; color: #007bff; }}
            table {{ width: 100%; border-collapse: collapse; margin: 10px 0; }}
            th, td {{ border: 1px solid #ddd; padding: 8px; text-align: left; }}
            th {{ background-color: #f8f9fa; }}
        </style>
    </head>
    <body>
        <div class="header">
            <h1>Qlib量化工作流执行报告</h1>
            <p>生成时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</p>
            <p>工作流名称: {config.get('workflow_name', '未命名工作流')}</p>
        </div>
        
        <div class="section">
            <h2>执行概况</h2>
            <div class="metric">
                <div>总执行时间</div>
                <div class="metric-value">{config.get('duration', 'N/A')}</div>
            </div>
            <div class="metric">
                <div>数据范围</div>
                <div class="metric-value">{config.get('start_time', 'N/A')} - {config.get('end_time', 'N/A')}</div>
            </div>
            <div class="metric">
                <div>股票数量</div>
                <div class="metric-value">{config.get('instruments_count', 'N/A')}</div>
            </div>
        </div>
        
        <div class="section">
            <h2>策略性能</h2>
            <div class="metric">
                <div>夏普比率</div>
                <div class="metric-value">{analysis.get('sharpe_ratio', 'N/A'):.4f}</div>
            </div>
            <div class="metric">
                <div>信息比率</div>
                <div class="metric-value">{analysis.get('information_ratio', 'N/A'):.4f}</div>
            </div>
            <div class="metric">
                <div>最大回撤</div>
                <div class="metric-value">{analysis.get('max_drawdown', 'N/A'):.2%}</div>
            </div>
            <div class="metric">
                <div>月胜率</div>
                <div class="metric-value">{analysis.get('win_rate', 'N/A'):.2%}</div>
            </div>
        </div>
        
        <div class="section">
            <h2>风险指标</h2>
            <table>
                <tr>
                    <th>指标</th>
                    <th>数值</th>
                    <th>说明</th>
                </tr>
                <tr>
                    <td>跟踪误差</td>
                    <td>{analysis.get('tracking_error', 'N/A'):.4f}</td>
                    <td>相对基准的波动率</td>
                </tr>
                <tr>
                    <td>月度波动率</td>
                    <td>{analysis.get('monthly_volatility', 'N/A'):.4f}</td>
                    <td>月收益率标准差</td>
                </tr>
                <tr>
                    <td>平均回撤持续天数</td>
                    <td>{analysis.get('avg_drawdown_duration', 'N/A'):.1f}</td>
                    <td>回撤期平均持续时间</td>
                </tr>
                <tr>
                    <td>最大回撤持续天数</td>
                    <td>{analysis.get('max_drawdown_duration', 'N/A')}</td>
                    <td>最长回撤期持续时间</td>
                </tr>
            </table>
        </div>
        
        <div class="section">
            <h2>工作流步骤</h2>
            <ol>
                <li>数据准备 - 获取和清洗市场数据</li>
                <li>因子生成 - 计算技术指标和因子</li>
                <li>模型训练 - 训练预测模型</li>
                <li>策略回测 - 模拟交易和计算收益</li>
                <li>结果分析 - 计算风险收益指标</li>
                <li>报告生成 - 生成此分析报告</li>
            </ol>
        </div>
        
        <div class="section">
            <h2>免责声明</h2>
            <p>本报告仅供研究参考，不构成投资建议。过往业绩不代表未来表现，投资有风险，决策需谨慎。</p>
        </div>
    </body>
    </html>
    """
    
    return html

if __name__ == "__main__":
    config = json.loads(sys.argv[1])
    result = generate_report(config)
    print(json.dumps(result))
`
	
	return we.executeScript(ctx, script, step.Config, stepContext, result)
}

// executeScript 执行Python脚本
func (we *WorkflowEngine) executeScript(ctx context.Context, script string, stepConfig map[string]interface{}, stepContext map[string]interface{}, result *StepResult) error {
	// 创建临时脚本文件
	scriptFile := filepath.Join(we.scriptDir, fmt.Sprintf("workflow_step_%d.py", time.Now().UnixNano()))
	if err := os.WriteFile(scriptFile, []byte(script), 0755); err != nil {
		return fmt.Errorf("创建脚本文件失败: %v", err)
	}
	defer os.Remove(scriptFile)
	
	// 合并配置
	config := make(map[string]interface{})
	for k, v := range stepContext {
		config[k] = v
	}
	for k, v := range stepConfig {
		config[k] = v
	}
	
	configJSON, _ := json.Marshal(config)
	
	// 执行脚本
	cmd := exec.CommandContext(ctx, we.pythonPath, scriptFile, string(configJSON))
	output, err := cmd.Output()
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("脚本执行失败: %v, stderr: %s", err, string(exitError.Stderr))
		}
		return fmt.Errorf("脚本执行失败: %v", err)
	}
	
	// 解析输出
	var scriptResult map[string]interface{}
	if err := json.Unmarshal(output, &scriptResult); err != nil {
		return fmt.Errorf("解析脚本输出失败: %v, output: %s", err, string(output))
	}
	
	if success, ok := scriptResult["success"].(bool); !ok || !success {
		if errorMsg, ok := scriptResult["error"].(string); ok {
			return fmt.Errorf("脚本执行失败: %s", errorMsg)
		}
		return fmt.Errorf("脚本执行失败: 未知错误")
	}
	
	// 更新结果
	result.Output = scriptResult
	
	return nil
}

// mergeConfig 合并配置
func (we *WorkflowEngine) mergeConfig(templateConfig, userConfig map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	
	// 复制模板配置
	for k, v := range templateConfig {
		merged[k] = v
	}
	
	// 覆盖用户配置
	for k, v := range userConfig {
		merged[k] = v
	}
	
	return merged
}

// GetBuiltinTemplates 获取内置工作流模板
func (we *WorkflowEngine) GetBuiltinTemplates() []WorkflowTemplate {
	return []WorkflowTemplate{
		{
			Name:        "基础量化策略工作流",
			Description: "包含数据准备、因子生成、模型训练、策略回测的完整工作流",
			Category:    "strategy",
			Config: map[string]interface{}{
				"instruments":     []string{"AAPL", "MSFT", "GOOGL", "TSLA", "AMZN"},
				"start_time":      "2020-01-01",
				"end_time":        "2023-12-31",
				"model_type":      "lightgbm",
				"top_k":          50,
				"rebalance_freq": "monthly",
			},
			Steps: []WorkflowStep{
				{
					Name:        "数据准备",
					Type:        "data_preparation",
					Description: "获取和清洗市场数据",
					Required:    true,
				},
				{
					Name:         "因子生成",
					Type:         "factor_generation",
					Description:  "计算技术指标和因子",
					Dependencies: []string{"数据准备"},
					Required:     true,
				},
				{
					Name:         "模型训练",
					Type:         "model_training",
					Description:  "训练预测模型",
					Dependencies: []string{"因子生成"},
					Required:     true,
				},
				{
					Name:         "策略回测",
					Type:         "strategy_backtest",
					Description:  "模拟交易和计算收益",
					Dependencies: []string{"模型训练"},
					Required:     true,
				},
				{
					Name:         "结果分析",
					Type:         "result_analysis",
					Description:  "计算风险收益指标",
					Dependencies: []string{"策略回测"},
					Required:     true,
				},
				{
					Name:         "报告生成",
					Type:         "report_generation",
					Description:  "生成分析报告",
					Dependencies: []string{"结果分析"},
					Required:     false,
				},
			},
		},
		{
			Name:        "因子研究工作流",
			Description: "专注于因子挖掘和分析的工作流",
			Category:    "research",
			Config: map[string]interface{}{
				"instruments": []string{"SPY", "QQQ", "IWM"},
				"start_time":  "2019-01-01",
				"end_time":    "2023-12-31",
			},
			Steps: []WorkflowStep{
				{
					Name:        "数据准备",
					Type:        "data_preparation",
					Description: "获取市场数据",
					Required:    true,
				},
				{
					Name:         "因子生成",
					Type:         "factor_generation",
					Description:  "生成多种技术因子",
					Dependencies: []string{"数据准备"},
					Required:     true,
				},
				{
					Name:         "因子分析",
					Type:         "result_analysis",
					Description:  "分析因子有效性",
					Dependencies: []string{"因子生成"},
					Required:     true,
				},
			},
		},
	}
}