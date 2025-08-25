package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileInfo 文件信息
type FileInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	OriginalName string   `json:"original_name"`
	Extension   string    `json:"extension"`
	Size        int64     `json:"size"`
	MimeType    string    `json:"mime_type"`
	Path        string    `json:"path"`
	URL         string    `json:"url"`
	MD5         string    `json:"md5"`
	UploadTime  time.Time `json:"upload_time"`
}

// FileHandler 文件处理器
type FileHandler struct {
	uploadDir   string
	maxSize     int64
	allowedExts []string
	allowedMimes []string
}

// NewFileHandler 创建文件处理器
func NewFileHandler(uploadDir string, maxSize int64) *FileHandler {
	return &FileHandler{
		uploadDir:    uploadDir,
		maxSize:      maxSize,
		allowedExts:  []string{},
		allowedMimes: []string{},
	}
}

// SetAllowedExtensions 设置允许的文件扩展名
func (fh *FileHandler) SetAllowedExtensions(extensions []string) {
	fh.allowedExts = extensions
}

// SetAllowedMimeTypes 设置允许的MIME类型
func (fh *FileHandler) SetAllowedMimeTypes(mimeTypes []string) {
	fh.allowedMimes = mimeTypes
}

// SaveFile 保存上传的文件
func (fh *FileHandler) SaveFile(fileHeader *multipart.FileHeader, subDir string) (*FileInfo, error) {
	// 打开上传的文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 验证文件
	if err := fh.validateFile(fileHeader, file); err != nil {
		return nil, err
	}

	// 创建目标目录
	targetDir := filepath.Join(fh.uploadDir, subDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 生成唯一文件名
	fileInfo := &FileInfo{
		ID:           generateFileID(),
		OriginalName: fileHeader.Filename,
		Extension:    strings.ToLower(filepath.Ext(fileHeader.Filename)),
		Size:         fileHeader.Size,
		MimeType:     fileHeader.Header.Get("Content-Type"),
		UploadTime:   time.Now(),
	}

	// 构建文件路径
	fileName := fmt.Sprintf("%s%s", fileInfo.ID, fileInfo.Extension)
	fileInfo.Name = fileName
	fileInfo.Path = filepath.Join(targetDir, fileName)
	fileInfo.URL = fmt.Sprintf("/files/%s/%s", subDir, fileName)

	// 重置文件指针
	file.Seek(0, 0)

	// 创建目标文件
	destFile, err := os.Create(fileInfo.Path)
	if err != nil {
		return nil, fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer destFile.Close()

	// 复制文件内容并计算MD5
	hasher := md5.New()
	writer := io.MultiWriter(destFile, hasher)
	
	if _, err := io.Copy(writer, file); err != nil {
		os.Remove(fileInfo.Path) // 清理已创建的文件
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	fileInfo.MD5 = hex.EncodeToString(hasher.Sum(nil))

	return fileInfo, nil
}

// validateFile 验证文件
func (fh *FileHandler) validateFile(fileHeader *multipart.FileHeader, file multipart.File) error {
	// 验证文件大小
	if fh.maxSize > 0 && fileHeader.Size > fh.maxSize {
		return fmt.Errorf("文件大小超过限制: %d bytes", fh.maxSize)
	}

	// 验证文件扩展名
	if len(fh.allowedExts) > 0 {
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if !fh.isExtensionAllowed(ext) {
			return fmt.Errorf("不支持的文件格式: %s", ext)
		}
	}

	// 验证MIME类型
	if len(fh.allowedMimes) > 0 {
		mimeType := fileHeader.Header.Get("Content-Type")
		if !fh.isMimeTypeAllowed(mimeType) {
			return fmt.Errorf("不支持的文件类型: %s", mimeType)
		}
	}

	// 验证文件内容（检查文件头）
	if err := fh.validateFileContent(file, fileHeader.Filename); err != nil {
		return err
	}

	return nil
}

// isExtensionAllowed 检查扩展名是否允许
func (fh *FileHandler) isExtensionAllowed(ext string) bool {
	for _, allowedExt := range fh.allowedExts {
		if ext == allowedExt {
			return true
		}
	}
	return false
}

// isMimeTypeAllowed 检查MIME类型是否允许
func (fh *FileHandler) isMimeTypeAllowed(mimeType string) bool {
	for _, allowedMime := range fh.allowedMimes {
		if mimeType == allowedMime {
			return true
		}
	}
	return false
}

// validateFileContent 验证文件内容
func (fh *FileHandler) validateFileContent(file multipart.File, filename string) error {
	// 读取文件头进行验证
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("读取文件头失败: %w", err)
	}

	// 重置文件指针
	file.Seek(0, 0)

	// 检查文件头是否与扩展名匹配
	ext := strings.ToLower(filepath.Ext(filename))
	if err := fh.validateFileHeader(buffer[:n], ext); err != nil {
		return err
	}

	return nil
}

// validateFileHeader 验证文件头
func (fh *FileHandler) validateFileHeader(header []byte, ext string) error {
	// 常见文件类型的魔数
	fileSignatures := map[string][]byte{
		".jpg":  {0xFF, 0xD8, 0xFF},
		".jpeg": {0xFF, 0xD8, 0xFF},
		".png":  {0x89, 0x50, 0x4E, 0x47},
		".gif":  {0x47, 0x49, 0x46, 0x38},
		".pdf":  {0x25, 0x50, 0x44, 0x46},
		".zip":  {0x50, 0x4B, 0x03, 0x04},
		".xlsx": {0x50, 0x4B, 0x03, 0x04},
		".docx": {0x50, 0x4B, 0x03, 0x04},
		".csv":  {}, // CSV没有固定的魔数
		".txt":  {}, // 文本文件没有固定的魔数
		".json": {}, // JSON文件没有固定的魔数
		".yaml": {}, // YAML文件没有固定的魔数
	}

	signature, exists := fileSignatures[ext]
	if !exists {
		return nil // 未知类型，跳过验证
	}

	if len(signature) == 0 {
		return nil // 没有特定魔数的文件类型
	}

	if len(header) < len(signature) {
		return fmt.Errorf("文件头过短，可能不是有效的 %s 文件", ext)
	}

	for i, b := range signature {
		if header[i] != b {
			return fmt.Errorf("文件头不匹配，可能不是有效的 %s 文件", ext)
		}
	}

	return nil
}

// DeleteFile 删除文件
func (fh *FileHandler) DeleteFile(filePath string) error {
	// 确保文件在上传目录内（安全检查）
	absUploadDir, err := filepath.Abs(fh.uploadDir)
	if err != nil {
		return fmt.Errorf("获取上传目录绝对路径失败: %w", err)
	}

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("获取文件绝对路径失败: %w", err)
	}

	if !strings.HasPrefix(absFilePath, absUploadDir) {
		return fmt.Errorf("不允许删除上传目录外的文件")
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}

// GetFileInfo 获取文件信息
func (fh *FileHandler) GetFileInfo(filePath string) (*FileInfo, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 计算文件MD5
	md5Hash, err := fh.calculateFileMD5(filePath)
	if err != nil {
		return nil, fmt.Errorf("计算文件MD5失败: %w", err)
	}

	fileName := stat.Name()
	return &FileInfo{
		Name:        fileName,
		OriginalName: fileName,
		Extension:   strings.ToLower(filepath.Ext(fileName)),
		Size:        stat.Size(),
		Path:        filePath,
		MD5:         md5Hash,
		UploadTime:  stat.ModTime(),
	}, nil
}

// calculateFileMD5 计算文件MD5
func (fh *FileHandler) calculateFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// MoveFile 移动文件
func (fh *FileHandler) MoveFile(srcPath, destPath string) error {
	// 确保源文件和目标文件都在上传目录内
	absUploadDir, err := filepath.Abs(fh.uploadDir)
	if err != nil {
		return fmt.Errorf("获取上传目录绝对路径失败: %w", err)
	}

	absSrcPath, err := filepath.Abs(srcPath)
	if err != nil {
		return fmt.Errorf("获取源文件绝对路径失败: %w", err)
	}

	absDestPath, err := filepath.Abs(destPath)
	if err != nil {
		return fmt.Errorf("获取目标文件绝对路径失败: %w", err)
	}

	if !strings.HasPrefix(absSrcPath, absUploadDir) || !strings.HasPrefix(absDestPath, absUploadDir) {
		return fmt.Errorf("只能在上传目录内移动文件")
	}

	// 确保目标目录存在
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 移动文件
	if err := os.Rename(srcPath, destPath); err != nil {
		return fmt.Errorf("移动文件失败: %w", err)
	}

	return nil
}

// CopyFile 复制文件
func (fh *FileHandler) CopyFile(srcPath, destPath string) error {
	// 打开源文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer srcFile.Close()

	// 确保目标目录存在
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 创建目标文件
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer destFile.Close()

	// 复制文件内容
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}

	return nil
}

// generateFileID 生成文件ID
func generateFileID() string {
	return fmt.Sprintf("%d_%s", time.Now().UnixNano(), generateRandomString(8))
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

// GetImageHandler 获取图片文件处理器
func GetImageHandler(uploadDir string) *FileHandler {
	handler := NewFileHandler(uploadDir, 10*1024*1024) // 10MB
	handler.SetAllowedExtensions([]string{".jpg", ".jpeg", ".png", ".gif", ".webp"})
	handler.SetAllowedMimeTypes([]string{
		"image/jpeg",
		"image/png", 
		"image/gif",
		"image/webp",
	})
	return handler
}

// GetDocumentHandler 获取文档文件处理器
func GetDocumentHandler(uploadDir string) *FileHandler {
	handler := NewFileHandler(uploadDir, 50*1024*1024) // 50MB
	handler.SetAllowedExtensions([]string{
		".pdf", ".doc", ".docx", ".xls", ".xlsx", 
		".ppt", ".pptx", ".txt", ".csv", ".json", ".yaml", ".yml",
	})
	handler.SetAllowedMimeTypes([]string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"text/plain",
		"text/csv",
		"application/json",
		"application/x-yaml",
		"text/yaml",
	})
	return handler
}

// GetDataFileHandler 获取数据文件处理器（用于Qlib数据文件）
func GetDataFileHandler(uploadDir string) *FileHandler {
	handler := NewFileHandler(uploadDir, 500*1024*1024) // 500MB
	handler.SetAllowedExtensions([]string{
		".csv", ".json", ".pkl", ".h5", ".parquet", ".feather", ".txt",
	})
	handler.SetAllowedMimeTypes([]string{
		"text/csv",
		"application/json",
		"application/octet-stream", // 用于.pkl等二进制文件
		"text/plain",
	})
	return handler
}