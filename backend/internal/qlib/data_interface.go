package qlib

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// DataInterface Qlib数据接口封装
type DataInterface struct {
	pythonPath string
	qlibPath   string
}

// NewDataInterface 创建新的数据接口实例
func NewDataInterface(pythonPath, qlibPath string) *DataInterface {
	if pythonPath == "" {
		pythonPath = "python3"
	}
	return &DataInterface{
		pythonPath: pythonPath,
		qlibPath:   qlibPath,
	}
}

// DataLoadRequest 数据加载请求
type DataLoadRequest struct {
	Instruments []string `json:"instruments"`
	Fields      []string `json:"fields"`
	StartTime   string   `json:"start_time"`
	EndTime     string   `json:"end_time"`
	Freq        string   `json:"freq"`      // day, 30min, 5min
	Provider    string   `json:"provider"`  // yahoo, local
}

// DataInfo 数据信息
type DataInfo struct {
	Instruments   []string               `json:"instruments"`
	Fields        []string               `json:"fields"`
	DateRange     []string               `json:"date_range"`
	TotalRecords  int64                  `json:"total_records"`
	MissingRatio  float64                `json:"missing_ratio"`
	Statistics    map[string]interface{} `json:"statistics"`
}

// InstrumentInfo 股票信息
type InstrumentInfo struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Market      string `json:"market"`
	Sector      string `json:"sector"`
	Industry    string `json:"industry"`
	ListDate    string `json:"list_date"`
	DelistDate  string `json:"delist_date,omitempty"`
	Status      string `json:"status"`
}

// LoadData 加载数据
func (d *DataInterface) LoadData(req DataLoadRequest) (map[string]interface{}, error) {
	// 构建Python脚本参数
	scriptArgs := map[string]interface{}{
		"action":      "load_data",
		"instruments": req.Instruments,
		"fields":      req.Fields,
		"start_time":  req.StartTime,
		"end_time":    req.EndTime,
		"freq":        req.Freq,
		"provider":    req.Provider,
	}

	result, err := d.executePythonScript(scriptArgs)
	if err != nil {
		return nil, fmt.Errorf("加载数据失败: %v", err)
	}

	return result, nil
}

// GetDataInfo 获取数据信息
func (d *DataInterface) GetDataInfo(instruments []string, fields []string) (*DataInfo, error) {
	scriptArgs := map[string]interface{}{
		"action":      "get_data_info",
		"instruments": instruments,
		"fields":      fields,
	}

	result, err := d.executePythonScript(scriptArgs)
	if err != nil {
		return nil, fmt.Errorf("获取数据信息失败: %v", err)
	}

	// 解析结果
	dataInfo := &DataInfo{}
	if data, ok := result["data"].(map[string]interface{}); ok {
		if instruments, ok := data["instruments"].([]interface{}); ok {
			for _, inst := range instruments {
				if instStr, ok := inst.(string); ok {
					dataInfo.Instruments = append(dataInfo.Instruments, instStr)
				}
			}
		}
		if fields, ok := data["fields"].([]interface{}); ok {
			for _, field := range fields {
				if fieldStr, ok := field.(string); ok {
					dataInfo.Fields = append(dataInfo.Fields, fieldStr)
				}
			}
		}
		if dateRange, ok := data["date_range"].([]interface{}); ok {
			for _, date := range dateRange {
				if dateStr, ok := date.(string); ok {
					dataInfo.DateRange = append(dataInfo.DateRange, dateStr)
				}
			}
		}
		if totalRecords, ok := data["total_records"].(float64); ok {
			dataInfo.TotalRecords = int64(totalRecords)
		}
		if missingRatio, ok := data["missing_ratio"].(float64); ok {
			dataInfo.MissingRatio = missingRatio
		}
		if statistics, ok := data["statistics"].(map[string]interface{}); ok {
			dataInfo.Statistics = statistics
		}
	}

	return dataInfo, nil
}

// GetInstruments 获取股票列表
func (d *DataInterface) GetInstruments(market string) ([]InstrumentInfo, error) {
	scriptArgs := map[string]interface{}{
		"action": "get_instruments",
		"market": market,
	}

	result, err := d.executePythonScript(scriptArgs)
	if err != nil {
		return nil, fmt.Errorf("获取股票列表失败: %v", err)
	}

	var instruments []InstrumentInfo
	if data, ok := result["data"].([]interface{}); ok {
		for _, item := range data {
			if instMap, ok := item.(map[string]interface{}); ok {
				inst := InstrumentInfo{}
				if symbol, ok := instMap["symbol"].(string); ok {
					inst.Symbol = symbol
				}
				if name, ok := instMap["name"].(string); ok {
					inst.Name = name
				}
				if market, ok := instMap["market"].(string); ok {
					inst.Market = market
				}
				if sector, ok := instMap["sector"].(string); ok {
					inst.Sector = sector
				}
				if industry, ok := instMap["industry"].(string); ok {
					inst.Industry = industry
				}
				if listDate, ok := instMap["list_date"].(string); ok {
					inst.ListDate = listDate
				}
				if delistDate, ok := instMap["delist_date"].(string); ok {
					inst.DelistDate = delistDate
				}
				if status, ok := instMap["status"].(string); ok {
					inst.Status = status
				}
				instruments = append(instruments, inst)
			}
		}
	}

	return instruments, nil
}

// GetMarkets 获取市场列表
func (d *DataInterface) GetMarkets() ([]string, error) {
	scriptArgs := map[string]interface{}{
		"action": "get_markets",
	}

	result, err := d.executePythonScript(scriptArgs)
	if err != nil {
		return nil, fmt.Errorf("获取市场列表失败: %v", err)
	}

	var markets []string
	if data, ok := result["data"].([]interface{}); ok {
		for _, market := range data {
			if marketStr, ok := market.(string); ok {
				markets = append(markets, marketStr)
			}
		}
	}

	return markets, nil
}

// GetFields 获取可用字段列表
func (d *DataInterface) GetFields() ([]string, error) {
	scriptArgs := map[string]interface{}{
		"action": "get_fields",
	}

	result, err := d.executePythonScript(scriptArgs)
	if err != nil {
		return nil, fmt.Errorf("获取字段列表失败: %v", err)
	}

	var fields []string
	if data, ok := result["data"].([]interface{}); ok {
		for _, field := range data {
			if fieldStr, ok := field.(string); ok {
				fields = append(fields, fieldStr)
			}
		}
	}

	return fields, nil
}

// ValidateDataPath 验证数据路径
func (d *DataInterface) ValidateDataPath(dataPath string) (bool, error) {
	scriptArgs := map[string]interface{}{
		"action":    "validate_data_path",
		"data_path": dataPath,
	}

	result, err := d.executePythonScript(scriptArgs)
	if err != nil {
		return false, fmt.Errorf("验证数据路径失败: %v", err)
	}

	if valid, ok := result["valid"].(bool); ok {
		return valid, nil
	}

	return false, fmt.Errorf("无效的验证结果")
}

// InitializeQlib 初始化Qlib
func (d *DataInterface) InitializeQlib(provider string, region string, dataDirs map[string]string) error {
	scriptArgs := map[string]interface{}{
		"action":    "init_qlib",
		"provider":  provider,
		"region":    region,
		"data_dirs": dataDirs,
	}

	_, err := d.executePythonScript(scriptArgs)
	if err != nil {
		return fmt.Errorf("初始化Qlib失败: %v", err)
	}

	return nil
}

// executePythonScript 执行Python脚本
func (d *DataInterface) executePythonScript(args map[string]interface{}) (map[string]interface{}, error) {
	// 将参数序列化为JSON
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("序列化参数失败: %v", err)
	}

	// 构建Python脚本命令
	pythonScript := `
import json
import sys
import qlib
from qlib.data import D
from qlib.data.dataset.handler import DataHandlerLP
import pandas as pd

def main():
    try:
        # 从标准输入读取参数
        args_json = sys.stdin.read()
        args = json.loads(args_json)
        action = args.get('action')
        
        result = {"success": True, "data": None, "error": None}
        
        if action == "load_data":
            # 加载数据逻辑
            instruments = args.get('instruments', [])
            fields = args.get('fields', [])
            start_time = args.get('start_time')
            end_time = args.get('end_time')
            freq = args.get('freq', 'day')
            
            # 这里是模拟数据
            result["data"] = {
                "instruments": instruments,
                "fields": fields,
                "records": 1000,
                "date_range": [start_time, end_time]
            }
            
        elif action == "get_data_info":
            # 获取数据信息
            result["data"] = {
                "instruments": args.get('instruments', []),
                "fields": args.get('fields', []),
                "date_range": ["2020-01-01", "2023-12-31"],
                "total_records": 50000,
                "missing_ratio": 0.02,
                "statistics": {
                    "mean_price": 25.6,
                    "max_volume": 1000000,
                    "trading_days": 1000
                }
            }
            
        elif action == "get_instruments":
            # 获取股票列表
            market = args.get('market', 'csi300')
            result["data"] = [
                {
                    "symbol": "000001.XSHE",
                    "name": "平安银行",
                    "market": "XSHE",
                    "sector": "金融",
                    "industry": "银行",
                    "list_date": "1991-04-03",
                    "status": "L"
                },
                {
                    "symbol": "000002.XSHE",
                    "name": "万科A",
                    "market": "XSHE", 
                    "sector": "房地产",
                    "industry": "房地产开发",
                    "list_date": "1991-01-29",
                    "status": "L"
                }
            ]
            
        elif action == "get_markets":
            # 获取市场列表
            result["data"] = ["csi300", "csi500", "csi800", "all"]
            
        elif action == "get_fields":
            # 获取字段列表
            result["data"] = ["$open", "$high", "$low", "$close", "$volume", "$factor", "$vwap"]
            
        elif action == "validate_data_path":
            # 验证数据路径
            data_path = args.get('data_path')
            result["valid"] = True  # 模拟验证通过
            
        elif action == "init_qlib":
            # 初始化Qlib
            provider = args.get('provider', 'LocalDataProvider')
            region = args.get('region', 'cn')
            # qlib.init(provider=provider, region=region)
            result["data"] = {"initialized": True}
            
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

	// 执行Python命令
	cmd := exec.Command(d.pythonPath, "-c", pythonScript)
	cmd.Stdin = strings.NewReader(string(argsJSON))
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行Python脚本失败: %v", err)
	}

	// 解析输出
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("解析Python输出失败: %v", err)
	}

	// 检查执行结果
	if success, ok := result["success"].(bool); !ok || !success {
		if errorMsg, ok := result["error"].(string); ok {
			return nil, fmt.Errorf("Python脚本执行失败: %s", errorMsg)
		}
		return nil, fmt.Errorf("Python脚本执行失败")
	}

	return result, nil
}