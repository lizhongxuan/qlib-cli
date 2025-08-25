package qlib

import (
	"context"
	"fmt"
	"testing"
	"time"

	"qlib-backend/internal/testutils"
)

func TestFactorCalculator(t *testing.T) {
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	// 创建模拟的Qlib客户端和因子计算器
	mockClient := testutils.NewMockQlibClient()
	calculator := NewFactorCalculator(mockClient)

	t.Run("CalculatorCreation", func(t *testing.T) {
		if calculator == nil {
			t.Error("Factor calculator should not be nil")
		}

		if calculator.client != mockClient {
			t.Error("Calculator should use the provided client")
		}
	})

	t.Run("FactorExpressionValidation", func(t *testing.T) {
		ctx := context.Background()

		// 测试不同的因子表达式
		validExpressions := []string{
			"$close",
			"$close / Ref($close, 1) - 1",
			"($high + $low + $close) / 3",
			"Mean($close, 5)",
			"Std($close, 20)",
			"($close - Mean($close, 20)) / Std($close, 20)",
		}

		for _, expr := range validExpressions {
			t.Run(fmt.Sprintf("ValidExpression_%s", expr[:10]), func(t *testing.T) {
				isValid, err := calculator.ValidateFactorExpression(ctx, expr)
				if err != nil {
					t.Errorf("Validation failed for expression '%s': %v", expr, err)
				}

				if !isValid {
					t.Errorf("Expression '%s' should be valid", expr)
				}
			})
		}

		// 测试无效表达式
		invalidExpressions := []string{
			"",
			"invalid syntax",
			"$close +",
			"Ref($close)",  // 缺少参数
			"undefined_function($close)",
		}

		for _, expr := range invalidExpressions {
			t.Run(fmt.Sprintf("InvalidExpression_%s", expr), func(t *testing.T) {
				isValid, err := calculator.ValidateFactorExpression(ctx, expr)
				
				// 空表达式应该返回错误
				if expr == "" && err == nil {
					t.Error("Empty expression should return error")
				}

				// 其他无效表达式应该返回false
				if expr != "" && err == nil && isValid {
					t.Errorf("Expression '%s' should be invalid", expr)
				}
			})
		}
	})

	t.Run("FactorCalculation", func(t *testing.T) {
		ctx := context.Background()

		// 测试因子计算
		testExpressions := []FactorExpression{
			{
				Name:       "momentum_1d",
				Expression: "$close / Ref($close, 1) - 1",
				Universe:   "csi300",
				Frequency:  "day",
				StartDate:  "2022-01-01",
				EndDate:    "2023-12-31",
			},
			{
				Name:       "price_volume_trend",
				Expression: "Corr($close, $volume, 10)",
				Universe:   "csi500",
				Frequency:  "day",
				StartDate:  "2022-01-01",
				EndDate:    "2023-12-31",
			},
			{
				Name:       "volatility_factor",
				Expression: "Std($close, 20) / Mean($close, 20)",
				Universe:   "csi300",
				Frequency:  "day",
				StartDate:  "2022-01-01",
				EndDate:    "2023-12-31",
			},
		}

		for _, factorExpr := range testExpressions {
			t.Run(fmt.Sprintf("Calculate_%s", factorExpr.Name), func(t *testing.T) {
				result, err := calculator.CalculateFactor(ctx, factorExpr)
				if err != nil {
					t.Errorf("Factor calculation failed for '%s': %v", factorExpr.Name, err)
					return
				}

				if result == nil {
					t.Errorf("Factor result should not be nil for '%s'", factorExpr.Name)
					return
				}

				// 验证结果结构
				if !result.Success {
					t.Errorf("Factor calculation should succeed for '%s', error: %s", factorExpr.Name, result.Error)
				}

				if result.FactorName != factorExpr.Name {
					t.Errorf("Result factor name mismatch: expected %s, got %s", factorExpr.Name, result.FactorName)
				}

				// 验证元数据
				if result.Metadata == nil {
					t.Errorf("Factor result should contain metadata for '%s'", factorExpr.Name)
				}

				// 验证统计数据
				if len(result.Stats) == 0 {
					t.Errorf("Factor result should contain statistics for '%s'", factorExpr.Name)
				}
			})
		}
	})

	t.Run("BatchFactorCalculation", func(t *testing.T) {
		ctx := context.Background()

		expressions := []FactorExpression{
			{
				Name:       "factor_1",
				Expression: "$close",
				Universe:   "csi300",
				Frequency:  "day",
				StartDate:  "2023-01-01",
				EndDate:    "2023-03-31",
			},
			{
				Name:       "factor_2",
				Expression: "$open",
				Universe:   "csi300",
				Frequency:  "day",
				StartDate:  "2023-01-01",
				EndDate:    "2023-03-31",
			},
			{
				Name:       "factor_3",
				Expression: "$volume",
				Universe:   "csi300",
				Frequency:  "day",
				StartDate:  "2023-01-01",
				EndDate:    "2023-03-31",
			},
		}

		results, err := calculator.BatchCalculateFactors(ctx, expressions)
		if err != nil {
			t.Errorf("Batch factor calculation failed: %v", err)
		}

		if len(results) != len(expressions) {
			t.Errorf("Expected %d results, got %d", len(expressions), len(results))
		}

		// 验证每个结果
		for i, result := range results {
			expectedName := expressions[i].Name
			if result.FactorName != expectedName {
				t.Errorf("Result %d: expected factor name %s, got %s", i, expectedName, result.FactorName)
			}
		}
	})

	t.Run("GetBuiltinFactors", func(t *testing.T) {
		ctx := context.Background()

		factors, err := calculator.GetBuiltinFactors(ctx)
		if err != nil {
			t.Errorf("Failed to get builtin factors: %v", err)
		}

		if factors == nil {
			t.Error("Builtin factors should not be nil")
		}

		// 验证包含基本分类
		expectedCategories := []string{"price", "technical", "momentum", "volatility", "volume"}
		for _, category := range expectedCategories {
			if factorList, exists := factors[category]; !exists {
				t.Errorf("Missing category: %s", category)
			} else if len(factorList) == 0 {
				t.Errorf("Category %s should contain factors", category)
			}
		}
	})
}

func TestFactorPerformanceAnalysis(t *testing.T) {
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	mockClient := testutils.NewMockQlibClient()
	calculator := NewFactorCalculator(mockClient)

	t.Run("FactorPerformanceCalculation", func(t *testing.T) {
		ctx := context.Background()

		// 创建模拟的因子数据和收益数据
		factorData := []FactorValue{
			{Instrument: "000001.SZ", Date: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), Value: 0.02, IsValid: true},
			{Instrument: "000002.SZ", Date: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), Value: -0.01, IsValid: true},
			{Instrument: "000001.SZ", Date: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC), Value: 0.015, IsValid: true},
			{Instrument: "000002.SZ", Date: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC), Value: -0.005, IsValid: true},
		}

		returnData := []FactorValue{
			{Instrument: "000001.SZ", Date: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), Value: 0.025, IsValid: true},
			{Instrument: "000002.SZ", Date: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), Value: -0.012, IsValid: true},
			{Instrument: "000001.SZ", Date: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC), Value: 0.018, IsValid: true},
			{Instrument: "000002.SZ", Date: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC), Value: -0.008, IsValid: true},
		}

		performance, err := calculator.CalculateFactorPerformance(ctx, "test_factor", factorData, returnData)
		if err != nil {
			t.Errorf("Factor performance calculation failed: %v", err)
		}

		if performance == nil {
			t.Error("Performance result should not be nil")
		}

		// 验证性能指标
		if performance.IC == nil {
			t.Error("Performance should contain IC analysis")
		}

		if performance.RankIC == nil {
			t.Error("Performance should contain Rank IC analysis")
		}

		if performance.ICIR == 0 {
			t.Error("ICIR should be calculated")
		}
	})

	t.Run("FactorCorrelationAnalysis", func(t *testing.T) {
		ctx := context.Background()

		// 创建两个相关的因子数据
		factor1 := []FactorValue{
			{Instrument: "000001.SZ", Date: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), Value: 0.02, IsValid: true},
			{Instrument: "000002.SZ", Date: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), Value: -0.01, IsValid: true},
		}

		factor2 := []FactorValue{
			{Instrument: "000001.SZ", Date: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), Value: 0.022, IsValid: true},
			{Instrument: "000002.SZ", Date: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), Value: -0.008, IsValid: true},
		}

		correlation, err := calculator.GetFactorCorrelation(ctx, factor1, factor2)
		if err != nil {
			t.Errorf("Factor correlation calculation failed: %v", err)
		}

		// 相关性应该在-1到1之间
		if correlation < -1 || correlation > 1 {
			t.Errorf("Correlation should be between -1 and 1, got %f", correlation)
		}

		// 对于正相关的数据，相关性应该为正
		if correlation < 0 {
			t.Errorf("Expected positive correlation for similar data, got %f", correlation)
		}
	})
}

func TestFactorCalculatorEdgeCases(t *testing.T) {
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	t.Run("UninitializedClient", func(t *testing.T) {
		// 测试未初始化的客户端
		uninitializedClient := &testutils.MockQlibClient{
			Initialized: false,
		}
		calculator := NewFactorCalculator(uninitializedClient)

		ctx := context.Background()
		expr := FactorExpression{
			Name:       "test_factor",
			Expression: "$close",
			Universe:   "csi300",
			Frequency:  "day",
			StartDate:  "2023-01-01",
			EndDate:    "2023-01-31",
		}

		_, err := calculator.CalculateFactor(ctx, expr)
		if err == nil {
			t.Error("Should return error when client is not initialized")
		}
	})

	t.Run("InvalidDateRange", func(t *testing.T) {
		mockClient := testutils.NewMockQlibClient()
		calculator := NewFactorCalculator(mockClient)

		ctx := context.Background()

		// 测试无效的日期范围
		invalidExpressions := []FactorExpression{
			{
				Name:       "invalid_date_factor",
				Expression: "$close",
				Universe:   "csi300",
				Frequency:  "day",
				StartDate:  "2023-12-31",  // 结束日期早于开始日期
				EndDate:    "2023-01-01",
			},
			{
				Name:       "invalid_format_factor",
				Expression: "$close",
				Universe:   "csi300",
				Frequency:  "day",
				StartDate:  "invalid-date",
				EndDate:    "2023-01-01",
			},
		}

		for _, expr := range invalidExpressions {
			t.Run(fmt.Sprintf("InvalidDate_%s", expr.Name), func(t *testing.T) {
				// 在实际实现中，这些应该返回错误
				// 这里我们验证表达式的基本结构
				if expr.StartDate == "" || expr.EndDate == "" {
					t.Error("Start date and end date should not be empty")
				}

				if expr.StartDate == "2023-12-31" && expr.EndDate == "2023-01-01" {
					t.Error("End date should not be earlier than start date")
				}

				if expr.StartDate == "invalid-date" {
					t.Error("Date format should be valid")
				}
			})
		}
	})

	t.Run("EmptyFactorData", func(t *testing.T) {
		mockClient := testutils.NewMockQlibClient()
		calculator := NewFactorCalculator(mockClient)

		ctx := context.Background()

		// 测试空因子数据的相关性计算
		emptyFactor1 := []FactorValue{}
		emptyFactor2 := []FactorValue{}

		correlation, err := calculator.GetFactorCorrelation(ctx, emptyFactor1, emptyFactor2)
		if err != nil {
			t.Errorf("Empty factor correlation calculation failed: %v", err)
		}

		// 空数据的相关性应该为0
		if correlation != 0 {
			t.Errorf("Expected correlation 0 for empty data, got %f", correlation)
		}
	})

	t.Run("SingleDataPoint", func(t *testing.T) {
		mockClient := testutils.NewMockQlibClient()
		calculator := NewFactorCalculator(mockClient)

		ctx := context.Background()

		// 测试单个数据点
		singleFactor1 := []FactorValue{
			{Instrument: "000001.SZ", Date: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), Value: 0.02, IsValid: true},
		}

		singleFactor2 := []FactorValue{
			{Instrument: "000001.SZ", Date: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), Value: 0.022, IsValid: true},
		}

		correlation, err := calculator.GetFactorCorrelation(ctx, singleFactor1, singleFactor2)
		if err != nil {
			t.Errorf("Single point correlation calculation failed: %v", err)
		}

		// 单个数据点无法计算相关性，应该返回0或特殊值
		if correlation != 0 {
			t.Logf("Single point correlation: %f", correlation)
		}
	})
}

func TestFactorCalculatorPerformance(t *testing.T) {
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	mockClient := testutils.NewMockQlibClient()
	calculator := NewFactorCalculator(mockClient)

	t.Run("LargeDatasetCalculation", func(t *testing.T) {
		ctx := context.Background()

		// 创建大量因子数据进行性能测试
		largeFactorData := make([]FactorValue, 10000)
		for i := range largeFactorData {
			largeFactorData[i] = FactorValue{
				Instrument: fmt.Sprintf("%06d.SZ", i%1000),
				Date:       time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i/1000),
				Value:      float64(i) * 0.001,
				IsValid:    true,
			}
		}

		largeReturnData := make([]FactorValue, 10000)
		for i := range largeReturnData {
			largeReturnData[i] = FactorValue{
				Instrument: fmt.Sprintf("%06d.SZ", i%1000),
				Date:       time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i/1000),
				Value:      float64(i) * 0.0012,
				IsValid:    true,
			}
		}

		start := time.Now()
		_, err := calculator.CalculateFactorPerformance(ctx, "large_factor", largeFactorData, largeReturnData)
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Large dataset performance calculation failed: %v", err)
		}

		maxDuration := 5 * time.Second
		if duration > maxDuration {
			t.Errorf("Performance calculation took too long: %v (max: %v)", duration, maxDuration)
		}
	})

	t.Run("ConcurrentCalculations", func(t *testing.T) {
		ctx := context.Background()
		done := make(chan bool, 5)

		// 并发执行多个因子计算
		for i := 0; i < 5; i++ {
			go func(index int) {
				defer func() { done <- true }()

				expr := FactorExpression{
					Name:       fmt.Sprintf("concurrent_factor_%d", index),
					Expression: "$close",
					Universe:   "csi300",
					Frequency:  "day",
					StartDate:  "2023-01-01",
					EndDate:    "2023-01-31",
				}

				_, err := calculator.CalculateFactor(ctx, expr)
				if err != nil {
					t.Errorf("Concurrent calculation %d failed: %v", index, err)
				}
			}(i)
		}

		// 等待所有计算完成
		for i := 0; i < 5; i++ {
			select {
			case <-done:
				// 成功完成
			case <-time.After(10 * time.Second):
				t.Error("Timeout waiting for concurrent calculations")
				return
			}
		}
	})
}