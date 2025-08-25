package services

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"qlib-backend/internal/models"

	"gorm.io/gorm"
)

type DatasetService struct {
	db *gorm.DB
}

func NewDatasetService(db *gorm.DB) *DatasetService {
	return &DatasetService{db: db}
}

// CreateDataset 创建新数据集
func (s *DatasetService) CreateDataset(req DatasetCreateRequest) (*models.Dataset, error) {
	dataset := &models.Dataset{
		Name:        req.Name,
		Description: req.Description,
		DataPath:    req.DataPath,
		Status:      "active",
		Market:      req.Market,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		FileSize:    req.FileSize,
		RecordCount: req.RecordCount,
	}

	if err := s.db.Create(dataset).Error; err != nil {
		return nil, fmt.Errorf("创建数据集失败: %v", err)
	}

	return dataset, nil
}

// GetDatasets 获取数据集列表
func (s *DatasetService) GetDatasets(page, pageSize int, market, status string) (*PaginatedDatasets, error) {
	var datasets []models.Dataset
	var total int64

	query := s.db.Model(&models.Dataset{})

	// 添加过滤条件
	if market != "" {
		query = query.Where("market = ?", market)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取数据集总数失败: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&datasets).Error; err != nil {
		return nil, fmt.Errorf("获取数据集列表失败: %v", err)
	}

	return &PaginatedDatasets{
		Data:       datasets,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	}, nil
}

// GetDatasetByID 根据ID获取数据集
func (s *DatasetService) GetDatasetByID(id uint) (*models.Dataset, error) {
	var dataset models.Dataset
	if err := s.db.First(&dataset, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("数据集不存在")
		}
		return nil, fmt.Errorf("获取数据集失败: %v", err)
	}
	return &dataset, nil
}

// UpdateDataset 更新数据集
func (s *DatasetService) UpdateDataset(id uint, req DatasetUpdateRequest) (*models.Dataset, error) {
	var dataset models.Dataset
	if err := s.db.First(&dataset, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("数据集不存在")
		}
		return nil, fmt.Errorf("获取数据集失败: %v", err)
	}

	// 更新字段
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Market != "" {
		updates["market"] = req.Market
	}
	if req.StartDate != "" {
		updates["start_date"] = req.StartDate
	}
	if req.EndDate != "" {
		updates["end_date"] = req.EndDate
	}

	if err := s.db.Model(&dataset).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新数据集失败: %v", err)
	}

	return &dataset, nil
}

// DeleteDataset 删除数据集
func (s *DatasetService) DeleteDataset(id uint) error {
	var dataset models.Dataset
	if err := s.db.First(&dataset, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("数据集不存在")
		}
		return fmt.Errorf("获取数据集失败: %v", err)
	}

	// 软删除
	if err := s.db.Delete(&dataset).Error; err != nil {
		return fmt.Errorf("删除数据集失败: %v", err)
	}

	// 可选：删除相关文件
	if dataset.DataPath != "" {
		if err := os.Remove(dataset.DataPath); err != nil {
			// 记录日志但不阻止删除操作
			fmt.Printf("Warning: 删除数据文件失败: %v\n", err)
		}
	}

	return nil
}

// ExploreDataset 数据探索 - 获取数据集的基本统计信息
func (s *DatasetService) ExploreDataset(id uint, limit int) (*DatasetExploreResult, error) {
	dataset, err := s.GetDatasetByID(id)
	if err != nil {
		return nil, err
	}

	// TODO: 实际的数据探索逻辑，这里返回模拟数据
	result := &DatasetExploreResult{
		DatasetID:   dataset.ID,
		Name:        dataset.Name,
		RecordCount: dataset.RecordCount,
		Columns: []DataColumnInfo{
			{Name: "date", Type: "datetime", Description: "交易日期"},
			{Name: "instrument", Type: "string", Description: "股票代码"},
			{Name: "open", Type: "float", Description: "开盘价"},
			{Name: "high", Type: "float", Description: "最高价"},
			{Name: "low", Type: "float", Description: "最低价"},
			{Name: "close", Type: "float", Description: "收盘价"},
			{Name: "volume", Type: "int", Description: "成交量"},
			{Name: "factor", Type: "float", Description: "复权因子"},
		},
		SampleData: []map[string]interface{}{
			{
				"date":       "2023-01-03",
				"instrument": "000001.XSHE",
				"open":       13.50,
				"high":       13.80,
				"low":        13.40,
				"close":      13.75,
				"volume":     1000000,
				"factor":     1.0,
			},
			{
				"date":       "2023-01-04",
				"instrument": "000001.XSHE",
				"open":       13.75,
				"high":       14.00,
				"low":        13.65,
				"close":      13.90,
				"volume":     1200000,
				"factor":     1.0,
			},
		},
		Statistics: map[string]interface{}{
			"total_instruments": 300,
			"date_range":       fmt.Sprintf("%s 到 %s", dataset.StartDate, dataset.EndDate),
			"missing_data_rate": 0.02,
		},
	}

	return result, nil
}

// UploadDataset 上传数据文件
func (s *DatasetService) UploadDataset(fileHeader *multipart.FileHeader, req DatasetUploadRequest) (*models.Dataset, error) {
	// 创建上传目录
	uploadDir := "uploads/datasets"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("创建上传目录失败: %v", err)
	}

	// 生成文件名
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), fileHeader.Filename)
	filepath := filepath.Join(uploadDir, filename)

	// 保存文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %v", err)
	}
	defer file.Close()

	dst, err := os.Create(filepath)
	if err != nil {
		return nil, fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := dst.ReadFrom(file); err != nil {
		return nil, fmt.Errorf("保存文件失败: %v", err)
	}

	// 创建数据集记录
	dataset := &models.Dataset{
		Name:        req.Name,
		Description: req.Description,
		DataPath:    filepath,
		Status:      "processing",
		Market:      req.Market,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		FileSize:    fileHeader.Size,
		RecordCount: 0, // 将在后台处理时更新
	}

	if err := s.db.Create(dataset).Error; err != nil {
		// 删除已上传的文件
		os.Remove(filepath)
		return nil, fmt.Errorf("创建数据集记录失败: %v", err)
	}

	// TODO: 启动后台任务处理数据文件，更新记录数等信息

	return dataset, nil
}

// 请求和响应结构体
type DatasetCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	DataPath    string `json:"data_path" binding:"required"`
	Market      string `json:"market"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	FileSize    int64  `json:"file_size"`
	RecordCount int64  `json:"record_count"`
}

type DatasetUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Market      string `json:"market"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

type DatasetUploadRequest struct {
	Name        string `form:"name" binding:"required"`
	Description string `form:"description"`
	Market      string `form:"market"`
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
}

type PaginatedDatasets struct {
	Data       []models.Dataset `json:"data"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int64            `json:"total_pages"`
}

type DataColumnInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type DatasetExploreResult struct {
	DatasetID   uint                     `json:"dataset_id"`
	Name        string                   `json:"name"`
	RecordCount int64                    `json:"record_count"`
	Columns     []DataColumnInfo         `json:"columns"`
	SampleData  []map[string]interface{} `json:"sample_data"`
	Statistics  map[string]interface{}   `json:"statistics"`
}