package utils

import (
	"encoding/json"
)

// ChartData 图表数据结构
type ChartData struct {
	Type   string                 `json:"type"`   // line, bar, pie, heatmap等
	Title  string                 `json:"title"`
	Data   []interface{}          `json:"data"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// ChartGenerator 图表生成器
type ChartGenerator struct {
}

// NewChartGenerator 创建新的图表生成器
func NewChartGenerator() *ChartGenerator {
	return &ChartGenerator{}
}

// GenerateLineChart 生成折线图
func (cg *ChartGenerator) GenerateLineChart(title string, xData []string, yData []float64) *ChartData {
	data := make([]interface{}, len(xData))
	for i, x := range xData {
		data[i] = map[string]interface{}{
			"x": x,
			"y": yData[i],
		}
	}

	return &ChartData{
		Type:  "line",
		Title: title,
		Data:  data,
		Config: map[string]interface{}{
			"xAxis": map[string]string{"type": "category"},
			"yAxis": map[string]string{"type": "value"},
		},
	}
}

// GenerateBarChart 生成柱状图
func (cg *ChartGenerator) GenerateBarChart(title string, categories []string, values []float64) *ChartData {
	data := make([]interface{}, len(categories))
	for i, cat := range categories {
		data[i] = map[string]interface{}{
			"category": cat,
			"value":    values[i],
		}
	}

	return &ChartData{
		Type:  "bar",
		Title: title,
		Data:  data,
	}
}

// GeneratePieChart 生成饼图
func (cg *ChartGenerator) GeneratePieChart(title string, categories []string, values []float64) *ChartData {
	data := make([]interface{}, len(categories))
	for i, cat := range categories {
		data[i] = map[string]interface{}{
			"name":  cat,
			"value": values[i],
		}
	}

	return &ChartData{
		Type:  "pie",
		Title: title,
		Data:  data,
	}
}

// GenerateHeatmapChart 生成热力图
func (cg *ChartGenerator) GenerateHeatmapChart(title string, xCategories, yCategories []string, values [][]float64) *ChartData {
	data := make([]interface{}, 0)
	for i := range yCategories {
		for j := range xCategories {
			if i < len(values) && j < len(values[i]) {
				data = append(data, []interface{}{j, i, values[i][j]})
			}
		}
	}

	return &ChartData{
		Type:  "heatmap",
		Title: title,
		Data:  data,
		Config: map[string]interface{}{
			"xAxis": map[string]interface{}{
				"type": "category",
				"data": xCategories,
			},
			"yAxis": map[string]interface{}{
				"type": "category",
				"data": yCategories,
			},
		},
	}
}

// GenerateScatterChart 生成散点图
func (cg *ChartGenerator) GenerateScatterChart(title string, xData, yData []float64) *ChartData {
	data := make([]interface{}, len(xData))
	for i := range xData {
		data[i] = []float64{xData[i], yData[i]}
	}

	return &ChartData{
		Type:  "scatter",
		Title: title,
		Data:  data,
	}
}

// GenerateCandlestickChart 生成K线图
func (cg *ChartGenerator) GenerateCandlestickChart(title string, ohlcData [][]float64, dates []string) *ChartData {
	data := make([]interface{}, len(ohlcData))
	for i, ohlc := range ohlcData {
		if len(ohlc) >= 4 {
			data[i] = map[string]interface{}{
				"date":  dates[i],
				"open":  ohlc[0],
				"high":  ohlc[1],
				"low":   ohlc[2],
				"close": ohlc[3],
			}
		}
	}

	return &ChartData{
		Type:  "candlestick",
		Title: title,
		Data:  data,
	}
}

// ToJSON 将图表数据转换为JSON
func (cd *ChartData) ToJSON() ([]byte, error) {
	return json.Marshal(cd)
}