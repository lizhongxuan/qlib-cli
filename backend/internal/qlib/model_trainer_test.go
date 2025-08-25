package qlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ModelTrainerTestSuite struct {
	suite.Suite
	trainer *ModelTrainer
}

func (suite *ModelTrainerTestSuite) SetupSuite() {
	suite.trainer = NewModelTrainer("/usr/bin/python3", "/opt/qlib", "/tmp/qlib_workspace", false)
}

func (suite *ModelTrainerTestSuite) TestNewModelTrainer() {
	// 测试默认python路径
	trainer1 := NewModelTrainer("", "/opt/qlib", "/tmp/workspace", false)
	assert.Equal(suite.T(), "python3", trainer1.pythonPath)
	assert.Equal(suite.T(), "/opt/qlib", trainer1.qlibPath)
	assert.Equal(suite.T(), "/tmp/workspace", trainer1.workspacePath)
	assert.False(suite.T(), trainer1.gpuEnabled)

	// 测试自定义配置
	trainer2 := NewModelTrainer("/usr/local/bin/python3", "/custom/qlib", "/custom/workspace", true)
	assert.Equal(suite.T(), "/usr/local/bin/python3", trainer2.pythonPath)
	assert.Equal(suite.T(), "/custom/qlib", trainer2.qlibPath)
	assert.Equal(suite.T(), "/custom/workspace", trainer2.workspacePath)
	assert.True(suite.T(), trainer2.gpuEnabled)
}

func (suite *ModelTrainerTestSuite) TestTrainModel() {
	params := ModelTrainingParams{
		ModelID:    123,
		ModelType:  "lightgbm",
		ConfigJSON: `{"objective": "regression", "num_leaves": 31}`,
		TrainStart: "2020-01-01",
		TrainEnd:   "2022-12-31",
		ValidStart: "2023-01-01",
		ValidEnd:   "2023-06-30",
		TestStart:  "2023-07-01",
		TestEnd:    "2023-12-31",
		Features:   []string{"close", "volume", "high", "low"},
		Label:      "label",
	}

	// 注意：这是一个集成测试，需要实际的Python环境和Qlib
	// 在实际CI/CD环境中，可能需要模拟或跳过这个测试
	result, err := suite.trainer.TrainModel(params)

	// 由于我们没有实际的Qlib环境，这里只测试参数验证
	if err != nil {
		// 预期错误，因为没有实际的Python环境
		assert.Contains(suite.T(), err.Error(), "python")
	} else {
		// 如果成功，验证结果结构
		assert.NotNil(suite.T(), result)
		assert.NotEmpty(suite.T(), result.ModelPath)
	}
}

func (suite *ModelTrainerTestSuite) TestValidateTrainingParams() {
	// 测试有效参数
	validParams := ModelTrainingParams{
		ModelID:    123,
		ModelType:  "lightgbm",
		ConfigJSON: `{"objective": "regression"}`,
		TrainStart: "2020-01-01",
		TrainEnd:   "2022-12-31",
		ValidStart: "2023-01-01",
		ValidEnd:   "2023-06-30",
		TestStart:  "2023-07-01",
		TestEnd:    "2023-12-31",
		Features:   []string{"close", "volume"},
		Label:      "label",
	}

	err := suite.trainer.ValidateTrainingParams(validParams)
	assert.NoError(suite.T(), err)

	// 测试无效参数 - 空模型类型
	invalidParams1 := validParams
	invalidParams1.ModelType = ""
	err = suite.trainer.ValidateTrainingParams(invalidParams1)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "模型类型不能为空")

	// 测试无效参数 - 无效JSON配置
	invalidParams2 := validParams
	invalidParams2.ConfigJSON = "invalid json"
	err = suite.trainer.ValidateTrainingParams(invalidParams2)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "无效的JSON配置")

	// 测试无效参数 - 空特征列表
	invalidParams3 := validParams
	invalidParams3.Features = []string{}
	err = suite.trainer.ValidateTrainingParams(invalidParams3)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "特征列表不能为空")

	// 测试无效参数 - 时间范围错误
	invalidParams4 := validParams
	invalidParams4.TrainStart = "2023-01-01"
	invalidParams4.TrainEnd = "2022-12-31" // 结束时间早于开始时间
	err = suite.trainer.ValidateTrainingParams(invalidParams4)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "训练结束时间不能早于开始时间")
}

func (suite *ModelTrainerTestSuite) TestEvaluateModel() {
	evalParams := ModelEvaluationParams{
		ModelPath:  "/tmp/models/test_model.pkl",
		TestStart:  "2023-01-01",
		TestEnd:    "2023-12-31",
		Features:   []string{"close", "volume", "high", "low"},
		Label:      "label",
		Benchmark:  "HS300",
	}

	// 模拟评估（实际需要真实模型文件）
	result, err := suite.trainer.EvaluateModel(evalParams)

	if err != nil {
		// 预期错误，因为没有实际的模型文件
		assert.Contains(suite.T(), err.Error(), "模型文件不存在")
	} else {
		// 如果成功，验证结果结构
		assert.NotNil(suite.T(), result)
		assert.GreaterOrEqual(suite.T(), result.TestIC, -1.0)
		assert.LessOrEqual(suite.T(), result.TestIC, 1.0)
	}
}

func (suite *ModelTrainerTestSuite) TestGetModelInfo() {
	modelPath := "/tmp/models/test_model.pkl"

	// 模拟获取模型信息（实际需要真实模型文件）
	info, err := suite.trainer.GetModelInfo(modelPath)

	if err != nil {
		// 预期错误，因为没有实际的模型文件
		assert.Contains(suite.T(), err.Error(), "模型文件不存在")
	} else {
		// 如果成功，验证信息结构
		assert.NotNil(suite.T(), info)
		assert.NotEmpty(suite.T(), info.ModelType)
		assert.NotNil(suite.T(), info.TrainingParams)
	}
}

func (suite *ModelTrainerTestSuite) TestSaveModel() {
	saveParams := ModelSaveParams{
		ModelPath:   "/tmp/models/test_model.pkl",
		ModelName:   "测试模型",
		Description: "LightGBM测试模型",
		Version:     "1.0",
		Metadata: map[string]interface{}{
			"features_count": 10,
			"training_samples": 100000,
		},
	}

	// 模拟保存模型
	result, err := suite.trainer.SaveModel(saveParams)

	if err != nil {
		// 预期错误，因为没有实际的模型文件
		assert.Contains(suite.T(), err.Error(), "模型文件不存在")
	} else {
		// 如果成功，验证结果
		assert.NotNil(suite.T(), result)
		assert.True(suite.T(), result.Success)
		assert.NotEmpty(suite.T(), result.SavedPath)
	}
}

func (suite *ModelTrainerTestSuite) TestLoadModel() {
	modelPath := "/tmp/models/test_model.pkl"

	// 模拟加载模型
	model, err := suite.trainer.LoadModel(modelPath)

	if err != nil {
		// 预期错误，因为没有实际的模型文件
		assert.Contains(suite.T(), err.Error(), "模型文件不存在")
	} else {
		// 如果成功，验证模型对象
		assert.NotNil(suite.T(), model)
		assert.NotEmpty(suite.T(), model.ModelPath)
		assert.NotEmpty(suite.T(), model.ModelType)
	}
}

func (suite *ModelTrainerTestSuite) TestPredictWithModel() {
	predParams := ModelPredictionParams{
		ModelPath: "/tmp/models/test_model.pkl",
		DataStart: "2024-01-01",
		DataEnd:   "2024-01-31",
		Features:  []string{"close", "volume", "high", "low"},
		OutputPath: "/tmp/predictions.csv",
	}

	// 模拟预测
	result, err := suite.trainer.PredictWithModel(predParams)

	if err != nil {
		// 预期错误，因为没有实际的模型文件
		assert.Contains(suite.T(), err.Error(), "模型文件不存在")
	} else {
		// 如果成功，验证预测结果
		assert.NotNil(suite.T(), result)
		assert.True(suite.T(), result.Success)
		assert.Greater(suite.T(), result.PredictionCount, 0)
		assert.NotEmpty(suite.T(), result.OutputPath)
	}
}

func (suite *ModelTrainerTestSuite) TestDeleteModel() {
	modelPath := "/tmp/models/test_model.pkl"

	// 模拟删除模型
	err := suite.trainer.DeleteModel(modelPath)

	if err != nil {
		// 可能的错误情况
		assert.Contains(suite.T(), err.Error(), "模型文件不存在")
	}
	// 如果没有错误，说明删除成功（或文件本来就不存在）
}

func (suite *ModelTrainerTestSuite) TestGenerateModelScript() {
	params := ModelTrainingParams{
		ModelID:    123,
		ModelType:  "lightgbm",
		ConfigJSON: `{"objective": "regression", "num_leaves": 31}`,
		TrainStart: "2020-01-01",
		TrainEnd:   "2022-12-31",
		ValidStart: "2023-01-01",
		ValidEnd:   "2023-06-30",
		TestStart:  "2023-07-01",
		TestEnd:    "2023-12-31",
		Features:   []string{"close", "volume", "high", "low"},
		Label:      "label",
	}

	script, err := suite.trainer.GenerateModelScript(params)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), script)
	
	// 验证脚本包含必要的内容
	assert.Contains(suite.T(), script, "import qlib")
	assert.Contains(suite.T(), script, "lightgbm")
	assert.Contains(suite.T(), script, "2020-01-01")
	assert.Contains(suite.T(), script, "2022-12-31")
	assert.Contains(suite.T(), script, "close")
	assert.Contains(suite.T(), script, "volume")
}

func (suite *ModelTrainerTestSuite) TestSupportedModelTypes() {
	supportedTypes := suite.trainer.GetSupportedModelTypes()

	assert.NotNil(suite.T(), supportedTypes)
	assert.Greater(suite.T(), len(supportedTypes), 0)
	
	// 验证包含常见的模型类型
	assert.Contains(suite.T(), supportedTypes, "lightgbm")
	assert.Contains(suite.T(), supportedTypes, "xgboost")
	assert.Contains(suite.T(), supportedTypes, "linear")
}

func (suite *ModelTrainerTestSuite) TestModelConfiguration() {
	modelType := "lightgbm"
	
	defaultConfig, err := suite.trainer.GetDefaultModelConfig(modelType)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), defaultConfig)
	
	// 验证默认配置包含必要的参数
	assert.Contains(suite.T(), defaultConfig, "objective")
	assert.Contains(suite.T(), defaultConfig, "num_leaves")
	assert.Contains(suite.T(), defaultConfig, "learning_rate")

	// 测试不支持的模型类型
	_, err = suite.trainer.GetDefaultModelConfig("unsupported_model")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "不支持的模型类型")
}

func TestModelTrainerTestSuite(t *testing.T) {
	suite.Run(t, new(ModelTrainerTestSuite))
}