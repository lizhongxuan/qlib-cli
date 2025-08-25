package qlib

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// QlibClient 封装了对Qlib Python库的调用
type QlibClient struct {
	pythonPath   string
	scriptDir    string
	dataProvider string
	region       string
	initialized  bool
}

// QlibConfig Qlib配置结构
type QlibConfig struct {
	Provider   string `json:"provider"`   // 数据提供商：yahoo, baostock等
	Region     string `json:"region"`     // 地区：cn, us等
	DataDir    string `json:"data_dir"`   // 数据目录路径
	Mount      bool   `json:"mount"`      // 是否挂载数据
	ExpName    string `json:"exp_name"`   // 实验名称
	RedisHost  string `json:"redis_host"` // Redis主机地址
	RedisPort  int    `json:"redis_port"` // Redis端口
	MongoHost  string `json:"mongo_host"` // MongoDB主机地址
	MongoPort  int    `json:"mongo_port"` // MongoDB端口
}

// InitResponse Qlib初始化响应
type InitResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewQlibClient 创建新的Qlib客户端
func NewQlibClient() *QlibClient {
	pythonPath := os.Getenv("QLIB_PYTHON_PATH")
	if pythonPath == "" {
		pythonPath = "python3"
	}

	scriptDir := os.Getenv("QLIB_SCRIPT_DIR")
	if scriptDir == "" {
		scriptDir = "./scripts/qlib"
	}

	return &QlibClient{
		pythonPath:   pythonPath,
		scriptDir:    scriptDir,
		dataProvider: "yahoo",
		region:       "us",
		initialized:  false,
	}
}

// Initialize 初始化Qlib环境
func (c *QlibClient) Initialize(ctx context.Context, config QlibConfig) error {
	log.Printf("正在初始化Qlib环境...")

	// 确保脚本目录存在
	if err := os.MkdirAll(c.scriptDir, 0755); err != nil {
		return fmt.Errorf("创建脚本目录失败: %w", err)
	}

	// 生成初始化脚本
	initScript := c.generateInitScript(config)
	scriptPath := filepath.Join(c.scriptDir, "init_qlib.py")

	if err := os.WriteFile(scriptPath, []byte(initScript), 0644); err != nil {
		return fmt.Errorf("写入初始化脚本失败: %w", err)
	}

	// 执行初始化脚本
	cmd := exec.CommandContext(ctx, c.pythonPath, scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("执行Qlib初始化失败: %w, 输出: %s", err, string(output))
	}

	// 解析初始化结果
	var result InitResponse
	if err := json.Unmarshal(output, &result); err != nil {
		log.Printf("警告：无法解析初始化响应，但脚本已执行: %s", string(output))
		result.Success = true // 假定成功
	}

	if !result.Success {
		return fmt.Errorf("Qlib初始化失败: %s", result.Message)
	}

	c.dataProvider = config.Provider
	c.region = config.Region
	c.initialized = true

	log.Printf("Qlib环境初始化成功")
	return nil
}

// generateInitScript 生成Qlib初始化脚本
func (c *QlibClient) generateInitScript(config QlibConfig) string {
	return fmt.Sprintf(`
import json
import sys
import os
sys.path.append('%s')

try:
	import qlib
	from qlib.config import REG_CN, REG_US
	
	# 设置数据目录
	data_dir = '%s'
	if not data_dir:
		data_dir = '~/.qlib/qlib_data/%s'
	
	# 设置区域配置
	region = REG_CN if '%s' == 'cn' else REG_US
	
	# 初始化qlib
	qlib.init(
		provider_uri=data_dir,
		region=region,
		dataset_cache=None,
		auto_mount=%s,
		exp_manager={
			'class': 'MLflowExpManager',
			'module_path': 'qlib.workflow.expm',
			'kwargs': {
				'uri': 'file://mlruns',
				'default_exp_name': '%s'
			}
		} if '%s' else None,
		redis_host='%s' if '%s' else None,
		redis_port=%d if '%s' else None,
		mongo_host='%s' if '%s' else None,
		mongo_port=%d if '%s' else None
	)
	
	result = {'success': True, 'message': 'Qlib initialized successfully'}
	
except Exception as e:
	result = {'success': False, 'message': str(e)}

print(json.dumps(result))
`, 
		os.Getenv("PYTHONPATH"),
		config.DataDir, config.Region,
		config.Region,
		fmt.Sprintf("%t", config.Mount),
		config.ExpName, config.ExpName,
		config.RedisHost, config.RedisHost, config.RedisPort, config.RedisHost,
		config.MongoHost, config.MongoHost, config.MongoPort, config.MongoHost)
}

// ExecuteScript 执行Python脚本并返回结果
func (c *QlibClient) ExecuteScript(ctx context.Context, scriptContent string) ([]byte, error) {
	if !c.initialized {
		return nil, fmt.Errorf("Qlib客户端未初始化")
	}

	// 创建临时脚本文件
	timestamp := time.Now().Unix()
	scriptPath := filepath.Join(c.scriptDir, fmt.Sprintf("temp_script_%d.py", timestamp))

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		return nil, fmt.Errorf("写入脚本文件失败: %w", err)
	}
	defer os.Remove(scriptPath)

	// 执行脚本
	cmd := exec.CommandContext(ctx, c.pythonPath, scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("执行Python脚本失败: %w, 输出: %s", err, string(output))
	}

	return output, nil
}

// CallQlibFunction 调用Qlib函数
func (c *QlibClient) CallQlibFunction(ctx context.Context, modulePath, functionName string, params map[string]interface{}) ([]byte, error) {
	paramsJson, _ := json.Marshal(params)
	
	script := fmt.Sprintf(`
import json
import sys
sys.path.append('%s')

try:
	from %s import %s
	
	# 解析参数
	params = json.loads('%s')
	
	# 调用函数
	result = %s(**params)
	
	# 序列化结果
	output = {'success': True, 'data': result}
	
except Exception as e:
	output = {'success': False, 'error': str(e)}

print(json.dumps(output, default=str))
`, 
		os.Getenv("PYTHONPATH"),
		modulePath, functionName, 
		string(paramsJson), 
		functionName)

	return c.ExecuteScript(ctx, script)
}

// IsInitialized 检查客户端是否已初始化
func (c *QlibClient) IsInitialized() bool {
	return c.initialized
}

// GetDataProvider 获取数据提供商
func (c *QlibClient) GetDataProvider() string {
	return c.dataProvider
}

// GetRegion 获取地区配置
func (c *QlibClient) GetRegion() string {
	return c.region
}

// SetPythonPath 设置Python路径
func (c *QlibClient) SetPythonPath(path string) {
	c.pythonPath = path
}

// SetScriptDir 设置脚本目录
func (c *QlibClient) SetScriptDir(dir string) {
	c.scriptDir = dir
}

// Close 关闭客户端资源
func (c *QlibClient) Close() error {
	// 清理临时文件
	if _, err := os.Stat(c.scriptDir); err == nil {
		files, _ := filepath.Glob(filepath.Join(c.scriptDir, "temp_script_*.py"))
		for _, file := range files {
			os.Remove(file)
		}
	}

	c.initialized = false
	return nil
}

// ValidatePythonEnvironment 验证Python环境
func (c *QlibClient) ValidatePythonEnvironment(ctx context.Context) error {
	script := `
import json
import sys

try:
	import qlib
	import pandas as pd
	import numpy as np
	
	version_info = {
		'qlib_version': qlib.__version__,
		'pandas_version': pd.__version__,
		'numpy_version': np.__version__,
		'python_version': sys.version
	}
	
	output = {'success': True, 'versions': version_info}
	
except ImportError as e:
	output = {'success': False, 'error': f'缺少依赖包: {str(e)}'}
except Exception as e:
	output = {'success': False, 'error': str(e)}

print(json.dumps(output))
`

	result, err := c.ExecuteScript(ctx, script)
	if err != nil {
		return fmt.Errorf("验证Python环境失败: %w", err)
	}

	var response struct {
		Success bool                   `json:"success"`
		Error   string                 `json:"error"`
		Versions map[string]string     `json:"versions"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return fmt.Errorf("解析环境验证结果失败: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("Python环境验证失败: %s", response.Error)
	}

	log.Printf("Python环境验证成功: %+v", response.Versions)
	return nil
}

// GetAvailableDatasets 获取可用数据集列表
func (c *QlibClient) GetAvailableDatasets(ctx context.Context) ([]string, error) {
	script := `
import json
import qlib
from qlib import data

try:
	# 获取可用的数据集
	datasets = []
	
	# 这里可以根据实际需求实现获取数据集的逻辑
	# 例如：遍历数据目录，查找可用的股票数据等
	
	output = {'success': True, 'datasets': datasets}
	
except Exception as e:
	output = {'success': False, 'error': str(e)}

print(json.dumps(output))
`

	result, err := c.ExecuteScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("获取数据集列表失败: %w", err)
	}

	var response struct {
		Success  bool     `json:"success"`
		Error    string   `json:"error"`
		Datasets []string `json:"datasets"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("解析数据集列表失败: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("获取数据集失败: %s", response.Error)
	}

	return response.Datasets, nil
}

// ParseQlibOutput 解析Qlib输出结果
func ParseQlibOutput(output []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	
	// 尝试解析JSON输出
	if err := json.Unmarshal(output, &result); err != nil {
		// 如果不是JSON格式，返回原始文本
		return map[string]interface{}{
			"raw_output": string(output),
		}, nil
	}

	return result, nil
}

// GetQlibInfo 获取Qlib系统信息
func (c *QlibClient) GetQlibInfo(ctx context.Context) (map[string]interface{}, error) {
	script := `
import json
import qlib
import os

try:
	info = {
		'qlib_version': qlib.__version__,
		'data_path': qlib.config.C.get('data_path', 'Not set'),
		'provider_uri': qlib.config.C.get('provider_uri', 'Not set'),
		'region': qlib.config.C.get('region', 'Not set'),
		'pid': os.getpid(),
		'working_dir': os.getcwd()
	}
	
	output = {'success': True, 'info': info}
	
except Exception as e:
	output = {'success': False, 'error': str(e)}

print(json.dumps(output))
`

	result, err := c.ExecuteScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("获取Qlib信息失败: %w", err)
	}

	parsed, err := ParseQlibOutput(result)
	if err != nil {
		return nil, fmt.Errorf("解析Qlib信息失败: %w", err)
	}

	return parsed, nil
}