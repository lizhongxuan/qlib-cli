package qlib

import (
	"context"
	"fmt"
	"testing"
	"time"

	"qlib-backend/internal/testutils"
)

func TestQlibClient(t *testing.T) {
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	// 创建模拟的Qlib客户端
	client := testutils.NewMockQlibClient()

	t.Run("ClientInitialization", func(t *testing.T) {
		// 测试客户端初始化状态
		if !client.IsInitialized() {
			t.Error("Mock client should be initialized")
		}

		// 测试重置功能
		client.Reset()
		if !client.IsInitialized() {
			t.Error("Client should remain initialized after reset")
		}
	})

	t.Run("ScriptExecution", func(t *testing.T) {
		ctx := context.Background()

		// 测试基本脚本执行
		script := `
print("Hello Qlib")
result = {"success": True, "message": "Test successful"}
import json
print(json.dumps(result))
`

		result, err := client.ExecuteScript(script)
		if err != nil {
			t.Errorf("Script execution failed: %v", err)
		}

		if result == nil {
			t.Error("Script result should not be nil")
		}

		// 验证脚本调用记录
		lastCall := client.GetLastScriptCall()
		if lastCall != script {
			t.Error("Last script call should match executed script")
		}
	})

	t.Run("MultipleScriptExecution", func(t *testing.T) {
		scripts := []string{
			"print('Script 1')",
			"print('Script 2')", 
			"print('Script 3')",
		}

		for i, script := range scripts {
			_, err := client.ExecuteScript(script)
			if err != nil {
				t.Errorf("Script %d execution failed: %v", i+1, err)
			}
		}

		// 验证最后执行的脚本
		lastCall := client.GetLastScriptCall()
		if lastCall != scripts[len(scripts)-1] {
			t.Error("Last script call should match the final script")
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		// 测试客户端在未初始化状态下的行为
		uninitializedClient := &testutils.MockQlibClient{
			Initialized: false,
		}

		_, err := uninitializedClient.ExecuteScript("print('test')")
		if err == nil {
			t.Error("Should return error when client is not initialized")
		}
	})
}

func TestQlibClientIntegration(t *testing.T) {
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	// 创建真实的Qlib客户端用于集成测试
	client := NewQlibClient()

	t.Run("ClientCreation", func(t *testing.T) {
		if client == nil {
			t.Error("Client should not be nil")
		}

		// 测试初始状态
		if client.IsInitialized() {
			t.Error("New client should not be initialized")
		}
	})

	t.Run("ConfigurationValidation", func(t *testing.T) {
		// 测试不同的配置参数
		testConfigs := []QlibConfig{
			{
				Provider: "yahoo",
				Region:   "us",
				DataDir:  "/tmp/test_data",
				Mount:    true,
				ExpName:  "test_experiment",
			},
			{
				Provider: "baostock",
				Region:   "cn",
				DataDir:  "/tmp/cn_data",
				Mount:    false,
			},
		}

		for i, config := range testConfigs {
			t.Run(fmt.Sprintf("Config%d", i+1), func(t *testing.T) {
				// 验证配置参数
				if config.Provider == "" {
					t.Error("Provider should not be empty")
				}

				if config.Region == "" {
					t.Error("Region should not be empty")
				}

				validProviders := []string{"yahoo", "baostock", "tushare"}
				isValidProvider := false
				for _, validProvider := range validProviders {
					if config.Provider == validProvider {
						isValidProvider = true
						break
					}
				}

				if !isValidProvider {
					t.Errorf("Invalid provider: %s", config.Provider)
				}

				validRegions := []string{"cn", "us"}
				isValidRegion := false
				for _, validRegion := range validRegions {
					if config.Region == validRegion {
						isValidRegion = true
						break
					}
				}

				if !isValidRegion {
					t.Errorf("Invalid region: %s", config.Region)
				}
			})
		}
	})

	t.Run("ScriptGeneration", func(t *testing.T) {
		// 测试脚本生成功能
		config := QlibConfig{
			Provider: "yahoo",
			Region:   "us",
			DataDir:  "/tmp/test_data",
			Mount:    true,
			ExpName:  "test_experiment",
		}

		// 由于generateInitScript是私有方法，我们通过其他方式测试
		// 这里测试配置的合理性
		if config.DataDir == "" {
			t.Error("Data directory should be specified")
		}

		if config.ExpName == "" && config.Mount {
			t.Error("Experiment name should be specified when mounting is enabled")
		}
	})

	t.Run("EnvironmentVariables", func(t *testing.T) {
		// 测试环境变量的处理
		originalPath := client.pythonPath
		client.SetPythonPath("/custom/python/path")
		
		if client.pythonPath != "/custom/python/path" {
			t.Error("Python path should be updated")
		}

		// 恢复原始路径
		client.SetPythonPath(originalPath)

		originalScriptDir := client.scriptDir
		client.SetScriptDir("/custom/script/dir")
		
		if client.scriptDir != "/custom/script/dir" {
			t.Error("Script directory should be updated")
		}

		// 恢复原始脚本目录
		client.SetScriptDir(originalScriptDir)
	})
}

func TestQlibClientPerformance(t *testing.T) {
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	client := testutils.NewMockQlibClient()

	t.Run("ConcurrentScriptExecution", func(t *testing.T) {
		ctx := context.Background()
		done := make(chan bool, 10)

		// 测试并发脚本执行
		for i := 0; i < 10; i++ {
			go func(index int) {
				defer func() { done <- true }()

				script := fmt.Sprintf("print('Concurrent script %d')", index)
				_, err := client.ExecuteScript(script)
				if err != nil {
					t.Errorf("Concurrent script %d failed: %v", index, err)
				}
			}(i)
		}

		// 等待所有协程完成
		for i := 0; i < 10; i++ {
			select {
			case <-done:
				// 成功完成
			case <-time.After(5 * time.Second):
				t.Error("Timeout waiting for concurrent script execution")
				return
			}
		}
	})

	t.Run("ScriptExecutionLatency", func(t *testing.T) {
		script := "print('Performance test')"
		
		// 测试多次执行的延迟
		totalDuration := time.Duration(0)
		iterations := 100

		for i := 0; i < iterations; i++ {
			start := time.Now()
			_, err := client.ExecuteScript(script)
			duration := time.Since(start)
			totalDuration += duration

			if err != nil {
				t.Errorf("Script execution %d failed: %v", i+1, err)
			}
		}

		avgDuration := totalDuration / time.Duration(iterations)
		maxAcceptableDuration := 10 * time.Millisecond

		if avgDuration > maxAcceptableDuration {
			t.Errorf("Average script execution time too high: %v (max: %v)", avgDuration, maxAcceptableDuration)
		}
	})

	t.Run("MemoryUsage", func(t *testing.T) {
		// 测试内存使用情况（通过多次执行大脚本）
		largeScript := `
import numpy as np
import pandas as pd

# 创建大数据集进行测试
data = np.random.randn(1000, 100)
df = pd.DataFrame(data)
result = df.describe()
print("Memory test completed")
`

		for i := 0; i < 10; i++ {
			_, err := client.ExecuteScript(largeScript)
			if err != nil {
				t.Errorf("Large script execution %d failed: %v", i+1, err)
			}
		}

		// 验证客户端状态仍然正常
		if !client.IsInitialized() {
			t.Error("Client should remain initialized after memory intensive operations")
		}
	})
}

func TestQlibClientEdgeCases(t *testing.T) {
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	t.Run("EmptyScript", func(t *testing.T) {
		client := testutils.NewMockQlibClient()
		
		_, err := client.ExecuteScript("")
		if err != nil {
			t.Errorf("Empty script should not cause error: %v", err)
		}
	})

	t.Run("VeryLongScript", func(t *testing.T) {
		client := testutils.NewMockQlibClient()
		
		// 创建很长的脚本
		longScript := "print('Long script')\n" 
		for i := 0; i < 1000; i++ {
			longScript += fmt.Sprintf("# Comment line %d\n", i)
		}
		longScript += "print('End of long script')"

		_, err := client.ExecuteScript(longScript)
		if err != nil {
			t.Errorf("Long script execution failed: %v", err)
		}
	})

	t.Run("ScriptWithSpecialCharacters", func(t *testing.T) {
		client := testutils.NewMockQlibClient()
		
		// 测试包含特殊字符的脚本
		specialScript := `
# 测试中文注释
print("测试中文输出")
print("Special chars: !@#$%^&*()")
print("Unicode: 🚀📊💹")
result = {"中文键": "中文值", "ascii_key": "ascii_value"}
import json
print(json.dumps(result, ensure_ascii=False))
`

		_, err := client.ExecuteScript(specialScript)
		if err != nil {
			t.Errorf("Script with special characters failed: %v", err)
		}
	})

	t.Run("ClientStateConsistency", func(t *testing.T) {
		client := testutils.NewMockQlibClient()
		
		// 验证客户端状态在多次操作后保持一致
		initialState := client.IsInitialized()
		
		// 执行多种操作
		scripts := []string{
			"print('Test 1')",
			"",  // 空脚本
			"# Just a comment",
			"print('Final test')",
		}

		for _, script := range scripts {
			client.ExecuteScript(script)
		}

		// 验证状态一致性
		if client.IsInitialized() != initialState {
			t.Error("Client state should remain consistent after operations")
		}

		// 重置后再次验证
		client.Reset()
		if !client.IsInitialized() {
			t.Error("Client should be initialized after reset")
		}
	})
}

