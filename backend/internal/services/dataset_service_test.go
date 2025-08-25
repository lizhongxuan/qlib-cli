package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"qlib-backend/internal/models"
	"qlib-backend/internal/testutils"
)

type DatasetServiceTestSuite struct {
	suite.Suite
	service *DatasetService
	testDB  *testutils.TestDB
}

func (suite *DatasetServiceTestSuite) SetupSuite() {
	suite.testDB = testutils.SetupTestDB()
	suite.service = NewDatasetService(suite.testDB.DB)
}

func (suite *DatasetServiceTestSuite) TearDownSuite() {
	suite.testDB.Cleanup()
}

func (suite *DatasetServiceTestSuite) SetupTest() {
	suite.testDB.CleanupTables()
}

func (suite *DatasetServiceTestSuite) TestCreateDataset() {
	req := DatasetCreateRequest{
		Name:        "测试数据集",
		Description: "用于测试的数据集",
		DataPath:    "/data/test_dataset.csv",
		Market:      "CSI300",
		StartDate:   "2020-01-01",
		EndDate:     "2023-12-31",
		FileSize:    1024000,
		RecordCount: 10000,
	}

	dataset, err := suite.service.CreateDataset(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), dataset)
	assert.Equal(suite.T(), req.Name, dataset.Name)
	assert.Equal(suite.T(), req.Description, dataset.Description)
	assert.Equal(suite.T(), req.DataPath, dataset.DataPath)
	assert.Equal(suite.T(), "active", dataset.Status)
	assert.Equal(suite.T(), req.Market, dataset.Market)
	assert.Greater(suite.T(), dataset.ID, uint(0))
}

func (suite *DatasetServiceTestSuite) TestGetDatasets() {
	// 创建测试数据
	datasets := []models.Dataset{
		{
			Name:        "数据集1",
			Description: "描述1",
			DataPath:    "/data/dataset1.csv",
			Status:      "active",
			Market:      "CSI300",
			StartDate:   "2020-01-01",
			EndDate:     "2023-12-31",
			FileSize:    1000,
			RecordCount: 100,
		},
		{
			Name:        "数据集2",
			Description: "描述2",
			DataPath:    "/data/dataset2.csv",
			Status:      "inactive",
			Market:      "SSE50",
			StartDate:   "2021-01-01",
			EndDate:     "2023-12-31",
			FileSize:    2000,
			RecordCount: 200,
		},
	}

	for i := range datasets {
		suite.testDB.DB.Create(&datasets[i])
	}

	// 测试获取所有数据集
	result, err := suite.service.GetDatasets(1, 10, "", "")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(2), result.Total)
	assert.Len(suite.T(), result.Datasets, 2)

	// 测试按市场筛选
	result, err = suite.service.GetDatasets(1, 10, "CSI300", "")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.Total)
	assert.Len(suite.T(), result.Datasets, 1)
	assert.Equal(suite.T(), "数据集1", result.Datasets[0].Name)

	// 测试按状态筛选
	result, err = suite.service.GetDatasets(1, 10, "", "active")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.Total)
	assert.Len(suite.T(), result.Datasets, 1)
	assert.Equal(suite.T(), "数据集1", result.Datasets[0].Name)
}

func (suite *DatasetServiceTestSuite) TestGetDatasetByID() {
	// 创建测试数据
	dataset := models.Dataset{
		Name:        "测试数据集",
		Description: "测试描述",
		DataPath:    "/data/test.csv",
		Status:      "active",
		Market:      "CSI300",
		StartDate:   "2020-01-01",
		EndDate:     "2023-12-31",
		FileSize:    1000,
		RecordCount: 100,
	}
	suite.testDB.DB.Create(&dataset)

	// 测试获取存在的数据集
	result, err := suite.service.GetDatasetByID(dataset.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), dataset.Name, result.Name)
	assert.Equal(suite.T(), dataset.Description, result.Description)

	// 测试获取不存在的数据集
	result, err = suite.service.GetDatasetByID(999)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

func (suite *DatasetServiceTestSuite) TestUpdateDataset() {
	// 创建测试数据
	dataset := models.Dataset{
		Name:        "原始名称",
		Description: "原始描述",
		DataPath:    "/data/original.csv",
		Status:      "active",
		Market:      "CSI300",
		StartDate:   "2020-01-01",
		EndDate:     "2023-12-31",
		FileSize:    1000,
		RecordCount: 100,
	}
	suite.testDB.DB.Create(&dataset)

	// 更新数据集
	updateReq := DatasetUpdateRequest{
		Name:        "更新名称",
		Description: "更新描述",
		Status:      "inactive",
		Market:      "SSE50",
	}

	updatedDataset, err := suite.service.UpdateDataset(dataset.ID, updateReq)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updatedDataset)
	assert.Equal(suite.T(), updateReq.Name, updatedDataset.Name)
	assert.Equal(suite.T(), updateReq.Description, updatedDataset.Description)
	assert.Equal(suite.T(), updateReq.Status, updatedDataset.Status)
	assert.Equal(suite.T(), updateReq.Market, updatedDataset.Market)

	// 测试更新不存在的数据集
	_, err = suite.service.UpdateDataset(999, updateReq)
	assert.Error(suite.T(), err)
}

func (suite *DatasetServiceTestSuite) TestDeleteDataset() {
	// 创建测试数据
	dataset := models.Dataset{
		Name:        "待删除数据集",
		Description: "测试删除",
		DataPath:    "/data/delete_test.csv",
		Status:      "active",
		Market:      "CSI300",
		StartDate:   "2020-01-01",
		EndDate:     "2023-12-31",
		FileSize:    1000,
		RecordCount: 100,
	}
	suite.testDB.DB.Create(&dataset)

	// 删除数据集
	err := suite.service.DeleteDataset(dataset.ID)
	assert.NoError(suite.T(), err)

	// 验证数据集已被软删除
	var deletedDataset models.Dataset
	err = suite.testDB.DB.Unscoped().First(&deletedDataset, dataset.ID).Error
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), deletedDataset.DeletedAt)

	// 测试删除不存在的数据集
	err = suite.service.DeleteDataset(999)
	assert.Error(suite.T(), err)
}

func (suite *DatasetServiceTestSuite) TestValidateDataset() {
	validReq := DatasetValidationRequest{
		DataPath:    "/data/valid.csv",
		Market:      "CSI300",
		StartDate:   "2020-01-01",
		EndDate:     "2023-12-31",
		Columns:     []string{"date", "symbol", "close", "volume"},
		DateFormat:  "YYYY-MM-DD",
		Delimiter:   ",",
		Encoding:    "UTF-8",
	}

	result, err := suite.service.ValidateDataset(validReq)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.IsValid)
	assert.Empty(suite.T(), result.ValidationErrors)
	assert.Greater(suite.T(), result.EstimatedRecords, 0)

	// 测试无效数据集
	invalidReq := DatasetValidationRequest{
		DataPath:   "/invalid/path.csv",
		Market:     "",
		StartDate:  "invalid-date",
		EndDate:    "2023-12-31",
		Columns:    []string{},
		DateFormat: "YYYY-MM-DD",
	}

	result, err = suite.service.ValidateDataset(invalidReq)
	assert.NoError(suite.T(), err) // 验证函数本身不应该出错
	assert.NotNil(suite.T(), result)
	assert.False(suite.T(), result.IsValid)
	assert.NotEmpty(suite.T(), result.ValidationErrors)
}

func (suite *DatasetServiceTestSuite) TestProcessUploadedFile() {
	req := FileProcessRequest{
		FileName:    "test_data.csv",
		FileSize:    1024,
		TempPath:    "/tmp/test_data.csv",
		UserID:      1,
		DatasetName: "上传测试数据集",
		Market:      "CSI300",
		Description: "通过文件上传创建的数据集",
	}

	result, err := suite.service.ProcessUploadedFile(req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Greater(suite.T(), result.DatasetID, uint(0))
	assert.Equal(suite.T(), "processing", result.Status)
	assert.NotEmpty(suite.T(), result.TaskID)
}

func TestDatasetServiceTestSuite(t *testing.T) {
	suite.Run(t, new(DatasetServiceTestSuite))
}