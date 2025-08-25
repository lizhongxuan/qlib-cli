package qlib

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockQlibClient 模拟Qlib客户端
type MockQlibClient struct {
	mock.Mock
}

func (m *MockQlibClient) ExecuteScript(ctx context.Context, script string) ([]byte, error) {
	args := m.Called(ctx, script)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockQlibClient) Initialize() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockQlibClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

type DataLoaderTestSuite struct {
	suite.Suite
	loader     *DataLoader
	mockClient *MockQlibClient
}

func (suite *DataLoaderTestSuite) SetupSuite() {
	suite.mockClient = new(MockQlibClient)
	suite.loader = NewDataLoader(suite.mockClient)
}

func (suite *DataLoaderTestSuite) SetupTest() {
	suite.mockClient.ExpectedCalls = nil
}

func (suite *DataLoaderTestSuite) TestLoadStockData() {
	ctx := context.Background()
	
	req := DataRequest{
		Instruments: []string{"000001.SZ", "000002.SZ"},
		StartTime:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		Fields:      []string{"close", "volume", "high", "low"},
		Frequency:   "day",
	}

	// 模拟返回数据
	mockResponse := `{
		"success": true,
		"data": [
			{
				"instrument": "000001.SZ",
				"date": "2023-01-03",
				"features": {
					"close": 11.50,
					"volume": 1000000,
					"high": 11.80,
					"low": 11.20
				}
			},
			{
				"instrument": "000002.SZ",
				"date": "2023-01-03",
				"features": {
					"close": 25.30,
					"volume": 800000,
					"high": 25.50,
					"low": 24.90
				}
			}
		],
		"count": 2,
		"error": ""
	}`

	suite.mockClient.On("ExecuteScript", ctx, mock.MatchedBy(func(script string) bool {
		return len(script) > 0 // 简单验证脚本不为空
	})).Return([]byte(mockResponse), nil)

	response, err := suite.loader.LoadStockData(ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.True(suite.T(), response.Success)
	assert.Len(suite.T(), response.Data, 2)
	assert.Equal(suite.T(), 2, response.Count)
	assert.Empty(suite.T(), response.Error)

	// 验证数据结构
	firstStock := response.Data[0]
	assert.Equal(suite.T(), "000001.SZ", firstStock.Instrument)
	assert.Equal(suite.T(), "2023-01-03", firstStock.Date)
	assert.Equal(suite.T(), 11.50, firstStock.Features["close"])
	assert.Equal(suite.T(), float64(1000000), firstStock.Features["volume"])

	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *DataLoaderTestSuite) TestLoadMarketData() {
	ctx := context.Background()
	
	instrument := "000300.SH" // 沪深300指数
	startDate := "2023-01-01"
	endDate := "2023-01-31"

	// 模拟返回的市场数据
	mockResponse := `{
		"success": true,
		"data": [
			{
				"date": "2023-01-03",
				"open": 3900.50,
				"high": 3920.80,
				"low": 3885.20,
				"close": 3910.75,
				"volume": 50000000,
				"change": 0.25
			},
			{
				"date": "2023-01-04",
				"open": 3910.75,
				"high": 3935.60,
				"low": 3905.30,
				"close": 3925.40,
				"volume": 55000000,
				"change": 0.37
			}
		],
		"count": 2,
		"error": ""
	}`

	suite.mockClient.On("ExecuteScript", ctx, mock.MatchedBy(func(script string) bool {
		return len(script) > 0
	})).Return([]byte(mockResponse), nil)

	marketData, err := suite.loader.LoadMarketData(ctx, instrument, startDate, endDate)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), marketData)
	assert.Len(suite.T(), marketData, 2)

	// 验证第一天数据
	firstDay := marketData[0]
	assert.Equal(suite.T(), "2023-01-03", firstDay.Date)
	assert.Equal(suite.T(), 3900.50, firstDay.Open)
	assert.Equal(suite.T(), 3920.80, firstDay.High)
	assert.Equal(suite.T(), 3885.20, firstDay.Low)
	assert.Equal(suite.T(), 3910.75, firstDay.Close)
	assert.Equal(suite.T(), int64(50000000), firstDay.Volume)
	assert.Equal(suite.T(), 0.25, firstDay.Change)

	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *DataLoaderTestSuite) TestLoadFactorData() {
	ctx := context.Background()
	
	instruments := []string{"000001.SZ", "000002.SZ"}
	factors := []string{"PE", "PB", "ROE"}
	startDate := "2023-01-01"
	endDate := "2023-01-31"

	// 模拟返回的因子数据
	mockResponse := `{
		"success": true,
		"data": [
			{
				"instrument": "000001.SZ",
				"date": "2023-01-03",
				"features": {
					"PE": 15.5,
					"PB": 1.2,
					"ROE": 0.12
				}
			},
			{
				"instrument": "000002.SZ",
				"date": "2023-01-03",
				"features": {
					"PE": 22.3,
					"PB": 2.1,
					"ROE": 0.18
				}
			}
		],
		"count": 2,
		"error": ""
	}`

	suite.mockClient.On("ExecuteScript", ctx, mock.MatchedBy(func(script string) bool {
		return len(script) > 0
	})).Return([]byte(mockResponse), nil)

	factorData, err := suite.loader.LoadFactorData(ctx, instruments, factors, startDate, endDate)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), factorData)
	assert.Len(suite.T(), factorData.Data, 2)
	assert.True(suite.T(), factorData.Success)
	assert.Equal(suite.T(), 2, factorData.Count)

	// 验证因子数据
	firstStock := factorData.Data[0]
	assert.Equal(suite.T(), "000001.SZ", firstStock.Instrument)
	assert.Equal(suite.T(), 15.5, firstStock.Features["PE"])
	assert.Equal(suite.T(), 1.2, firstStock.Features["PB"])
	assert.Equal(suite.T(), 0.12, firstStock.Features["ROE"])

	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *DataLoaderTestSuite) TestGetDataInfo() {
	ctx := context.Background()
	
	// 模拟返回的数据信息
	mockResponse := `{
		"success": true,
		"data": {
			"datasets": ["qlib_data", "custom_data"],
			"instruments_count": 4500,
			"date_range": {
				"start": "2005-01-01",
				"end": "2023-12-31"
			},
			"available_fields": [
				"open", "high", "low", "close", "volume", 
				"PE", "PB", "ROE", "momentum", "volatility"
			],
			"frequencies": ["day", "minute"],
			"last_update": "2023-12-31T23:59:59Z"
		},
		"error": ""
	}`

	suite.mockClient.On("ExecuteScript", ctx, mock.MatchedBy(func(script string) bool {
		return len(script) > 0
	})).Return([]byte(mockResponse), nil)

	dataInfo, err := suite.loader.GetDataInfo(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), dataInfo)
	assert.True(suite.T(), dataInfo.Success)
	assert.NotNil(suite.T(), dataInfo.Data)

	// 验证数据信息结构
	data := dataInfo.Data.(map[string]interface{})
	assert.Contains(suite.T(), data, "datasets")
	assert.Contains(suite.T(), data, "instruments_count")
	assert.Contains(suite.T(), data, "date_range")
	assert.Contains(suite.T(), data, "available_fields")

	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *DataLoaderTestSuite) TestLoadInstrumentList() {
	ctx := context.Background()
	
	market := "CSI300"

	// 模拟返回的股票列表
	mockResponse := `{
		"success": true,
		"data": [
			{
				"instrument": "000001.SZ",
				"name": "平安银行",
				"sector": "金融",
				"market_cap": 500000000000
			},
			{
				"instrument": "000002.SZ", 
				"name": "万科A",
				"sector": "房地产",
				"market_cap": 300000000000
			}
		],
		"count": 2,
		"error": ""
	}`

	suite.mockClient.On("ExecuteScript", ctx, mock.MatchedBy(func(script string) bool {
		return len(script) > 0
	})).Return([]byte(mockResponse), nil)

	instrumentList, err := suite.loader.LoadInstrumentList(ctx, market)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), instrumentList)
	assert.True(suite.T(), instrumentList.Success)
	assert.Len(suite.T(), instrumentList.Data, 2)

	// 验证股票信息
	firstStock := instrumentList.Data[0].(map[string]interface{})
	assert.Equal(suite.T(), "000001.SZ", firstStock["instrument"])
	assert.Equal(suite.T(), "平安银行", firstStock["name"])
	assert.Equal(suite.T(), "金融", firstStock["sector"])

	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *DataLoaderTestSuite) TestValidateDataRequest() {
	// 测试有效请求
	validReq := DataRequest{
		Instruments: []string{"000001.SZ", "000002.SZ"},
		StartTime:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		Fields:      []string{"close", "volume"},
		Frequency:   "day",
	}

	err := suite.loader.ValidateDataRequest(validReq)
	assert.NoError(suite.T(), err)

	// 测试无效请求 - 空股票列表
	invalidReq1 := DataRequest{
		Instruments: []string{},
		StartTime:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		Fields:      []string{"close"},
		Frequency:   "day",
	}

	err = suite.loader.ValidateDataRequest(invalidReq1)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "股票代码列表不能为空")

	// 测试无效请求 - 时间范围错误
	invalidReq2 := DataRequest{
		Instruments: []string{"000001.SZ"},
		StartTime:   time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), // 结束时间早于开始时间
		Fields:      []string{"close"},
		Frequency:   "day",
	}

	err = suite.loader.ValidateDataRequest(invalidReq2)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "结束时间不能早于开始时间")

	// 测试无效请求 - 空字段列表
	invalidReq3 := DataRequest{
		Instruments: []string{"000001.SZ"},
		StartTime:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		Fields:      []string{},
		Frequency:   "day",
	}

	err = suite.loader.ValidateDataRequest(invalidReq3)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "字段列表不能为空")
}

func (suite *DataLoaderTestSuite) TestErrorHandling() {
	ctx := context.Background()
	
	req := DataRequest{
		Instruments: []string{"INVALID.SZ"},
		StartTime:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		Fields:      []string{"close"},
		Frequency:   "day",
	}

	// 模拟返回错误
	mockResponse := `{
		"success": false,
		"data": [],
		"count": 0,
		"error": "Invalid instrument code: INVALID.SZ"
	}`

	suite.mockClient.On("ExecuteScript", ctx, mock.MatchedBy(func(script string) bool {
		return len(script) > 0
	})).Return([]byte(mockResponse), nil)

	response, err := suite.loader.LoadStockData(ctx, req)

	assert.NoError(suite.T(), err) // 网络调用本身成功
	assert.NotNil(suite.T(), response)
	assert.False(suite.T(), response.Success)
	assert.Empty(suite.T(), response.Data)
	assert.Equal(suite.T(), 0, response.Count)
	assert.Contains(suite.T(), response.Error, "Invalid instrument code")

	suite.mockClient.AssertExpectations(suite.T())
}

func TestDataLoaderTestSuite(t *testing.T) {
	suite.Run(t, new(DataLoaderTestSuite))
}