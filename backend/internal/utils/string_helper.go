package utils

import (
	"crypto/rand"
	"encoding/base64"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// StringHelper 字符串工具
type StringHelper struct{}

// NewStringHelper 创建字符串工具
func NewStringHelper() *StringHelper {
	return &StringHelper{}
}

// IsEmpty 检查字符串是否为空
func (sh *StringHelper) IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty 检查字符串是否不为空
func (sh *StringHelper) IsNotEmpty(s string) bool {
	return !sh.IsEmpty(s)
}

// Truncate 截断字符串
func (sh *StringHelper) Truncate(s string, length int) string {
	if utf8.RuneCountInString(s) <= length {
		return s
	}
	
	runes := []rune(s)
	if length <= 3 {
		return string(runes[:length])
	}
	
	return string(runes[:length-3]) + "..."
}

// TruncateBytes 按字节截断字符串
func (sh *StringHelper) TruncateBytes(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	
	// 确保不会截断UTF-8字符中间
	for i := maxBytes; i > 0; i-- {
		if utf8.ValidString(s[:i]) {
			return s[:i]
		}
	}
	
	return ""
}

// PadLeft 左边填充字符串
func (sh *StringHelper) PadLeft(s string, length int, pad rune) string {
	if utf8.RuneCountInString(s) >= length {
		return s
	}
	
	padLength := length - utf8.RuneCountInString(s)
	return strings.Repeat(string(pad), padLength) + s
}

// PadRight 右边填充字符串
func (sh *StringHelper) PadRight(s string, length int, pad rune) string {
	if utf8.RuneCountInString(s) >= length {
		return s
	}
	
	padLength := length - utf8.RuneCountInString(s)
	return s + strings.Repeat(string(pad), padLength)
}

// Reverse 反转字符串
func (sh *StringHelper) Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// CamelCase 转换为驼峰命名
func (sh *StringHelper) CamelCase(s string) string {
	words := sh.splitWords(s)
	if len(words) == 0 {
		return ""
	}
	
	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		result += sh.capitalize(words[i])
	}
	
	return result
}

// PascalCase 转换为帕斯卡命名（首字母大写的驼峰）
func (sh *StringHelper) PascalCase(s string) string {
	words := sh.splitWords(s)
	result := ""
	
	for _, word := range words {
		result += sh.capitalize(word)
	}
	
	return result
}

// SnakeCase 转换为蛇形命名
func (sh *StringHelper) SnakeCase(s string) string {
	words := sh.splitWords(s)
	result := make([]string, len(words))
	
	for i, word := range words {
		result[i] = strings.ToLower(word)
	}
	
	return strings.Join(result, "_")
}

// KebabCase 转换为短横线命名
func (sh *StringHelper) KebabCase(s string) string {
	words := sh.splitWords(s)
	result := make([]string, len(words))
	
	for i, word := range words {
		result[i] = strings.ToLower(word)
	}
	
	return strings.Join(result, "-")
}

// splitWords 分割单词
func (sh *StringHelper) splitWords(s string) []string {
	// 移除特殊字符并分割
	reg := regexp.MustCompile(`[^a-zA-Z0-9\u4e00-\u9fff]+`)
	s = reg.ReplaceAllString(s, " ")
	
	// 在大写字母前插入空格（处理驼峰命名）
	reg = regexp.MustCompile(`([a-z])([A-Z])`)
	s = reg.ReplaceAllString(s, "$1 $2")
	
	words := strings.Fields(s)
	result := make([]string, 0, len(words))
	
	for _, word := range words {
		if word != "" {
			result = append(result, word)
		}
	}
	
	return result
}

// capitalize 首字母大写
func (sh *StringHelper) capitalize(s string) string {
	if s == "" {
		return s
	}
	
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	
	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}
	
	return string(runes)
}

// Contains 检查是否包含子字符串（忽略大小写）
func (sh *StringHelper) Contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// ContainsAny 检查是否包含任意一个子字符串
func (sh *StringHelper) ContainsAny(s string, substrs []string) bool {
	s = strings.ToLower(s)
	for _, substr := range substrs {
		if strings.Contains(s, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

// ContainsAll 检查是否包含所有子字符串
func (sh *StringHelper) ContainsAll(s string, substrs []string) bool {
	s = strings.ToLower(s)
	for _, substr := range substrs {
		if !strings.Contains(s, strings.ToLower(substr)) {
			return false
		}
	}
	return true
}

// StartsWith 检查是否以指定字符串开头（忽略大小写）
func (sh *StringHelper) StartsWith(s, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(s), strings.ToLower(prefix))
}

// EndsWith 检查是否以指定字符串结尾（忽略大小写）
func (sh *StringHelper) EndsWith(s, suffix string) bool {
	return strings.HasSuffix(strings.ToLower(s), strings.ToLower(suffix))
}

// RemovePrefix 移除前缀
func (sh *StringHelper) RemovePrefix(s, prefix string) string {
	if sh.StartsWith(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

// RemoveSuffix 移除后缀
func (sh *StringHelper) RemoveSuffix(s, suffix string) string {
	if sh.EndsWith(s, suffix) {
		return s[:len(s)-len(suffix)]
	}
	return s
}

// RemoveSpaces 移除所有空格
func (sh *StringHelper) RemoveSpaces(s string) string {
	return strings.ReplaceAll(s, " ", "")
}

// RemoveExtraSpaces 移除多余空格
func (sh *StringHelper) RemoveExtraSpaces(s string) string {
	reg := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(reg.ReplaceAllString(s, " "))
}

// SplitAndTrim 分割并修剪空格
func (sh *StringHelper) SplitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result
}

// Join 连接字符串数组
func (sh *StringHelper) Join(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

// JoinNonEmpty 连接非空字符串
func (sh *StringHelper) JoinNonEmpty(strs []string, sep string) string {
	result := make([]string, 0, len(strs))
	
	for _, str := range strs {
		if strings.TrimSpace(str) != "" {
			result = append(result, str)
		}
	}
	
	return strings.Join(result, sep)
}

// Repeat 重复字符串
func (sh *StringHelper) Repeat(s string, count int) string {
	return strings.Repeat(s, count)
}

// Replace 替换字符串
func (sh *StringHelper) Replace(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// ReplaceIgnoreCase 忽略大小写替换
func (sh *StringHelper) ReplaceIgnoreCase(s, old, new string) string {
	reg := regexp.MustCompile("(?i)" + regexp.QuoteMeta(old))
	return reg.ReplaceAllString(s, new)
}

// Count 计算子字符串出现次数
func (sh *StringHelper) Count(s, substr string) int {
	return strings.Count(s, substr)
}

// CountWords 计算单词数
func (sh *StringHelper) CountWords(s string) int {
	words := strings.Fields(s)
	return len(words)
}

// CountChars 计算字符数（Unicode字符）
func (sh *StringHelper) CountChars(s string) int {
	return utf8.RuneCountInString(s)
}

// CountBytes 计算字节数
func (sh *StringHelper) CountBytes(s string) int {
	return len(s)
}

// IsNumeric 检查是否为数字
func (sh *StringHelper) IsNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// IsInteger 检查是否为整数
func (sh *StringHelper) IsInteger(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// IsAlpha 检查是否只包含字母
func (sh *StringHelper) IsAlpha(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return s != ""
}

// IsAlphaNumeric 检查是否只包含字母和数字
func (sh *StringHelper) IsAlphaNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return s != ""
}

// IsLower 检查是否全为小写
func (sh *StringHelper) IsLower(s string) bool {
	return s == strings.ToLower(s) && strings.ContainsFunc(s, unicode.IsLetter)
}

// IsUpper 检查是否全为大写
func (sh *StringHelper) IsUpper(s string) bool {
	return s == strings.ToUpper(s) && strings.ContainsFunc(s, unicode.IsLetter)
}

// ToTitle 转换为标题格式（每个单词首字母大写）
func (sh *StringHelper) ToTitle(s string) string {
	return strings.Title(strings.ToLower(s))
}

// GenerateRandomString 生成随机字符串
func (sh *StringHelper) GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	
	if _, err := rand.Read(b); err != nil {
		// 如果随机数生成失败，使用时间戳作为种子
		for i := range b {
			b[i] = charset[int(b[i])%len(charset)]
		}
	} else {
		for i := range b {
			b[i] = charset[int(b[i])%len(charset)]
		}
	}
	
	return string(b)
}

// GenerateRandomStringWithCharset 使用指定字符集生成随机字符串
func (sh *StringHelper) GenerateRandomStringWithCharset(length int, charset string) string {
	if charset == "" {
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}
	
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	
	return string(b)
}

// Base64Encode Base64编码
func (sh *StringHelper) Base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// Base64Decode Base64解码
func (sh *StringHelper) Base64Decode(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// URLSafeBase64Encode URL安全的Base64编码
func (sh *StringHelper) URLSafeBase64Encode(s string) string {
	return base64.URLEncoding.EncodeToString([]byte(s))
}

// URLSafeBase64Decode URL安全的Base64解码
func (sh *StringHelper) URLSafeBase64Decode(s string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Similarity 计算字符串相似度（Levenshtein距离）
func (sh *StringHelper) Similarity(s1, s2 string) float64 {
	runes1 := []rune(s1)
	runes2 := []rune(s2)
	
	len1 := len(runes1)
	len2 := len(runes2)
	
	if len1 == 0 && len2 == 0 {
		return 1.0
	}
	
	if len1 == 0 || len2 == 0 {
		return 0.0
	}
	
	distance := sh.levenshteinDistance(runes1, runes2)
	maxLen := len1
	if len2 > maxLen {
		maxLen = len2
	}
	
	return 1.0 - float64(distance)/float64(maxLen)
}

// levenshteinDistance 计算Levenshtein距离
func (sh *StringHelper) levenshteinDistance(s1, s2 []rune) int {
	len1, len2 := len(s1), len(s2)
	
	// 创建矩阵
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}
	
	// 初始化矩阵
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}
	
	// 填充矩阵
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // 删除
				matrix[i][j-1]+1,      // 插入
				matrix[i-1][j-1]+cost, // 替换
			)
		}
	}
	
	return matrix[len1][len2]
}

// min 返回最小值
func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}

// ExtractNumbers 提取字符串中的数字
func (sh *StringHelper) ExtractNumbers(s string) []string {
	reg := regexp.MustCompile(`\d+\.?\d*`)
	return reg.FindAllString(s, -1)
}

// ExtractEmails 提取邮箱地址
func (sh *StringHelper) ExtractEmails(s string) []string {
	reg := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	return reg.FindAllString(s, -1)
}

// ExtractURLs 提取URL
func (sh *StringHelper) ExtractURLs(s string) []string {
	reg := regexp.MustCompile(`https?://[^\s<>"]+`)
	return reg.FindAllString(s, -1)
}

// MaskString 掩码字符串（如手机号、邮箱等）
func (sh *StringHelper) MaskString(s string, start, length int, mask rune) string {
	runes := []rune(s)
	runeLen := len(runes)
	
	if start < 0 || start >= runeLen {
		return s
	}
	
	end := start + length
	if end > runeLen {
		end = runeLen
	}
	
	for i := start; i < end; i++ {
		runes[i] = mask
	}
	
	return string(runes)
}

// MaskEmail 掩码邮箱
func (sh *StringHelper) MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	
	username := parts[0]
	domain := parts[1]
	
	if len(username) <= 2 {
		return email
	}
	
	maskedUsername := string(username[0]) + strings.Repeat("*", len(username)-2) + string(username[len(username)-1])
	return maskedUsername + "@" + domain
}

// MaskPhone 掩码手机号
func (sh *StringHelper) MaskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	
	return phone[:3] + "****" + phone[7:]
}

// WordWrap 文本换行
func (sh *StringHelper) WordWrap(s string, width int) string {
	if width <= 0 {
		return s
	}
	
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}
	
	var lines []string
	var currentLine []string
	currentLength := 0
	
	for _, word := range words {
		wordLength := utf8.RuneCountInString(word)
		
		if currentLength+wordLength+len(currentLine) > width && len(currentLine) > 0 {
			lines = append(lines, strings.Join(currentLine, " "))
			currentLine = []string{word}
			currentLength = wordLength
		} else {
			currentLine = append(currentLine, word)
			currentLength += wordLength
		}
	}
	
	if len(currentLine) > 0 {
		lines = append(lines, strings.Join(currentLine, " "))
	}
	
	return strings.Join(lines, "\n")
}

// EscapeHTML 转义HTML特殊字符
func (sh *StringHelper) EscapeHTML(s string) string {
	replacements := map[string]string{
		"&":  "&amp;",
		"<":  "&lt;",
		">":  "&gt;",
		"\"": "&quot;",
		"'":  "&#39;",
	}
	
	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}
	
	return s
}

// UnescapeHTML 反转义HTML特殊字符
func (sh *StringHelper) UnescapeHTML(s string) string {
	replacements := map[string]string{
		"&amp;":  "&",
		"&lt;":   "<",
		"&gt;":   ">",
		"&quot;": "\"",
		"&#39;":  "'",
	}
	
	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}
	
	return s
}

// Slugify 转换为URL友好的slug
func (sh *StringHelper) Slugify(s string) string {
	// 转换为小写
	s = strings.ToLower(s)
	
	// 移除特殊字符，保留字母、数字、空格和连字符
	reg := regexp.MustCompile(`[^a-z0-9\s-]`)
	s = reg.ReplaceAllString(s, "")
	
	// 将空格替换为连字符
	s = strings.ReplaceAll(s, " ", "-")
	
	// 移除多余的连字符
	reg = regexp.MustCompile(`-+`)
	s = reg.ReplaceAllString(s, "-")
	
	// 移除首尾连字符
	s = strings.Trim(s, "-")
	
	return s
}

// FormatNumber 格式化数字字符串（添加千分位分隔符）
func (sh *StringHelper) FormatNumber(s string) string {
	// 检查是否为有效数字
	if !sh.IsNumeric(s) {
		return s
	}
	
	// 分离整数和小数部分
	parts := strings.Split(s, ".")
	intPart := parts[0]
	
	// 从右到左每三位添加逗号
	var result []string
	for i, r := range sh.Reverse(intPart) {
		if i > 0 && i%3 == 0 {
			result = append(result, ",")
		}
		result = append(result, string(r))
	}
	
	formatted := sh.Reverse(strings.Join(result, ""))
	
	// 如果有小数部分，添加回去
	if len(parts) > 1 {
		formatted += "." + parts[1]
	}
	
	return formatted
}