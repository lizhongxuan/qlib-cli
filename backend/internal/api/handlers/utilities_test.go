package handlers

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"qlib-backend/internal/testutils"
)

// MockFileService 模拟文件服务
type MockFileService struct {
	mock.Mock
}

func (m *MockFileService) UploadFile(header interface{}, userID uint, category, description string, isPublic bool) (interface{}, error) {
	args := m.Called(header, userID, category, description, isPublic)
	return args.Get(0), args.Error(1)
}

func (m *MockFileService) DownloadFile(fileID, userID uint) (interface{}, error) {
	args := m.Called(fileID, userID)
	return args.Get(0), args.Error(1)
}

// MockTaskManager 模拟任务管理器
type MockTaskManager struct {
	mock.Mock
}

func (m *MockTaskManager) GetTasks(userID uint, status, taskType string, page, pageSize int) (interface{}, error) {
	args := m.Called(userID, status, taskType, page, pageSize)
	return args.Get(0), args.Error(1)
}

func (m *MockTaskManager) GetTaskStatus(taskID uint) (interface{}, error) {
	args := m.Called(taskID)
	return args.Get(0), args.Error(1)
}

func (m *MockTaskManager) CancelTask(taskID uint) error {
	args := m.Called(taskID)
	return args.Error(0)
}

// MockDownloadInfo 模拟下载信息结构
type MockDownloadInfo struct {
	OriginalName string
	FilePath     string
	FileSize     int64
}

// MockTaskStatus 模拟任务状态结构
type MockTaskStatus struct {
	Status string
}

func TestUtilitiesHandler_UploadFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         uint
		setupRequest   func() (*http.Request, error)
		setupMock      func(*MockFileService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "成功上传文件",
			userID: 1,
			setupRequest: func() (*http.Request, error) {
				// 创建临时文件
				tmpFile, err := os.CreateTemp("", "test*.txt")
				if err != nil {
					return nil, err
				}
				defer tmpFile.Close()
				tmpFile.WriteString("test content")

				// 创建multipart表单
				var buf bytes.Buffer
				writer := multipart.NewWriter(&buf)
				
				// 添加文件字段
				fileWriter, err := writer.CreateFormFile("file", "test.txt")
				if err != nil {
					return nil, err
				}
				fileWriter.Write([]byte("test content"))
				
				// 添加其他字段
				writer.WriteField("category", "data")
				writer.WriteField("description", "测试文件")
				writer.WriteField("is_public", "true")
				
				writer.Close()

				req, err := http.NewRequest("POST", "/files/upload", &buf)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req, nil
			},
			setupMock: func(m *MockFileService) {
				m.On("UploadFile", mock.Anything, uint(1), "data", "测试文件", true).Return(map[string]interface{}{
					"file_id":    1,
					"filename":   "test.txt",
					"size":       12,
					"upload_url": "/files/1",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "无文件上传",
			userID: 1,
			setupRequest: func() (*http.Request, error) {
				var buf bytes.Buffer
				writer := multipart.NewWriter(&buf)
				writer.WriteField("category", "data")
				writer.Close()

				req, err := http.NewRequest("POST", "/files/upload", &buf)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req, nil
			},
			setupMock:      func(m *MockFileService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "获取上传文件失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置模拟服务
			mockFileService := new(MockFileService)
			mockTaskManager := new(MockTaskManager)
			tt.setupMock(mockFileService)

			// 创建处理器
			handler := NewUtilitiesHandler(mockFileService, mockTaskManager)

			// 创建测试路由
			router := gin.New()
			router.POST("/files/upload", testutils.MockAuthMiddleware(tt.userID), handler.UploadFile)

			// 创建请求
			req, err := tt.setupRequest()
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 验证结果
			assert.Equal(t, tt.expectedStatus, w.Code)

			mockFileService.AssertExpectations(t)
		})
	}
}

func TestUtilitiesHandler_DownloadFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		fileID         string
		userID         uint
		setupMock      func(*MockFileService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "成功下载文件",
			fileID: "1",
			userID: 1,
			setupMock: func(m *MockFileService) {
				m.On("DownloadFile", uint(1), uint(1)).Return(&MockDownloadInfo{
					OriginalName: "test.txt",
					FilePath:     "/tmp/test.txt",
					FileSize:     12,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "无效的文件ID",
			fileID:         "invalid",
			userID:         1,
			setupMock:      func(m *MockFileService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "无效的文件ID",
		},
		{
			name:   "文件不存在",
			fileID: "999",
			userID: 1,
			setupMock: func(m *MockFileService) {
				m.On("DownloadFile", uint(999), uint(1)).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建临时文件用于测试下载
			if tt.expectedStatus == http.StatusOK {
				tmpFile, err := os.CreateTemp("", "test*.txt")
				if err != nil {
					t.Fatal(err)
				}
				defer os.Remove(tmpFile.Name())
				tmpFile.WriteString("test content")
				tmpFile.Close()
			}

			// 设置模拟服务
			mockFileService := new(MockFileService)
			mockTaskManager := new(MockTaskManager)
			tt.setupMock(mockFileService)

			// 创建处理器
			handler := NewUtilitiesHandler(mockFileService, mockTaskManager)

			// 创建测试路由
			router := gin.New()
			router.GET("/files/:file_id/download", testutils.MockAuthMiddleware(tt.userID), handler.DownloadFile)

			// 创建请求
			req, _ := http.NewRequest("GET", "/files/"+tt.fileID+"/download", nil)
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 验证结果
			assert.Equal(t, tt.expectedStatus, w.Code)

			mockFileService.AssertExpectations(t)
		})
	}
}

func TestUtilitiesHandler_GetTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		userID         uint
		setupMock      func(*MockTaskManager)
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "成功获取任务列表",
			queryParams: "?status=running&type=training&page=1&page_size=10",
			userID:      1,
			setupMock: func(m *MockTaskManager) {
				m.On("GetTasks", uint(1), "running", "training", 1, 10).Return([]map[string]interface{}{
					{
						"task_id":    1,
						"name":       "模型训练任务",
						"status":     "running",
						"type":       "training",
						"progress":   50,
						"created_at": "2023-01-01T00:00:00Z",
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "默认参数获取任务",
			queryParams: "",
			userID:      1,
			setupMock: func(m *MockTaskManager) {
				m.On("GetTasks", uint(1), "", "", 1, 20).Return([]map[string]interface{}{
					{
						"task_id": 1,
						"name":    "测试任务",
						"status":  "completed",
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "服务错误",
			queryParams: "",
			userID:      1,
			setupMock: func(m *MockTaskManager) {
				m.On("GetTasks", uint(1), "", "", 1, 20).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置模拟服务
			mockFileService := new(MockFileService)
			mockTaskManager := new(MockTaskManager)
			tt.setupMock(mockTaskManager)

			// 创建处理器
			handler := NewUtilitiesHandler(mockFileService, mockTaskManager)

			// 创建测试路由
			router := gin.New()
			router.GET("/tasks", testutils.MockAuthMiddleware(tt.userID), handler.GetTasks)

			// 创建请求
			req, _ := http.NewRequest("GET", "/tasks"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 验证结果
			assert.Equal(t, tt.expectedStatus, w.Code)

			mockTaskManager.AssertExpectations(t)
		})
	}
}

func TestUtilitiesHandler_CancelTask(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		taskID         string
		userID         uint
		setupMock      func(*MockTaskManager)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "成功取消任务",
			taskID: "1",
			userID: 1,
			setupMock: func(m *MockTaskManager) {
				m.On("GetTaskStatus", uint(1)).Return(&MockTaskStatus{Status: "running"}, nil)
				m.On("CancelTask", uint(1)).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "无效的任务ID",
			taskID:         "invalid",
			userID:         1,
			setupMock:      func(m *MockTaskManager) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "无效的任务ID",
		},
		{
			name:   "任务不存在",
			taskID: "999",
			userID: 1,
			setupMock: func(m *MockTaskManager) {
				m.On("GetTaskStatus", uint(999)).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "取消任务失败",
			taskID: "1",
			userID: 1,
			setupMock: func(m *MockTaskManager) {
				m.On("GetTaskStatus", uint(1)).Return(&MockTaskStatus{Status: "running"}, nil)
				m.On("CancelTask", uint(1)).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置模拟服务
			mockFileService := new(MockFileService)
			mockTaskManager := new(MockTaskManager)
			tt.setupMock(mockTaskManager)

			// 创建处理器
			handler := NewUtilitiesHandler(mockFileService, mockTaskManager)

			// 创建测试路由
			router := gin.New()
			router.POST("/tasks/:task_id/cancel", testutils.MockAuthMiddleware(tt.userID), handler.CancelTask)

			// 创建请求
			req, _ := http.NewRequest("POST", "/tasks/"+tt.taskID+"/cancel", nil)
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 验证结果
			assert.Equal(t, tt.expectedStatus, w.Code)

			mockTaskManager.AssertExpectations(t)
		})
	}
}