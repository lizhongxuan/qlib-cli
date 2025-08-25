package qlib

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// FactorCalculator 因子计算接口
type FactorCalculator struct {
	client *QlibClient
}

// FactorExpression 因子表达式
type FactorExpression struct {
	Name       string `json:"name"`
	Expression string `json:"expression"`
	Universe   string `json:"universe"`   // 股票池
	Frequency  string `json:"frequency"`  // 计算频率
	StartDate  string `json:"start_date"` // 开始日期
	EndDate    string `json:"end_date"`   // 结束日期
}

// FactorResult 因子计算结果
type FactorResult struct {
	Success    bool                           `json:"success"`
	FactorName string                         `json:"factor_name"`
	Data       []FactorValue                  `json:"data"`
	Stats      map[string]float64             `json:"stats"`
	Error      string                         `json:"error"`
	Metadata   map[string]interface{}         `json:"metadata"`
}

// FactorValue 因子值
type FactorValue struct {
	Instrument string    `json:"instrument"`
	Date       time.Time `json:"date"`
	Value      float64   `json:"value"`
	IsValid    bool      `json:"is_valid"`
}

// FactorPerformance 因子性能分析结果
type FactorPerformance struct {
	IC         map[string]float64 `json:"ic"`          // 信息系数
	ICIR       float64            `json:"icir"`        // 信息比率
	RankIC     map[string]float64 `json:"rank_ic"`     // 排序信息系数
	Turnover   float64            `json:"turnover"`    // 换手率
	Coverage   float64            `json:"coverage"`    // 覆盖率
	Statistics map[string]float64 `json:"statistics"`  // 其他统计指标
}

// NewFactorCalculator 创建因子计算器
func NewFactorCalculator(client *QlibClient) *FactorCalculator {
	return &FactorCalculator{
		client: client,
	}
}

// CalculateFactor 计算单个因子
func (fc *FactorCalculator) CalculateFactor(ctx context.Context, expr FactorExpression) (*FactorResult, error) {
	if !fc.client.IsInitialized() {
		return nil, fmt.Errorf("Qlib客户端未初始化")
	}

	log.Printf("正在计算因子: %s", expr.Name)

	// 构建因子计算脚本
	script := fc.buildFactorScript(expr)

	// 执行脚本
	output, err := fc.client.ExecuteScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("计算因子失败: %w", err)
	}

	// 解析结果
	var result FactorResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("解析因子结果失败: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("因子计算失败: %s", result.Error)
	}

	log.Printf("因子 %s 计算完成，共 %d 个数据点", expr.Name, len(result.Data))
	return &result, nil
}

// buildFactorScript 构建因子计算脚本
func (fc *FactorCalculator) buildFactorScript(expr FactorExpression) string {
	return fmt.Sprintf(`
import json
import qlib
from qlib import data
import pandas as pd
import numpy as np
from datetime import datetime

try:
	# 因子参数
	factor_name = '%s'
	expression = '''%s'''
	universe = '%s'
	frequency = '%s'
	start_date = '%s'
	end_date = '%s'
	
	# 设置股票池
	if universe == 'csi300':
		instruments = data.D.instruments(market='csi300')
	elif universe == 'csi500':
		instruments = data.D.instruments(market='csi500')
	elif universe == 'all':
		instruments = data.D.instruments(market='all')
	else:
		# 默认使用CSI300
		instruments = data.D.instruments(market='csi300')
	
	# 计算因子
	if frequency == 'day':
		factor_data = data.D.features(
			instruments=instruments,
			fields=[expression],
			start_time=start_date,
			end_time=end_date,
			freq='day'
		)
	elif frequency == 'minute':
		factor_data = data.D.features(
			instruments=instruments,
			fields=[expression],
			start_time=start_date,
			end_time=end_date,
			freq='1min'
		)
	else:
		factor_data = data.D.features(
			instruments=instruments,
			fields=[expression],
			start_time=start_date,
			end_time=end_date,
			freq='day'
		)
	
	# 转换数据格式
	factor_values = []
	stats = {}
	
	if factor_data is not None and not factor_data.empty:
		# 重置索引
		factor_reset = factor_data.reset_index()
		
		for _, row in factor_reset.iterrows():
			instrument = row['instrument'] if 'instrument' in row else str(row.name[0])
			date = row['datetime'] if 'datetime' in row else str(row.name[1])
			value = row[expression] if expression in row else None
			
			if pd.notna(value):
				factor_values.append({
					'instrument': instrument,
					'date': str(date),
					'value': float(value),
					'is_valid': True
				})
			else:
				factor_values.append({
					'instrument': instrument,
					'date': str(date),
					'value': 0.0,
					'is_valid': False
				})
		
		# 计算统计指标
		valid_values = [fv['value'] for fv in factor_values if fv['is_valid']]
		if valid_values:
			stats = {
				'count': len(valid_values),
				'mean': float(np.mean(valid_values)),
				'std': float(np.std(valid_values)),
				'min': float(np.min(valid_values)),
				'max': float(np.max(valid_values)),
				'median': float(np.median(valid_values)),
				'skew': float(pd.Series(valid_values).skew()),
				'kurt': float(pd.Series(valid_values).kurtosis()),
				'coverage': len(valid_values) / len(factor_values)
			}
	
	result = {
		'success': True,
		'factor_name': factor_name,
		'data': factor_values,
		'stats': stats,
		'metadata': {
			'expression': expression,
			'universe': universe,
			'frequency': frequency,
			'start_date': start_date,
			'end_date': end_date,
			'total_points': len(factor_values)
		}
	}

except Exception as e:
	result = {
		'success': False,
		'factor_name': factor_name,
		'data': [],
		'stats': {},
		'error': str(e),
		'metadata': {}
	}

print(json.dumps(result, default=str))
`,
		expr.Name, expr.Expression, expr.Universe, expr.Frequency, expr.StartDate, expr.EndDate)
}

// CalculateFactorPerformance 计算因子性能
func (fc *FactorCalculator) CalculateFactorPerformance(ctx context.Context, factorName string, factorData []FactorValue, returnData []FactorValue) (*FactorPerformance, error) {
	// 构建性能分析脚本
	script := fc.buildPerformanceScript(factorName, factorData, returnData)

	// 执行脚本
	output, err := fc.client.ExecuteScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("计算因子性能失败: %w", err)
	}

	// 解析结果
	var result struct {
		Success     bool              `json:"success"`
		Performance FactorPerformance `json:"performance"`
		Error       string            `json:"error"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("解析性能结果失败: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("性能计算失败: %s", result.Error)
	}

	return &result.Performance, nil
}

// buildPerformanceScript 构建性能分析脚本
func (fc *FactorCalculator) buildPerformanceScript(factorName string, factorData, returnData []FactorValue) string {
	factorJson, _ := json.Marshal(factorData)
	returnJson, _ := json.Marshal(returnData)

	return fmt.Sprintf(`
import json
import pandas as pd
import numpy as np
from scipy import stats

try:
	# 解析数据
	factor_data = json.loads('''%s''')
	return_data = json.loads('''%s''')
	
	# 构建DataFrame
	factor_df = pd.DataFrame(factor_data)
	return_df = pd.DataFrame(return_data)
	
	# 转换日期
	factor_df['date'] = pd.to_datetime(factor_df['date'])
	return_df['date'] = pd.to_datetime(return_df['date'])
	
	# 合并数据
	merged = pd.merge(factor_df, return_df, on=['instrument', 'date'], suffixes=('_factor', '_return'))
	merged = merged[(merged['is_valid_factor']) & (merged['is_valid_return'])]
	
	# 计算信息系数
	ic_values = {}
	rank_ic_values = {}
	
	# 按日期分组计算IC
	for date, group in merged.groupby('date'):
		if len(group) >= 10:  # 至少需要10个样本
			# 普通IC
			ic = stats.pearsonr(group['value_factor'], group['value_return'])[0]
			if not np.isnan(ic):
				ic_values[str(date)] = ic
			
			# 排序IC
			rank_ic = stats.spearmanr(group['value_factor'], group['value_return'])[0]
			if not np.isnan(rank_ic):
				rank_ic_values[str(date)] = rank_ic
	
	# 计算统计指标
	ic_list = list(ic_values.values())
	rank_ic_list = list(rank_ic_values.values())
	
	ic_mean = np.mean(ic_list) if ic_list else 0
	ic_std = np.std(ic_list) if ic_list else 0
	icir = ic_mean / ic_std if ic_std > 0 else 0
	
	rank_ic_mean = np.mean(rank_ic_list) if rank_ic_list else 0
	rank_ic_std = np.std(rank_ic_list) if rank_ic_list else 0
	
	# 计算覆盖率
	total_possible = len(merged.groupby('date'))
	coverage = len(ic_values) / total_possible if total_possible > 0 else 0
	
	# 计算换手率（简化版本）
	turnover = 0.5  # 占位值，实际需要根据持仓变化计算
	
	performance = {
		'ic': {
			'mean': ic_mean,
			'std': ic_std,
			'positive_rate': sum(1 for ic in ic_list if ic > 0) / len(ic_list) if ic_list else 0,
			'values': ic_values
		},
		'icir': icir,
		'rank_ic': {
			'mean': rank_ic_mean,
			'std': rank_ic_std,
			'positive_rate': sum(1 for ic in rank_ic_list if ic > 0) / len(rank_ic_list) if rank_ic_list else 0,
			'values': rank_ic_values
		},
		'turnover': turnover,
		'coverage': coverage,
		'statistics': {
			'sample_count': len(merged),
			'date_count': len(ic_values),
			'avg_stocks_per_day': len(merged) / len(ic_values) if ic_values else 0
		}
	}
	
	result = {
		'success': True,
		'performance': performance
	}

except Exception as e:
	result = {
		'success': False,
		'performance': {},
		'error': str(e)
	}

print(json.dumps(result, default=str))
`,
		string(factorJson), string(returnJson))
}

// BatchCalculateFactors 批量计算因子
func (fc *FactorCalculator) BatchCalculateFactors(ctx context.Context, expressions []FactorExpression) ([]FactorResult, error) {
	results := make([]FactorResult, 0, len(expressions))

	for _, expr := range expressions {
		result, err := fc.CalculateFactor(ctx, expr)
		if err != nil {
			log.Printf("因子 %s 计算失败: %v", expr.Name, err)
			// 添加失败的结果
			results = append(results, FactorResult{
				Success:    false,
				FactorName: expr.Name,
				Error:      err.Error(),
			})
		} else {
			results = append(results, *result)
		}
	}

	return results, nil
}

// ValidateFactorExpression 验证因子表达式
func (fc *FactorCalculator) ValidateFactorExpression(ctx context.Context, expression string) (bool, error) {
	script := fmt.Sprintf(`
import json
import qlib
from qlib import data

try:
	expression = '''%s'''
	
	# 尝试解析表达式
	# 这里使用简单的测试数据验证表达式语法
	test_instruments = ['000001.SZ']  # 使用单个股票测试
	test_start = '2023-01-01'
	test_end = '2023-01-02'
	
	# 尝试计算表达式
	test_data = data.D.features(
		instruments=test_instruments,
		fields=[expression],
		start_time=test_start,
		end_time=test_end,
		freq='day'
	)
	
	# 如果没有异常，则表达式有效
	result = {
		'success': True,
		'valid': True,
		'message': '表达式语法有效'
	}

except Exception as e:
	result = {
		'success': True,
		'valid': False,
		'message': str(e)
	}

print(json.dumps(result))
`, expression)

	output, err := fc.client.ExecuteScript(ctx, script)
	if err != nil {
		return false, fmt.Errorf("验证因子表达式失败: %w", err)
	}

	var response struct {
		Success bool   `json:"success"`
		Valid   bool   `json:"valid"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return false, fmt.Errorf("解析验证结果失败: %w", err)
	}

	if !response.Success {
		return false, fmt.Errorf("验证过程失败")
	}

	return response.Valid, nil
}

// GetBuiltinFactors 获取内置因子列表
func (fc *FactorCalculator) GetBuiltinFactors(ctx context.Context) (map[string][]string, error) {
	script := `
import json

try:
	# 常用的内置因子分类
	builtin_factors = {
		'price': [
			'$open', '$high', '$low', '$close', '$volume', '$money',
			'Ref($close, 1)', 'Ref($open, 1)', 'Ref($high, 1)', 'Ref($low, 1)'
		],
		'technical': [
			'($close - Ref($close, 1)) / Ref($close, 1)',  # 收益率
			'($high + $low + $close) / 3',  # 典型价格
			'($high - $low) / $close',  # 价格波动率
			'Mean($volume, 5)', 'Mean($volume, 10)', 'Mean($volume, 20)',  # 成交量均线
			'Std($close, 5)', 'Std($close, 10)', 'Std($close, 20)',  # 价格标准差
			'($close - Mean($close, 20)) / Std($close, 20)',  # 标准化价格
		],
		'momentum': [
			'$close / Ref($close, 1) - 1',  # 1日收益率
			'$close / Ref($close, 5) - 1',  # 5日收益率  
			'$close / Ref($close, 10) - 1', # 10日收益率
			'$close / Ref($close, 20) - 1', # 20日收益率
			'Corr($close, Log($volume), 10)',  # 价量相关性
		],
		'volatility': [
			'Std($close, 5) / Mean($close, 5)',   # 5日波动率
			'Std($close, 10) / Mean($close, 10)', # 10日波动率
			'Std($close, 20) / Mean($close, 20)', # 20日波动率
			'($high - $low) / $open',  # 当日波动率
			'Max($high, 5) / Min($low, 5) - 1',  # 5日价格区间
		],
		'volume': [
			'$volume / Mean($volume, 5)',   # 相对成交量5日
			'$volume / Mean($volume, 10)',  # 相对成交量10日  
			'$volume / Mean($volume, 20)',  # 相对成交量20日
			'$money / $volume',  # 平均成交价
			'Corr($volume, $close, 10)',  # 量价相关性
		],
		'reversal': [
			'Rank($close, 5)',   # 5日排序
			'Rank($close, 10)',  # 10日排序
			'Rank($close, 20)',  # 20日排序
			'($close - Mean($close, 20)) / Std($close, 20)',  # Z分数
		]
	}
	
	result = {
		'success': True,
		'factors': builtin_factors
	}

except Exception as e:
	result = {
		'success': False,
		'factors': {},
		'error': str(e)
	}

print(json.dumps(result))
`

	output, err := fc.client.ExecuteScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("获取内置因子失败: %w", err)
	}

	var response struct {
		Success bool                `json:"success"`
		Factors map[string][]string `json:"factors"`
		Error   string              `json:"error"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("解析内置因子失败: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("获取内置因子失败: %s", response.Error)
	}

	return response.Factors, nil
}

// GetFactorCorrelation 计算因子相关性
func (fc *FactorCalculator) GetFactorCorrelation(ctx context.Context, factor1, factor2 []FactorValue) (float64, error) {
	factor1Json, _ := json.Marshal(factor1)
	factor2Json, _ := json.Marshal(factor2)

	script := fmt.Sprintf(`
import json
import pandas as pd
import numpy as np
from scipy import stats

try:
	factor1_data = json.loads('''%s''')
	factor2_data = json.loads('''%s''')
	
	# 构建DataFrame
	df1 = pd.DataFrame(factor1_data)
	df2 = pd.DataFrame(factor2_data)
	
	# 转换日期
	df1['date'] = pd.to_datetime(df1['date'])
	df2['date'] = pd.to_datetime(df2['date'])
	
	# 合并数据
	merged = pd.merge(df1, df2, on=['instrument', 'date'], suffixes=('_1', '_2'))
	merged = merged[(merged['is_valid_1']) & (merged['is_valid_2'])]
	
	# 计算相关性
	if len(merged) > 0:
		correlation = stats.pearsonr(merged['value_1'], merged['value_2'])[0]
		if np.isnan(correlation):
			correlation = 0.0
	else:
		correlation = 0.0
	
	result = {
		'success': True,
		'correlation': float(correlation),
		'sample_size': len(merged)
	}

except Exception as e:
	result = {
		'success': False,
		'correlation': 0.0,
		'error': str(e)
	}

print(json.dumps(result, default=str))
`,
		string(factor1Json), string(factor2Json))

	output, err := fc.client.ExecuteScript(ctx, script)
	if err != nil {
		return 0, fmt.Errorf("计算因子相关性失败: %w", err)
	}

	var response struct {
		Success     bool    `json:"success"`
		Correlation float64 `json:"correlation"`
		Error       string  `json:"error"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return 0, fmt.Errorf("解析相关性结果失败: %w", err)
	}

	if !response.Success {
		return 0, fmt.Errorf("计算相关性失败: %s", response.Error)
	}

	return response.Correlation, nil
}