package services

import (
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

// FileService 文件服务
type FileService struct {
	db        *gorm.DB
	uploadDir string
	maxSize   int64 // 最大文件大小（字节）
}

// FileInfo 文件信息
type FileInfo struct {
	ID           uint      `json:"id"`
	OriginalName string    `json:"original_name"`
	StoredName   string    `json:"stored_name"`
	FilePath     string    `json:"file_path"`
	FileSize     int64     `json:"file_size"`
	ContentType  string    `json:"content_type"`
	FileHash     string    `json:"file_hash"`
	UploadedBy   uint      `json:"uploaded_by"`
	UploadedAt   time.Time `json:"uploaded_at"`
	Category     string    `json:"category"`
	Description  string    `json:"description"`
	IsPublic     bool      `json:"is_public"`
	DownloadURL  string    `json:"download_url"`
}

// FileRecord 文件记录模型
type FileRecord struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	OriginalName string    `json:"original_name" gorm:"size:255;not null"`
	StoredName   string    `json:"stored_name" gorm:"size:255;not null"`
	FilePath     string    `json:"file_path" gorm:"size:500;not null"`
	FileSize     int64     `json:"file_size" gorm:"not null"`
	ContentType  string    `json:"content_type" gorm:"size:100"`
	FileHash     string    `json:"file_hash" gorm:"size:64;index"`
	UploadedBy   uint      `json:"uploaded_by" gorm:"not null"`
	Category     string    `json:"category" gorm:"size:50"`
	Description  string    `json:"description" gorm:"size:500"`
	IsPublic     bool      `json:"is_public" gorm:"default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (FileRecord) TableName() string {
	return "file_records"
}

// NewFileService 创建新的文件服务实例
func NewFileService(db *gorm.DB, uploadDir string, maxSize int64) *FileService {
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	if maxSize == 0 {
		maxSize = 100 * 1024 * 1024 // 默认100MB
	}
	
	// 确保上传目录存在
	os.MkdirAll(uploadDir, 0755)
	
	return &FileService{
		db:        db,
		uploadDir: uploadDir,
		maxSize:   maxSize,
	}
}

// UploadFile 上传文件
func (s *FileService) UploadFile(fileHeader *multipart.FileHeader, userID uint, category, description string, isPublic bool) (*FileInfo, error) {
	// 验证文件大小
	if fileHeader.Size > s.maxSize {
		return nil, fmt.Errorf("文件大小超过限制 %d MB", s.maxSize/(1024*1024))
	}
	
	// 验证文件类型
	if err := s.validateFileType(fileHeader.Header.Get("Content-Type")); err != nil {
		return nil, err
	}
	
	// 打开上传的文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %v", err)
	}
	defer file.Close()
	
	// 计算文件哈希
	fileHash, err := s.calculateFileHash(file)
	if err != nil {
		return nil, fmt.Errorf("计算文件哈希失败: %v", err)
	}
	
	// 检查文件是否已存在
	existingFile, err := s.findFileByHash(fileHash, userID)
	if err == nil {
		// 文件已存在，返回现有文件信息
		return s.fileRecordToInfo(existingFile), nil
	}
	
	// 重置文件指针
	file.Seek(0, 0)
	
	// 生成存储文件名
	storedName := s.generateStoredName(fileHeader.Filename)
	
	// 根据分类创建子目录
	categoryDir := filepath.Join(s.uploadDir, category)
	if err := os.MkdirAll(categoryDir, 0755); err != nil {
		return nil, fmt.Errorf("创建分类目录失败: %v", err)
	}
	
	// 完整的文件路径
	filePath := filepath.Join(categoryDir, storedName)
	
	// 保存文件
	if err := s.saveFile(file, filePath); err != nil {
		return nil, fmt.Errorf("保存文件失败: %v", err)
	}
	
	// 创建文件记录
	fileRecord := &FileRecord{
		OriginalName: fileHeader.Filename,
		StoredName:   storedName,
		FilePath:     filePath,
		FileSize:     fileHeader.Size,
		ContentType:  fileHeader.Header.Get("Content-Type"),
		FileHash:     fileHash,
		UploadedBy:   userID,
		Category:     category,
		Description:  description,
		IsPublic:     isPublic,
	}
	
	if err := s.db.Create(fileRecord).Error; err != nil {
		// 删除已保存的文件
		os.Remove(filePath)
		return nil, fmt.Errorf("创建文件记录失败: %v", err)
	}
	
	return s.fileRecordToInfo(fileRecord), nil
}

// DownloadFile 下载文件
func (s *FileService) DownloadFile(fileID uint, userID uint) (*FileDownloadInfo, error) {
	var fileRecord FileRecord
	query := s.db.Where("id = ?", fileID)
	
	// 权限检查：只能下载自己的文件或公开文件
	query = query.Where("uploaded_by = ? OR is_public = ?", userID, true)
	
	if err := query.First(&fileRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("文件不存在或无权限访问")
		}
		return nil, fmt.Errorf("获取文件记录失败: %v", err)
	}
	
	// 检查文件是否存在
	if _, err := os.Stat(fileRecord.FilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在于存储系统中")
	}
	
	return &FileDownloadInfo{
		FilePath:     fileRecord.FilePath,
		OriginalName: fileRecord.OriginalName,
		ContentType:  fileRecord.ContentType,
		FileSize:     fileRecord.FileSize,
	}, nil
}

// GetFiles 获取文件列表
func (s *FileService) GetFiles(userID uint, category string, isPublic *bool, page, pageSize int) (*PaginatedFiles, error) {
	var files []FileRecord
	var total int64
	
	query := s.db.Model(&FileRecord{})
	
	// 权限过滤
	if isPublic != nil && *isPublic {
		query = query.Where("is_public = ?", true)
	} else {
		query = query.Where("uploaded_by = ? OR is_public = ?", userID, true)
	}
	
	// 分类过滤
	if category != "" {
		query = query.Where("category = ?", category)
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取文件总数失败: %v", err)
	}
	
	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&files).Error; err != nil {
		return nil, fmt.Errorf("获取文件列表失败: %v", err)
	}
	
	// 转换为FileInfo
	fileInfos := make([]FileInfo, len(files))
	for i, file := range files {
		fileInfos[i] = *s.fileRecordToInfo(&file)
	}
	
	return &PaginatedFiles{
		Data:       fileInfos,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	}, nil
}

// DeleteFile 删除文件
func (s *FileService) DeleteFile(fileID uint, userID uint) error {
	var fileRecord FileRecord
	if err := s.db.Where("id = ? AND uploaded_by = ?", fileID, userID).First(&fileRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("文件不存在或无权限删除")
		}
		return fmt.Errorf("获取文件记录失败: %v", err)
	}
	
	// 软删除文件记录
	if err := s.db.Delete(&fileRecord).Error; err != nil {
		return fmt.Errorf("删除文件记录失败: %v", err)
	}
	
	// 删除物理文件
	if err := os.Remove(fileRecord.FilePath); err != nil && !os.IsNotExist(err) {
		// 记录警告但不阻止操作
		fmt.Printf("Warning: 删除物理文件失败: %v\n", err)
	}
	
	return nil
}

// UpdateFileInfo 更新文件信息
func (s *FileService) UpdateFileInfo(fileID uint, userID uint, req UpdateFileInfoRequest) (*FileInfo, error) {
	var fileRecord FileRecord
	if err := s.db.Where("id = ? AND uploaded_by = ?", fileID, userID).First(&fileRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("文件不存在或无权限修改")
		}
		return nil, fmt.Errorf("获取文件记录失败: %v", err)
	}
	
	// 更新字段
	updates := map[string]interface{}{}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}
	
	if len(updates) > 0 {
		if err := s.db.Model(&fileRecord).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("更新文件信息失败: %v", err)
		}
	}
	
	return s.fileRecordToInfo(&fileRecord), nil
}

// GetFileCategories 获取文件分类列表
func (s *FileService) GetFileCategories(userID uint) ([]FileCategory, error) {
	var categories []string
	if err := s.db.Model(&FileRecord{}).
		Select("DISTINCT category").
		Where("uploaded_by = ? OR is_public = ?", userID, true).
		Where("category != ''").
		Pluck("category", &categories).Error; err != nil {
		return nil, fmt.Errorf("获取文件分类失败: %v", err)
	}
	
	result := make([]FileCategory, 0, len(categories))
	for _, category := range categories {
		// 获取该分类下的文件数量
		var count int64
		s.db.Model(&FileRecord{}).
			Where("category = ?", category).
			Where("uploaded_by = ? OR is_public = ?", userID, true).
			Count(&count)
		
		result = append(result, FileCategory{
			Name:  category,
			Count: int(count),
		})
	}
	
	return result, nil
}

// GetStorageStats 获取存储统计信息
func (s *FileService) GetStorageStats(userID uint) (*StorageStats, error) {
	var stats StorageStats
	
	// 获取用户文件统计
	if err := s.db.Model(&FileRecord{}).
		Select("COUNT(*) as file_count, COALESCE(SUM(file_size), 0) as total_size").
		Where("uploaded_by = ?", userID).
		Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("获取存储统计失败: %v", err)
	}
	
	// 按分类统计
	type CategoryStat struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
		Size     int64  `json:"size"`
	}
	
	var categoryStats []CategoryStat
	if err := s.db.Model(&FileRecord{}).
		Select("category, COUNT(*) as count, COALESCE(SUM(file_size), 0) as size").
		Where("uploaded_by = ?", userID).
		Group("category").
		Scan(&categoryStats).Error; err != nil {
		return nil, fmt.Errorf("获取分类统计失败: %v", err)
	}
	
	stats.CategoryStats = make(map[string]interface{})
	for _, stat := range categoryStats {
		stats.CategoryStats[stat.Category] = map[string]interface{}{
			"count": stat.Count,
			"size":  stat.Size,
		}
	}
	
	return &stats, nil
}

// 内部方法

// validateFileType 验证文件类型
func (s *FileService) validateFileType(contentType string) error {
	allowedTypes := map[string]bool{
		"text/csv":                                true,
		"application/json":                        true,
		"application/vnd.ms-excel":               true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
		"text/plain":                             true,
		"application/pdf":                        true,
		"image/jpeg":                             true,
		"image/png":                              true,
		"image/gif":                              true,
		"application/zip":                        true,
		"application/x-zip-compressed":           true,
		"application/octet-stream":               true, // 允许二进制文件
	}
	
	if !allowedTypes[contentType] {
		return fmt.Errorf("不支持的文件类型: %s", contentType)
	}
	
	return nil
}

// calculateFileHash 计算文件哈希
func (s *FileService) calculateFileHash(file multipart.File) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// findFileByHash 根据哈希查找文件
func (s *FileService) findFileByHash(hash string, userID uint) (*FileRecord, error) {
	var fileRecord FileRecord
	if err := s.db.Where("file_hash = ? AND uploaded_by = ?", hash, userID).First(&fileRecord).Error; err != nil {
		return nil, err
	}
	return &fileRecord, nil
}

// generateStoredName 生成存储文件名
func (s *FileService) generateStoredName(originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%d_%s%s", timestamp, generateRandomString(8), ext)
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// saveFile 保存文件
func (s *FileService) saveFile(src multipart.File, dst string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	
	_, err = io.Copy(out, src)
	return err
}

// fileRecordToInfo 转换文件记录到文件信息
func (s *FileService) fileRecordToInfo(record *FileRecord) *FileInfo {
	return &FileInfo{
		ID:           record.ID,
		OriginalName: record.OriginalName,
		StoredName:   record.StoredName,
		FilePath:     record.FilePath,
		FileSize:     record.FileSize,
		ContentType:  record.ContentType,
		FileHash:     record.FileHash,
		UploadedBy:   record.UploadedBy,
		UploadedAt:   record.CreatedAt,
		Category:     record.Category,
		Description:  record.Description,
		IsPublic:     record.IsPublic,
		DownloadURL:  fmt.Sprintf("/api/files/%d/download", record.ID),
	}
}

// 请求和响应结构体

type UpdateFileInfoRequest struct {
	Description string `json:"description"`
	Category    string `json:"category"`
	IsPublic    *bool  `json:"is_public"`
}

type FileDownloadInfo struct {
	FilePath     string `json:"file_path"`
	OriginalName string `json:"original_name"`
	ContentType  string `json:"content_type"`
	FileSize     int64  `json:"file_size"`
}

type PaginatedFiles struct {
	Data       []FileInfo `json:"data"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int64      `json:"total_pages"`
}

type FileCategory struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type StorageStats struct {
	FileCount     int64                  `json:"file_count"`
	TotalSize     int64                  `json:"total_size"`
	CategoryStats map[string]interface{} `json:"category_stats"`
}