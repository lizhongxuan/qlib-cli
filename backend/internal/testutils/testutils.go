package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TestDB 测试数据库实例
var TestDB *gorm.DB

// SetupTestDB 设置测试数据库
func SetupTestDB() *gorm.DB {
	// 对于测试，我们使用模拟的数据库或跳过数据库操作
	// 在实际实现中，可以使用内存数据库或测试数据库
	TestDB = nil // 暂时设为nil，在service测试中使用mock
	return TestDB
}

// CleanupTestDB 清理测试数据库
func CleanupTestDB() {
	if TestDB != nil {
		sqlDB, _ := TestDB.DB()
		sqlDB.Close()
	}
}

// SetupTestRouter 设置测试路由器
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// CreateJSONRequest 创建JSON请求
func CreateJSONRequest(method, url string, body interface{}) (*http.Request, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// ParseJSONResponse 解析JSON响应
func ParseJSONResponse(w *httptest.ResponseRecorder, v interface{}) error {
	return json.Unmarshal(w.Body.Bytes(), v)
}

// PerformRequest 执行HTTP请求
func PerformRequest(router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// AssertStatusCode 断言状态码
func AssertStatusCode(t *testing.T, expected, actual int) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected status code %d, got %d", expected, actual)
	}
}

// AssertJSONResponse 断言JSON响应
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expected interface{}) {
	t.Helper()
	
	var actual interface{}
	if err := ParseJSONResponse(w, &actual); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	expectedJSON, _ := json.Marshal(expected)
	actualJSON, _ := json.Marshal(actual)

	if !bytes.Equal(expectedJSON, actualJSON) {
		t.Errorf("Expected JSON %s, got %s", expectedJSON, actualJSON)
	}
}

// SetupTestEnv 设置测试环境变量
func SetupTestEnv() {
	os.Setenv("GIN_MODE", "test")
	os.Setenv("DB_TYPE", "sqlite")
	os.Setenv("JWT_SECRET", "test-secret-key")
	os.Setenv("QLIB_PYTHON_PATH", "python3")
}

// CleanupTestEnv 清理测试环境
func CleanupTestEnv() {
	os.Unsetenv("GIN_MODE")
	os.Unsetenv("DB_TYPE")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("QLIB_PYTHON_PATH")
}

// MockAuthMiddleware 模拟认证中间件
func MockAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 为测试设置用户信息
		c.Set("user_id", uint(1))
		c.Set("username", "testuser")
		c.Next()
	}
}

// TestCase 测试用例结构
type TestCase struct {
	Name           string
	Method         string
	URL            string
	Body           interface{}
	ExpectedStatus int
	ExpectedBody   interface{}
	SetupFunc      func()
	CleanupFunc    func()
}

// RunTestCases 运行测试用例
func RunTestCases(t *testing.T, router *gin.Engine, testCases []TestCase) {
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// 执行设置函数
			if tc.SetupFunc != nil {
				tc.SetupFunc()
			}

			// 执行清理函数
			if tc.CleanupFunc != nil {
				defer tc.CleanupFunc()
			}

			// 创建请求
			req, err := CreateJSONRequest(tc.Method, tc.URL, tc.Body)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// 执行请求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 断言状态码
			AssertStatusCode(t, tc.ExpectedStatus, w.Code)

			// 断言响应体
			if tc.ExpectedBody != nil {
				AssertJSONResponse(t, w, tc.ExpectedBody)
			}
		})
	}
}

// MockQlibClient 模拟Qlib客户端
type MockQlibClient struct {
	Initialized bool
	ScriptCalls []string
}

// NewMockQlibClient 创建模拟Qlib客户端
func NewMockQlibClient() *MockQlibClient {
	return &MockQlibClient{
		Initialized: true,
		ScriptCalls: make([]string, 0),
	}
}

// IsInitialized 模拟初始化检查
func (m *MockQlibClient) IsInitialized() bool {
	return m.Initialized
}

// ExecuteScript 模拟脚本执行
func (m *MockQlibClient) ExecuteScript(script string) ([]byte, error) {
	m.ScriptCalls = append(m.ScriptCalls, script)
	
	// 返回模拟的成功响应
	response := map[string]interface{}{
		"success": true,
		"data":    "mock_data",
	}
	
	return json.Marshal(response)
}

// GetLastScriptCall 获取最后一次脚本调用
func (m *MockQlibClient) GetLastScriptCall() string {
	if len(m.ScriptCalls) == 0 {
		return ""
	}
	return m.ScriptCalls[len(m.ScriptCalls)-1]
}

// Reset 重置模拟客户端
func (m *MockQlibClient) Reset() {
	m.Initialized = true
	m.ScriptCalls = make([]string, 0)
}