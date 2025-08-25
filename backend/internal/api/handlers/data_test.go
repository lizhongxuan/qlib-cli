package handlers

import (
	"net/http"
	"testing"

	"qlib-backend/internal/testutils"
)

func TestDataHandlers(t *testing.T) {
	// 设置测试环境
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	// 设置测试路由器
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())

	// 添加路由
	router.GET("/data/datasets", GetDatasets)
	router.POST("/data/datasets", CreateDataset)
	router.PUT("/data/datasets/:id", UpdateDataset)
	router.DELETE("/data/datasets/:id", DeleteDataset)
	router.GET("/data/sources", GetDataSources)
	router.POST("/data/sources/test-connection", TestDataSourceConnection)
	router.GET("/data/explore/:dataset_id", ExploreDataset)
	router.POST("/data/upload", UploadDataFile)

	testCases := []testutils.TestCase{
		{
			Name:           "获取数据集列表-成功",
			Method:         "GET",
			URL:            "/data/datasets",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "创建数据集-成功",
			Method: "POST",
			URL:    "/data/datasets",
			Body: map[string]interface{}{
				"name":        "test_dataset",
				"description": "Test dataset",
				"type":        "stock",
				"format":      "csv",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "创建数据集-缺少必要字段",
			Method: "POST",
			URL:    "/data/datasets",
			Body: map[string]interface{}{
				"description": "Test dataset",
			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:   "更新数据集-成功",
			Method: "PUT",
			URL:    "/data/datasets/1",
			Body: map[string]interface{}{
				"name":        "updated_dataset",
				"description": "Updated description",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "删除数据集-成功",
			Method:         "DELETE",
			URL:            "/data/datasets/1",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "获取数据源列表-成功",
			Method:         "GET",
			URL:            "/data/sources",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "测试数据源连接-成功",
			Method: "POST",
			URL:    "/data/sources/test-connection",
			Body: map[string]interface{}{
				"type": "mysql",
				"host": "localhost",
				"port": 3306,
				"username": "test",
				"password": "test",
				"database": "test_db",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "数据探索-成功",
			Method:         "GET",
			URL:            "/data/explore/1",
			ExpectedStatus: http.StatusOK,
		},
	}

	testutils.RunTestCases(t, router, testCases)
}

func TestGetDatasets(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.GET("/data/datasets", GetDatasets)

	// 测试分页参数
	testCases := []struct {
		name   string
		params string
		status int
	}{
		{"默认分页", "", http.StatusOK},
		{"第一页", "?page=1&limit=10", http.StatusOK},
		{"第二页", "?page=2&limit=5", http.StatusOK},
		{"无效页码", "?page=0", http.StatusOK}, // 应该使用默认值
		{"无效限制", "?limit=-1", http.StatusOK}, // 应该使用默认值
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/data/datasets" + tc.params
			req, _ := testutils.CreateJSONRequest("GET", url, nil)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)
		})
	}
}

func TestCreateDataset(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.POST("/data/datasets", CreateDataset)

	// 测试不同的创建场景
	testCases := []struct {
		name   string
		body   map[string]interface{}
		status int
	}{
		{
			name: "完整的数据集信息",
			body: map[string]interface{}{
				"name":        "complete_dataset",
				"description": "Complete dataset with all fields",
				"type":        "stock",
				"format":      "csv",
				"path":        "/data/complete.csv",
			},
			status: http.StatusOK,
		},
		{
			name: "最小必要信息",
			body: map[string]interface{}{
				"name":   "minimal_dataset",
				"type":   "stock",
				"format": "csv",
			},
			status: http.StatusOK,
		},
		{
			name: "缺少名称",
			body: map[string]interface{}{
				"type":   "stock",
				"format": "csv",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "无效的数据类型",
			body: map[string]interface{}{
				"name":   "invalid_type_dataset",
				"type":   "invalid_type",
				"format": "csv",
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := testutils.CreateJSONRequest("POST", "/data/datasets", tc.body)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)
		})
	}
}

func TestUpdateDataset(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.PUT("/data/datasets/:id", UpdateDataset)

	testCases := []struct {
		name   string
		id     string
		body   map[string]interface{}
		status int
	}{
		{
			name: "更新名称和描述",
			id:   "1",
			body: map[string]interface{}{
				"name":        "updated_name",
				"description": "updated description",
			},
			status: http.StatusOK,
		},
		{
			name: "更新类型",
			id:   "1",
			body: map[string]interface{}{
				"type": "futures",
			},
			status: http.StatusOK,
		},
		{
			name:   "无效ID",
			id:     "invalid",
			body:   map[string]interface{}{},
			status: http.StatusBadRequest,
		},
		{
			name:   "不存在的ID",
			id:     "999999",
			body:   map[string]interface{}{},
			status: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/data/datasets/" + tc.id
			req, _ := testutils.CreateJSONRequest("PUT", url, tc.body)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)
		})
	}
}

func TestDeleteDataset(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.DELETE("/data/datasets/:id", DeleteDataset)

	testCases := []struct {
		name   string
		id     string
		status int
	}{
		{"删除存在的数据集", "1", http.StatusOK},
		{"删除不存在的数据集", "999999", http.StatusNotFound},
		{"无效的ID格式", "invalid", http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/data/datasets/" + tc.id
			req, _ := testutils.CreateJSONRequest("DELETE", url, nil)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)
		})
	}
}

func TestTestDataSourceConnection(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.POST("/data/sources/test-connection", TestDataSourceConnection)

	testCases := []struct {
		name   string
		body   map[string]interface{}
		status int
	}{
		{
			name: "MySQL连接测试",
			body: map[string]interface{}{
				"type":     "mysql",
				"host":     "localhost",
				"port":     3306,
				"username": "test",
				"password": "test",
				"database": "test_db",
			},
			status: http.StatusOK,
		},
		{
			name: "PostgreSQL连接测试",
			body: map[string]interface{}{
				"type":     "postgresql",
				"host":     "localhost",
				"port":     5432,
				"username": "test",
				"password": "test",
				"database": "test_db",
			},
			status: http.StatusOK,
		},
		{
			name: "缺少必要参数",
			body: map[string]interface{}{
				"type": "mysql",
				"host": "localhost",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "不支持的数据库类型",
			body: map[string]interface{}{
				"type":     "oracle",
				"host":     "localhost",
				"port":     1521,
				"username": "test",
				"password": "test",
				"database": "test_db",
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := testutils.CreateJSONRequest("POST", "/data/sources/test-connection", tc.body)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)
		})
	}
}

func TestExploreDataset(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.GET("/data/explore/:dataset_id", ExploreDataset)

	testCases := []struct {
		name   string
		id     string
		params string
		status int
	}{
		{"基本数据探索", "1", "", http.StatusOK},
		{"带限制的探索", "1", "?limit=100", http.StatusOK},
		{"带偏移的探索", "1", "?offset=10&limit=50", http.StatusOK},
		{"无效的数据集ID", "invalid", "", http.StatusBadRequest},
		{"不存在的数据集", "999999", "", http.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/data/explore/" + tc.id + tc.params
			req, _ := testutils.CreateJSONRequest("GET", url, nil)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)
		})
	}
}