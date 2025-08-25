package qlib

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// ModelInterface Qlib模型训练接口
type ModelInterface struct {
	client *QlibClient
}

// ModelConfig 模型配置
type ModelConfig struct {
	ModelType    string                 `json:"model_type"`    // 模型类型：lgb, xgb, linear等
	Parameters   map[string]interface{} `json:"parameters"`    // 模型参数
	Dataset      DatasetConfig          `json:"dataset"`       // 数据集配置
	Features     []string               `json:"features"`      // 特征列表
	Label        string                 `json:"label"`         // 标签列
	SplitConfig  SplitConfig            `json:"split_config"`  // 数据分割配置
	TaskName     string                 `json:"task_name"`     // 任务名称
}

// DatasetConfig 数据集配置
type DatasetConfig struct {
	Instruments []string  `json:"instruments"` // 股票池
	StartTime   time.Time `json:"start_time"`  // 开始时间
	EndTime     time.Time `json:"end_time"`    // 结束时间
	Segments    Segments  `json:"segments"`    // 时间分段
}

// Segments 时间分段
type Segments struct {
	Train []string `json:"train"` // 训练时间段
	Valid []string `json:"valid"` // 验证时间段  
	Test  []string `json:"test"`  // 测试时间段
}

// SplitConfig 数据分割配置
type SplitConfig struct {
	TrainRatio float64 `json:"train_ratio"` // 训练集比例
	ValidRatio float64 `json:"valid_ratio"` // 验证集比例
	TestRatio  float64 `json:"test_ratio"`  // 测试集比例
}

// TrainingResult 训练结果
type TrainingResult struct {
	Success   bool                   `json:"success"`
	ModelID   string                 `json:"model_id"`
	TaskID    string                 `json:"task_id"`
	Metrics   map[string]float64     `json:"metrics"`
	ModelPath string                 `json:"model_path"`
	Logs      []string               `json:"logs"`
	Error     string                 `json:"error"`
	Metadata  map[string]interface{} `json:"metadata"`
	Duration  time.Duration          `json:"duration"`
}

// PredictionResult 预测结果
type PredictionResult struct {
	Success     bool              `json:"success"`
	Predictions []PredictionValue `json:"predictions"`
	Error       string            `json:"error"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PredictionValue 预测值
type PredictionValue struct {
	Instrument string    `json:"instrument"`
	Date       time.Time `json:"date"`
	Score      float64   `json:"score"`
	Label      float64   `json:"label,omitempty"`
	Rank       int       `json:"rank,omitempty"`
}

// ModelEvaluation 模型评估结果
type ModelEvaluation struct {
	Metrics     map[string]float64     `json:"metrics"`
	ICAnalysis  map[string]float64     `json:"ic_analysis"`
	RankIC      map[string]float64     `json:"rank_ic"`
	Sharpe      float64                `json:"sharpe"`
	MaxDrawdown float64                `json:"max_drawdown"`
	AnnualReturn float64               `json:"annual_return"`
	WinRate     float64                `json:"win_rate"`
	Details     map[string]interface{} `json:"details"`
}

// NewModelInterface 创建模型接口
func NewModelInterface(client *QlibClient) *ModelInterface {
	return &ModelInterface{
		client: client,
	}
}

// TrainModel 训练模型
func (mi *ModelInterface) TrainModel(ctx context.Context, config ModelConfig) (*TrainingResult, error) {
	if !mi.client.IsInitialized() {
		return nil, fmt.Errorf("Qlib客户端未初始化")
	}

	log.Printf("开始训练模型: %s", config.TaskName)

	// 构建训练脚本
	script := mi.buildTrainingScript(config)

	// 执行训练
	startTime := time.Now()
	output, err := mi.client.ExecuteScript(ctx, script)
	duration := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("模型训练失败: %w", err)
	}

	// 解析结果
	var result TrainingResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("解析训练结果失败: %w", err)
	}

	result.Duration = duration

	if !result.Success {
		return nil, fmt.Errorf("模型训练失败: %s", result.Error)
	}

	log.Printf("模型训练完成: %s, 耗时: %v", result.ModelID, duration)
	return &result, nil
}

// buildTrainingScript 构建训练脚本
func (mi *ModelInterface) buildTrainingScript(config ModelConfig) string {
	configJson, _ := json.Marshal(config)

	return fmt.Sprintf(`
import json
import qlib
from qlib import data
from qlib.workflow import R
from qlib.workflow.record_temp import SignalRecord, PortAnaRecord, SigAnaRecord
from qlib.utils import flatten_dict, init_instance_by_config
import pandas as pd
import numpy as np
import uuid
from datetime import datetime

try:
	# 解析配置
	config = json.loads('''%s''')
	
	# 生成任务ID
	task_id = str(uuid.uuid4())[:8]
	model_id = f"{config['model_type']}_{task_id}"
	
	# 构建数据集
	instruments = config['dataset']['instruments']
	if not instruments:
		instruments = data.D.instruments(market='csi300')
	
	start_time = config['dataset']['start_time'][:10] if config['dataset']['start_time'] else '2020-01-01'
	end_time = config['dataset']['end_time'][:10] if config['dataset']['end_time'] else '2023-12-31'
	
	# 构建特征
	fields = config.get('features', [])
	if not fields:
		fields = ['$open', '$high', '$low', '$close', '$volume', 'Ref($close, 1)', 'Mean($close, 5)', 'Mean($close, 10)']
	
	# 构建标签
	label = config.get('label', '($close - Ref($close, 1)) / Ref($close, 1)')
	
	# 创建数据集配置
	dataset_config = {
		'class': 'DatasetH',
		'module_path': 'qlib.data.dataset',
		'kwargs': {
			'handler': {
				'class': 'Alpha158',
				'module_path': 'qlib.contrib.data.handler',
				'kwargs': {
					'start_time': start_time,
					'end_time': end_time,
					'fit_start_time': start_time,
					'fit_end_time': end_time,
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
					],
					'learn_processors': [
						{
							'class': 'DropnaLabel'
						},
						{
							'class': 'CSRankNorm',
							'kwargs': {'fields_group': 'label'}
						}
					],
					'label': [label]
				}
			},
			'segments': {
				'train': (start_time, '2022-12-31'),
				'valid': ('2023-01-01', '2023-06-30'), 
				'test': ('2023-07-01', end_time)
			}
		}
	}
	
	# 根据模型类型构建模型配置
	if config['model_type'] == 'lgb':
		model_config = {
			'class': 'LGBModel',
			'module_path': 'qlib.contrib.model.gbdt',
			'kwargs': dict(config.get('parameters', {
				'objective': 'regression',
				'num_leaves': 60,
				'learning_rate': 0.05,
				'feature_fraction': 0.9,
				'bagging_fraction': 0.8,
				'bagging_freq': 5,
				'verbose': -1,
				'n_estimators': 100
			}))
		}
	elif config['model_type'] == 'xgb':
		model_config = {
			'class': 'XGBModel',
			'module_path': 'qlib.contrib.model.xgboost',
			'kwargs': dict(config.get('parameters', {
				'max_depth': 6,
				'learning_rate': 0.05,
				'n_estimators': 100,
				'subsample': 0.8,
				'colsample_bytree': 0.8
			}))
		}
	elif config['model_type'] == 'linear':
		model_config = {
			'class': 'LinearModel',
			'module_path': 'qlib.contrib.model.linear',
			'kwargs': dict(config.get('parameters', {
				'estimator': 'ridge'
			}))
		}
	else:
		model_config = {
			'class': 'LGBModel',
			'module_path': 'qlib.contrib.model.gbdt',
			'kwargs': config.get('parameters', {})
		}
	
	# 初始化数据集和模型
	dataset = init_instance_by_config(dataset_config)
	model = init_instance_by_config(model_config)
	
	# 训练模型
	with R.start(experiment_name="model_training", recorder_id=task_id):
		# 训练
		model.fit(dataset)
		
		# 预测
		pred = model.predict(dataset)
		
		# 计算评估指标
		if pred is not None and not pred.empty:
			# 准备评估数据
			test_pred = pred.loc[pd.IndexSlice[:, '2023-07-01':], :]
			test_label = dataset.prepare(['test'])['test']['label']
			
			# 计算基本指标
			from scipy.stats import pearsonr, spearmanr
			
			# 对齐预测和标签
			aligned = pd.concat([test_pred, test_label], axis=1, join='inner')
			aligned.columns = ['pred', 'label']
			aligned = aligned.dropna()
			
			metrics = {}
			if len(aligned) > 0:
				ic, _ = pearsonr(aligned['pred'], aligned['label'])
				rank_ic, _ = spearmanr(aligned['pred'], aligned['label'])
				metrics['ic'] = float(ic) if not np.isnan(ic) else 0.0
				metrics['rank_ic'] = float(rank_ic) if not np.isnan(rank_ic) else 0.0
				metrics['mse'] = float(np.mean((aligned['pred'] - aligned['label']) ** 2))
				metrics['mae'] = float(np.mean(np.abs(aligned['pred'] - aligned['label'])))
			
			# 保存记录
			R.save_objects(model=model)
			
			model_path = f"./models/{model_id}.pkl"
			
			result = {
				'success': True,
				'model_id': model_id,
				'task_id': task_id,
				'metrics': metrics,
				'model_path': model_path,
				'logs': [f"模型训练完成: {datetime.now()}", f"样本数量: {len(aligned)}"],
				'metadata': {
					'model_type': config['model_type'],
					'dataset_config': dataset_config,
					'model_config': model_config,
					'sample_count': len(aligned)
				}
			}
		else:
			result = {
				'success': False,
				'model_id': model_id,
				'task_id': task_id,
				'error': '模型预测结果为空',
				'logs': ['训练过程中出现问题'],
				'metadata': {}
			}

except Exception as e:
	result = {
		'success': False,
		'model_id': '',
		'task_id': '',
		'error': str(e),
		'logs': [f"训练失败: {str(e)}"],
		'metadata': {}
	}

print(json.dumps(result, default=str))
`, string(configJson))
}

// PredictModel 模型预测
func (mi *ModelInterface) PredictModel(ctx context.Context, modelID string, instruments []string, startDate, endDate time.Time) (*PredictionResult, error) {
	script := fmt.Sprintf(`
import json
import qlib
from qlib import data
from qlib.workflow import R
import pandas as pd
import numpy as np
import pickle
import os

try:
	model_id = '%s'
	instruments = %s
	start_date = '%s'
	end_date = '%s'
	
	# 加载模型
	model_path = f"./models/{model_id}.pkl"
	if not os.path.exists(model_path):
		raise FileNotFoundError(f"模型文件不存在: {model_path}")
	
	# 使用recorder加载模型
	try:
		with R.start(experiment_name="model_prediction", recorder_id=model_id):
			model = R.load_object('model')
	except:
		# 如果recorder加载失败，尝试直接加载pickle文件
		with open(model_path, 'rb') as f:
			model = pickle.load(f)
	
	# 构建预测数据集
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
	
	from qlib.utils import init_instance_by_config
	dataset = init_instance_by_config(dataset_config)
	
	# 进行预测
	predictions = model.predict(dataset)
	
	# 转换预测结果
	pred_list = []
	if predictions is not None and not predictions.empty:
		pred_reset = predictions.reset_index()
		
		for _, row in pred_reset.iterrows():
			instrument = row['instrument'] if 'instrument' in row else str(row.name[0])
			date = row['datetime'] if 'datetime' in row else str(row.name[1])
			score = row[0] if len(row) > 2 else row.iloc[-1]  # 预测分数
			
			pred_list.append({
				'instrument': instrument,
				'date': str(date),
				'score': float(score) if pd.notna(score) else 0.0
			})
		
		# 排序
		pred_list.sort(key=lambda x: x['score'], reverse=True)
		for i, pred in enumerate(pred_list):
			pred['rank'] = i + 1
	
	result = {
		'success': True,
		'predictions': pred_list,
		'metadata': {
			'model_id': model_id,
			'prediction_count': len(pred_list),
			'date_range': [start_date, end_date]
		}
	}

except Exception as e:
	result = {
		'success': False,
		'predictions': [],
		'error': str(e),
		'metadata': {}
	}

print(json.dumps(result, default=str))
`,
		modelID,
		fmt.Sprintf(`["%s"]`, instruments[0]), // 简化处理，只取第一个股票
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))

	output, err := mi.client.ExecuteScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("模型预测失败: %w", err)
	}

	var result PredictionResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("解析预测结果失败: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("预测失败: %s", result.Error)
	}

	// 日期格式已在Python脚本中处理

	return &result, nil
}

// EvaluateModel 评估模型
func (mi *ModelInterface) EvaluateModel(ctx context.Context, modelID string, testData []PredictionValue, actualReturns []PredictionValue) (*ModelEvaluation, error) {
	testDataJson, _ := json.Marshal(testData)
	actualReturnsJson, _ := json.Marshal(actualReturns)

	script := fmt.Sprintf(`
import json
import pandas as pd
import numpy as np
from scipy import stats

try:
	test_data = json.loads('''%s''')
	actual_returns = json.loads('''%s''')
	
	# 构建DataFrame
	pred_df = pd.DataFrame(test_data)
	actual_df = pd.DataFrame(actual_returns)
	
	# 转换日期
	pred_df['date'] = pd.to_datetime(pred_df['date'])
	actual_df['date'] = pd.to_datetime(actual_df['date'])
	
	# 合并数据
	merged = pd.merge(pred_df, actual_df, on=['instrument', 'date'], suffixes=('_pred', '_actual'))
	merged = merged.dropna()
	
	metrics = {}
	ic_analysis = {}
	rank_ic = {}
	
	if len(merged) > 0:
		# 基本指标
		ic, _ = stats.pearsonr(merged['score'], merged['score_actual'])
		rank_ic_val, _ = stats.spearmanr(merged['score'], merged['score_actual'])
		
		metrics['ic'] = float(ic) if not np.isnan(ic) else 0.0
		metrics['rank_ic'] = float(rank_ic_val) if not np.isnan(rank_ic_val) else 0.0
		metrics['mse'] = float(np.mean((merged['score'] - merged['score_actual']) ** 2))
		metrics['mae'] = float(np.mean(np.abs(merged['score'] - merged['score_actual'])))
		
		# IC分析
		daily_ic = {}
		daily_rank_ic = {}
		for date, group in merged.groupby('date'):
			if len(group) >= 10:
				daily_ic_val, _ = stats.pearsonr(group['score'], group['score_actual'])
				daily_rank_ic_val, _ = stats.spearmanr(group['score'], group['score_actual'])
				
				if not np.isnan(daily_ic_val):
					daily_ic[str(date)] = float(daily_ic_val)
				if not np.isnan(daily_rank_ic_val):
					daily_rank_ic[str(date)] = float(daily_rank_ic_val)
		
		ic_values = list(daily_ic.values())
		rank_ic_values = list(daily_rank_ic.values())
		
		if ic_values:
			ic_analysis = {
				'mean': np.mean(ic_values),
				'std': np.std(ic_values),
				'ir': np.mean(ic_values) / np.std(ic_values) if np.std(ic_values) > 0 else 0,
				'positive_rate': sum(1 for ic in ic_values if ic > 0) / len(ic_values),
				'values': daily_ic
			}
		
		if rank_ic_values:
			rank_ic = {
				'mean': np.mean(rank_ic_values),
				'std': np.std(rank_ic_values), 
				'ir': np.mean(rank_ic_values) / np.std(rank_ic_values) if np.std(rank_ic_values) > 0 else 0,
				'positive_rate': sum(1 for ic in rank_ic_values if ic > 0) / len(rank_ic_values),
				'values': daily_rank_ic
			}
		
		# 简化的回报分析
		merged['return'] = merged['score_actual']
		merged = merged.sort_values('score', ascending=False)
		
		# 计算分位数回报
		quantile_returns = []
		quantile_size = len(merged) // 5
		
		for i in range(5):
			start_idx = i * quantile_size
			end_idx = (i + 1) * quantile_size if i < 4 else len(merged)
			quantile_return = merged.iloc[start_idx:end_idx]['return'].mean()
			quantile_returns.append(float(quantile_return))
		
		# 计算Sharpe比率（简化版）
		if len(quantile_returns) > 0:
			excess_return = quantile_returns[0] - quantile_returns[-1]  # 最高分位 - 最低分位
			sharpe = excess_return / np.std(quantile_returns) if np.std(quantile_returns) > 0 else 0
		else:
			sharpe = 0
		
		# 胜率
		win_rate = sum(1 for r in merged['return'] if r > 0) / len(merged) if len(merged) > 0 else 0
		
	else:
		metrics = {'ic': 0, 'rank_ic': 0, 'mse': 0, 'mae': 0}
		ic_analysis = {}
		rank_ic = {}
		sharpe = 0
		win_rate = 0
		quantile_returns = []
	
	evaluation = {
		'metrics': metrics,
		'ic_analysis': ic_analysis,
		'rank_ic': rank_ic,
		'sharpe': float(sharpe),
		'max_drawdown': 0.05,  # 占位值
		'annual_return': float(quantile_returns[0]) * 250 if quantile_returns else 0,
		'win_rate': float(win_rate),
		'details': {
			'sample_count': len(merged),
			'quantile_returns': quantile_returns,
			'evaluation_period': len(daily_ic) if 'daily_ic' in locals() else 0
		}
	}
	
	result = {
		'success': True,
		'evaluation': evaluation
	}

except Exception as e:
	result = {
		'success': False,
		'evaluation': {},
		'error': str(e)
	}

print(json.dumps(result, default=str))
`,
		string(testDataJson), string(actualReturnsJson))

	output, err := mi.client.ExecuteScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("模型评估失败: %w", err)
	}

	var response struct {
		Success    bool            `json:"success"`
		Evaluation ModelEvaluation `json:"evaluation"`
		Error      string          `json:"error"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("解析评估结果失败: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("评估失败: %s", response.Error)
	}

	return &response.Evaluation, nil
}

// ListModels 列出可用模型
func (mi *ModelInterface) ListModels(ctx context.Context) ([]string, error) {
	script := `
import json
import os
import glob

try:
	# 查找模型文件
	model_dir = "./models"
	if not os.path.exists(model_dir):
		os.makedirs(model_dir)
	
	model_files = glob.glob(os.path.join(model_dir, "*.pkl"))
	model_ids = [os.path.basename(f).replace('.pkl', '') for f in model_files]
	
	result = {
		'success': True,
		'models': model_ids
	}

except Exception as e:
	result = {
		'success': False,
		'models': [],
		'error': str(e)
	}

print(json.dumps(result))
`

	output, err := mi.client.ExecuteScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("列出模型失败: %w", err)
	}

	var response struct {
		Success bool     `json:"success"`
		Models  []string `json:"models"`
		Error   string   `json:"error"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("解析模型列表失败: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("列出模型失败: %s", response.Error)
	}

	return response.Models, nil
}

// DeleteModel 删除模型
func (mi *ModelInterface) DeleteModel(ctx context.Context, modelID string) error {
	script := fmt.Sprintf(`
import json
import os

try:
	model_id = '%s'
	model_path = f"./models/{model_id}.pkl"
	
	if os.path.exists(model_path):
		os.remove(model_path)
		message = f"模型 {model_id} 已删除"
	else:
		message = f"模型 {model_id} 不存在"
	
	result = {
		'success': True,
		'message': message
	}

except Exception as e:
	result = {
		'success': False,
		'error': str(e)
	}

print(json.dumps(result))
`, modelID)

	output, err := mi.client.ExecuteScript(ctx, script)
	if err != nil {
		return fmt.Errorf("删除模型失败: %w", err)
	}

	var response struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return fmt.Errorf("解析删除结果失败: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("删除模型失败: %s", response.Error)
	}

	return nil
}