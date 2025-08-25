package qlib

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// DataLoader 数据加载接口
type DataLoader struct {
	client *QlibClient
}

// StockData 股票数据结构
type StockData struct {
	Instrument string             `json:"instrument"`
	Date       string             `json:"date"`
	Features   map[string]float64 `json:"features"`
}

// MarketData 市场数据结构
type MarketData struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
	Change float64 `json:"change"`
}

// DataRequest 数据请求参数
type DataRequest struct {
	Instruments []string  `json:"instruments"` // 股票代码列表
	StartTime   time.Time `json:"start_time"`  // 开始时间
	EndTime     time.Time `json:"end_time"`    // 结束时间
	Fields      []string  `json:"fields"`      // 字段列表
	Frequency   string    `json:"frequency"`   // 频率：day, minute等
}

// DataResponse 数据响应
type DataResponse struct {
	Success bool        `json:"success"`
	Data    []StockData `json:"data"`
	Error   string      `json:"error"`
	Count   int         `json:"count"`
}

// NewDataLoader 创建数据加载器
func NewDataLoader(client *QlibClient) *DataLoader {
	return &DataLoader{
		client: client,
	}
}

// LoadStockData 加载股票数据
func (dl *DataLoader) LoadStockData(ctx context.Context, req DataRequest) (*DataResponse, error) {
	if !dl.client.IsInitialized() {
		return nil, fmt.Errorf("Qlib客户端未初始化")
	}

	log.Printf("正在加载股票数据: %+v", req)

	// 构建数据加载脚本
	script := dl.buildDataLoadScript(req)

	// 执行脚本
	result, err := dl.client.ExecuteScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("加载股票数据失败: %w", err)
	}

	// 解析结果
	var response DataResponse
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("解析数据响应失败: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("数据加载失败: %s", response.Error)
	}

	log.Printf("成功加载 %d 条股票数据", response.Count)
	return &response, nil
}

// buildDataLoadScript 构建数据加载脚本
func (dl *DataLoader) buildDataLoadScript(req DataRequest) string {
	instrumentsJson, _ := json.Marshal(req.Instruments)
	fieldsJson, _ := json.Marshal(req.Fields)

	return fmt.Sprintf(`
import json
import qlib
from qlib import data
import pandas as pd
import numpy as np
from datetime import datetime

try:
	# 解析参数
	instruments = %s
	start_time = '%s'
	end_time = '%s'
	fields = %s
	frequency = '%s'
	
	# 如果没有指定字段，使用默认字段
	if not fields:
		fields = ['$open', '$high', '$low', '$close', '$volume']
	
	# 加载数据
	if frequency == 'minute':
		# 分钟级数据
		df = data.D.features(
			instruments=instruments,
			fields=fields,
			start_time=start_time,
			end_time=end_time,
			freq='1min'
		)
	else:
		# 日级数据
		df = data.D.features(
			instruments=instruments,
			fields=fields,
			start_time=start_time,
			end_time=end_time,
			freq='day'
		)
	
	# 转换数据格式
	data_list = []
	
	if df is not None and not df.empty:
		# 重置索引以便访问instrument和datetime
		df_reset = df.reset_index()
		
		for _, row in df_reset.iterrows():
			instrument = row['instrument'] if 'instrument' in row else str(row.name[0])
			date = row['datetime'] if 'datetime' in row else str(row.name[1])
			
			# 构建特征字典
			features = {}
			for field in fields:
				field_clean = field.replace('$', '').lower()
				if field in row and pd.notna(row[field]):
					features[field_clean] = float(row[field])
				else:
					features[field_clean] = None
			
			data_list.append({
				'instrument': instrument,
				'date': str(date),
				'features': features
			})
	
	result = {
		'success': True,
		'data': data_list,
		'count': len(data_list),
		'error': None
	}

except Exception as e:
	result = {
		'success': False,
		'data': [],
		'count': 0,
		'error': str(e)
	}

print(json.dumps(result, default=str))
`,
		string(instrumentsJson),
		req.StartTime.Format("2006-01-02"),
		req.EndTime.Format("2006-01-02"),
		string(fieldsJson),
		req.Frequency)
}

// GetMarketData 获取市场数据
func (dl *DataLoader) GetMarketData(ctx context.Context, instrument string, startDate, endDate time.Time) ([]MarketData, error) {
	req := DataRequest{
		Instruments: []string{instrument},
		StartTime:   startDate,
		EndTime:     endDate,
		Fields:      []string{"$open", "$high", "$low", "$close", "$volume"},
		Frequency:   "day",
	}

	response, err := dl.LoadStockData(ctx, req)
	if err != nil {
		return nil, err
	}

	// 转换为MarketData格式
	marketData := make([]MarketData, 0, len(response.Data))
	for _, data := range response.Data {
		md := MarketData{
			Date: data.Date,
		}

		if val, ok := data.Features["open"]; ok && val != 0 {
			md.Open = val
		}
		if val, ok := data.Features["high"]; ok && val != 0 {
			md.High = val
		}
		if val, ok := data.Features["low"]; ok && val != 0 {
			md.Low = val
		}
		if val, ok := data.Features["close"]; ok && val != 0 {
			md.Close = val
		}
		if val, ok := data.Features["volume"]; ok {
			md.Volume = int64(val)
		}

		// 计算变化率
		if md.Open > 0 && md.Close > 0 {
			md.Change = (md.Close - md.Open) / md.Open * 100
		}

		marketData = append(marketData, md)
	}

	return marketData, nil
}

// GetInstrumentList 获取可用的股票列表
func (dl *DataLoader) GetInstrumentList(ctx context.Context, market string) ([]string, error) {
	script := fmt.Sprintf(`
import json
import qlib
from qlib import data

try:
	# 获取股票列表
	market = '%s'
	
	if market.lower() == 'cn' or market.lower() == 'csi300':
		# 中国市场 - CSI300成分股
		instruments = data.D.instruments(market='csi300')
	elif market.lower() == 'us' or market.lower() == 'sp500':
		# 美国市场 - S&P500成分股
		instruments = data.D.instruments(market='sp500')
	elif market.lower() == 'all':
		# 获取所有可用股票
		instruments = data.D.instruments(market='all')
	else:
		# 默认获取当前配置的市场
		instruments = data.D.instruments()
	
	# 转换为列表
	if hasattr(instruments, 'tolist'):
		instrument_list = instruments.tolist()
	else:
		instrument_list = list(instruments)
	
	result = {
		'success': True,
		'instruments': instrument_list,
		'count': len(instrument_list)
	}

except Exception as e:
	result = {
		'success': False,
		'instruments': [],
		'count': 0,
		'error': str(e)
	}

print(json.dumps(result))
`, market)

	output, err := dl.client.ExecuteScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("获取股票列表失败: %w", err)
	}

	var response struct {
		Success     bool     `json:"success"`
		Instruments []string `json:"instruments"`
		Count       int      `json:"count"`
		Error       string   `json:"error"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("解析股票列表失败: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("获取股票列表失败: %s", response.Error)
	}

	return response.Instruments, nil
}

// GetDataRange 获取数据的时间范围
func (dl *DataLoader) GetDataRange(ctx context.Context, instrument string) (time.Time, time.Time, error) {
	script := fmt.Sprintf(`
import json
import qlib
from qlib import data
import pandas as pd

try:
	instrument = '%s'
	
	# 获取数据范围
	calendar = data.D.calendar(freq='day')
	
	if calendar is not None and len(calendar) > 0:
		start_date = str(calendar[0])
		end_date = str(calendar[-1])
		
		result = {
			'success': True,
			'start_date': start_date,
			'end_date': end_date
		}
	else:
		result = {
			'success': False,
			'error': '无法获取数据日历'
		}

except Exception as e:
	result = {
		'success': False,
		'error': str(e)
	}

print(json.dumps(result))
`, instrument)

	output, err := dl.client.ExecuteScript(ctx, script)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("获取数据范围失败: %w", err)
	}

	var response struct {
		Success   bool   `json:"success"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		Error     string `json:"error"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("解析数据范围失败: %w", err)
	}

	if !response.Success {
		return time.Time{}, time.Time{}, fmt.Errorf("获取数据范围失败: %s", response.Error)
	}

	startDate, err := time.Parse("2006-01-02", response.StartDate[:10])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("解析开始日期失败: %w", err)
	}

	endDate, err := time.Parse("2006-01-02", response.EndDate[:10])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("解析结束日期失败: %w", err)
	}

	return startDate, endDate, nil
}

// LoadFactorData 加载因子数据
func (dl *DataLoader) LoadFactorData(ctx context.Context, instruments []string, factors []string, startDate, endDate time.Time) (*DataResponse, error) {
	req := DataRequest{
		Instruments: instruments,
		StartTime:   startDate,
		EndTime:     endDate,
		Fields:      factors,
		Frequency:   "day",
	}

	return dl.LoadStockData(ctx, req)
}

// ValidateData 验证数据有效性
func (dl *DataLoader) ValidateData(ctx context.Context, instrument string, date time.Time) (bool, error) {
	req := DataRequest{
		Instruments: []string{instrument},
		StartTime:   date,
		EndTime:     date,
		Fields:      []string{"$close"},
		Frequency:   "day",
	}

	response, err := dl.LoadStockData(ctx, req)
	if err != nil {
		return false, err
	}

	return response.Count > 0, nil
}

// GetDataStats 获取数据统计信息
func (dl *DataLoader) GetDataStats(ctx context.Context, instrument string, field string, startDate, endDate time.Time) (map[string]float64, error) {
	script := fmt.Sprintf(`
import json
import qlib
from qlib import data
import pandas as pd
import numpy as np

try:
	instrument = '%s'
	field = '%s'
	start_time = '%s'
	end_time = '%s'
	
	# 加载数据
	df = data.D.features(
		instruments=[instrument],
		fields=[field],
		start_time=start_time,
		end_time=end_time,
		freq='day'
	)
	
	if df is not None and not df.empty:
		values = df[field].dropna()
		
		if len(values) > 0:
			stats = {
				'count': len(values),
				'mean': float(values.mean()),
				'std': float(values.std()),
				'min': float(values.min()),
				'max': float(values.max()),
				'median': float(values.median()),
				'q25': float(values.quantile(0.25)),
				'q75': float(values.quantile(0.75))
			}
		else:
			stats = {}
		
		result = {
			'success': True,
			'stats': stats
		}
	else:
		result = {
			'success': False,
			'error': '没有找到数据'
		}

except Exception as e:
	result = {
		'success': False,
		'error': str(e)
	}

print(json.dumps(result, default=str))
`,
		instrument, field,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))

	output, err := dl.client.ExecuteScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("获取数据统计失败: %w", err)
	}

	var response struct {
		Success bool              `json:"success"`
		Stats   map[string]float64 `json:"stats"`
		Error   string            `json:"error"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("解析统计结果失败: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("获取统计信息失败: %s", response.Error)
	}

	return response.Stats, nil
}