package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
)

// ValidationRule 验证规则
type ValidationRule struct {
	Field    string      `json:"field"`
	Required bool        `json:"required"`
	Type     string      `json:"type"`     // string, int, float, email, phone, date, etc.
	Min      interface{} `json:"min"`      // 最小值/最小长度
	Max      interface{} `json:"max"`      // 最大值/最大长度
	Pattern  string      `json:"pattern"`  // 正则表达式
	Message  string      `json:"message"`  // 自定义错误消息
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors"`
}

// Validator 验证器
type Validator struct {
	rules []ValidationRule
}

// NewValidator 创建验证器
func NewValidator() *Validator {
	return &Validator{
		rules: make([]ValidationRule, 0),
	}
}

// AddRule 添加验证规则
func (v *Validator) AddRule(rule ValidationRule) {
	v.rules = append(v.rules, rule)
}

// Validate 验证数据
func (v *Validator) Validate(data map[string]interface{}) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}

	for _, rule := range v.rules {
		if err := v.validateField(data, rule); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, *err)
		}
	}

	return result
}

// validateField 验证单个字段
func (v *Validator) validateField(data map[string]interface{}, rule ValidationRule) *ValidationError {
	value, exists := data[rule.Field]

	// 检查必填字段
	if rule.Required && (!exists || v.isEmpty(value)) {
		return &ValidationError{
			Field:   rule.Field,
			Message: getErrorMessage(rule, fmt.Sprintf("%s是必填字段", rule.Field)),
		}
	}

	// 如果字段不存在且不是必填，跳过验证
	if !exists {
		return nil
	}

	// 如果值为空且不是必填，跳过验证
	if v.isEmpty(value) {
		return nil
	}

	// 类型验证
	if err := v.validateType(value, rule); err != nil {
		return err
	}

	// 长度/范围验证
	if err := v.validateRange(value, rule); err != nil {
		return err
	}

	// 正则表达式验证
	if err := v.validatePattern(value, rule); err != nil {
		return err
	}

	return nil
}

// isEmpty 检查值是否为空
func (v *Validator) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}

// validateType 类型验证
func (v *Validator) validateType(value interface{}, rule ValidationRule) *ValidationError {
	switch rule.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return &ValidationError{
				Field:   rule.Field,
				Message: getErrorMessage(rule, fmt.Sprintf("%s必须是字符串类型", rule.Field)),
			}
		}
	case "int":
		if err := v.validateInt(value); err != nil {
			return &ValidationError{
				Field:   rule.Field,
				Message: getErrorMessage(rule, fmt.Sprintf("%s必须是整数类型", rule.Field)),
			}
		}
	case "float":
		if err := v.validateFloat(value); err != nil {
			return &ValidationError{
				Field:   rule.Field,
				Message: getErrorMessage(rule, fmt.Sprintf("%s必须是数字类型", rule.Field)),
			}
		}
	case "email":
		if err := v.validateEmail(value); err != nil {
			return &ValidationError{
				Field:   rule.Field,
				Message: getErrorMessage(rule, fmt.Sprintf("%s必须是有效的邮箱地址", rule.Field)),
			}
		}
	case "phone":
		if err := v.validatePhone(value); err != nil {
			return &ValidationError{
				Field:   rule.Field,
				Message: getErrorMessage(rule, fmt.Sprintf("%s必须是有效的手机号", rule.Field)),
			}
		}
	case "date":
		if err := v.validateDate(value); err != nil {
			return &ValidationError{
				Field:   rule.Field,
				Message: getErrorMessage(rule, fmt.Sprintf("%s必须是有效的日期格式", rule.Field)),
			}
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return &ValidationError{
				Field:   rule.Field,
				Message: getErrorMessage(rule, fmt.Sprintf("%s必须是布尔类型", rule.Field)),
			}
		}
	}

	return nil
}

// validateInt 验证整数
func (v *Validator) validateInt(value interface{}) error {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return nil
	case float64:
		// JSON数字默认解析为float64，需要检查是否为整数
		if f := value.(float64); f == float64(int64(f)) {
			return nil
		}
		return fmt.Errorf("不是整数")
	case string:
		if _, err := strconv.Atoi(value.(string)); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("不是整数类型")
	}
}

// validateFloat 验证浮点数
func (v *Validator) validateFloat(value interface{}) error {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return nil
	case string:
		if _, err := strconv.ParseFloat(value.(string), 64); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("不是数字类型")
	}
}

// validateEmail 验证邮箱
func (v *Validator) validateEmail(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("不是字符串类型")
	}

	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(emailPattern, str)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("邮箱格式无效")
	}
	return nil
}

// validatePhone 验证手机号
func (v *Validator) validatePhone(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("不是字符串类型")
	}

	// 中国手机号正则
	phonePattern := `^1[3-9]\d{9}$`
	matched, err := regexp.MatchString(phonePattern, str)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("手机号格式无效")
	}
	return nil
}

// validateDate 验证日期
func (v *Validator) validateDate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("不是字符串类型")
	}

	// 支持多种日期格式
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		time.RFC3339,
	}

	for _, format := range formats {
		if _, err := time.Parse(format, str); err == nil {
			return nil
		}
	}

	return fmt.Errorf("日期格式无效")
}

// validateRange 验证范围
func (v *Validator) validateRange(value interface{}, rule ValidationRule) *ValidationError {
	// 验证最小值/最小长度
	if rule.Min != nil {
		if err := v.validateMin(value, rule.Min, rule); err != nil {
			return err
		}
	}

	// 验证最大值/最大长度
	if rule.Max != nil {
		if err := v.validateMax(value, rule.Max, rule); err != nil {
			return err
		}
	}

	return nil
}

// validateMin 验证最小值
func (v *Validator) validateMin(value interface{}, min interface{}, rule ValidationRule) *ValidationError {
	switch v := value.(type) {
	case string:
		minLen, ok := min.(int)
		if ok && len(v) < minLen {
			return &ValidationError{
				Field:   rule.Field,
				Message: getErrorMessage(rule, fmt.Sprintf("%s长度不能少于%d个字符", rule.Field, minLen)),
			}
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		if err := compareNumbers(value, min, ">="); err != nil {
			return &ValidationError{
				Field:   rule.Field,
				Message: getErrorMessage(rule, fmt.Sprintf("%s不能小于%v", rule.Field, min)),
			}
		}
	}
	return nil
}

// validateMax 验证最大值
func (v *Validator) validateMax(value interface{}, max interface{}, rule ValidationRule) *ValidationError {
	switch v := value.(type) {
	case string:
		maxLen, ok := max.(int)
		if ok && len(v) > maxLen {
			return &ValidationError{
				Field:   rule.Field,
				Message: getErrorMessage(rule, fmt.Sprintf("%s长度不能超过%d个字符", rule.Field, maxLen)),
			}
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		if err := compareNumbers(value, max, "<="); err != nil {
			return &ValidationError{
				Field:   rule.Field,
				Message: getErrorMessage(rule, fmt.Sprintf("%s不能大于%v", rule.Field, max)),
			}
		}
	}
	return nil
}

// compareNumbers 比较数字
func compareNumbers(value1, value2 interface{}, operator string) error {
	val1 := toFloat64(value1)
	val2 := toFloat64(value2)

	switch operator {
	case ">=":
		if val1 < val2 {
			return fmt.Errorf("值太小")
		}
	case "<=":
		if val1 > val2 {
			return fmt.Errorf("值太大")
		}
	case ">":
		if val1 <= val2 {
			return fmt.Errorf("值不够大")
		}
	case "<":
		if val1 >= val2 {
			return fmt.Errorf("值不够小")
		}
	case "==":
		if val1 != val2 {
			return fmt.Errorf("值不相等")
		}
	}

	return nil
}

// toFloat64 转换为float64
func toFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	default:
		return 0
	}
}

// validatePattern 正则表达式验证
func (v *Validator) validatePattern(value interface{}, rule ValidationRule) *ValidationError {
	if rule.Pattern == "" {
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return &ValidationError{
			Field:   rule.Field,
			Message: getErrorMessage(rule, fmt.Sprintf("%s必须是字符串类型才能进行正则验证", rule.Field)),
		}
	}

	matched, err := regexp.MatchString(rule.Pattern, str)
	if err != nil {
		return &ValidationError{
			Field:   rule.Field,
			Message: getErrorMessage(rule, fmt.Sprintf("%s正则表达式验证失败", rule.Field)),
		}
	}

	if !matched {
		return &ValidationError{
			Field:   rule.Field,
			Message: getErrorMessage(rule, fmt.Sprintf("%s格式不正确", rule.Field)),
		}
	}

	return nil
}

// getErrorMessage 获取错误消息
func getErrorMessage(rule ValidationRule, defaultMsg string) string {
	if rule.Message != "" {
		return rule.Message
	}
	return defaultMsg
}

// ValidateStruct 验证结构体
func ValidateStruct(s interface{}) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}

	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "root",
			Message: "必须是结构体类型",
		})
		return result
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// 检查validate标签
		validateTag := fieldType.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		// 解析验证规则
		rules := parseValidateTag(validateTag)
		fieldName := getFieldName(fieldType)

		// 验证字段
		if err := validateStructField(field.Interface(), fieldName, rules); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, *err)
		}
	}

	return result
}

// parseValidateTag 解析validate标签
func parseValidateTag(tag string) map[string]string {
	rules := make(map[string]string)
	parts := strings.Split(tag, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			rules[kv[0]] = kv[1]
		} else {
			rules[part] = "true"
		}
	}

	return rules
}

// getFieldName 获取字段名
func getFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] != "" && parts[0] != "-" {
			return parts[0]
		}
	}
	return field.Name
}

// validateStructField 验证结构体字段
func validateStructField(value interface{}, fieldName string, rules map[string]string) *ValidationError {
	// 检查required规则
	if _, required := rules["required"]; required {
		if value == nil || (reflect.ValueOf(value).Kind() == reflect.String && value.(string) == "") {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%s是必填字段", fieldName),
			}
		}
	}

	// 如果值为空且不是必填，跳过其他验证
	if value == nil || (reflect.ValueOf(value).Kind() == reflect.String && value.(string) == "") {
		return nil
	}

	// email验证
	if _, ok := rules["email"]; ok {
		if str, isStr := value.(string); isStr {
			emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
			if matched, _ := regexp.MatchString(emailPattern, str); !matched {
				return &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("%s必须是有效的邮箱地址", fieldName),
				}
			}
		}
	}

	// min长度验证
	if minStr, ok := rules["min"]; ok {
		if min, err := strconv.Atoi(minStr); err == nil {
			if str, isStr := value.(string); isStr && len(str) < min {
				return &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("%s长度不能少于%d个字符", fieldName, min),
				}
			}
		}
	}

	// max长度验证
	if maxStr, ok := rules["max"]; ok {
		if max, err := strconv.Atoi(maxStr); err == nil {
			if str, isStr := value.(string); isStr && len(str) > max {
				return &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("%s长度不能超过%d个字符", fieldName, max),
				}
			}
		}
	}

	return nil
}

// ValidateGinRequest 验证Gin请求参数
func ValidateGinRequest(c *gin.Context, rules []ValidationRule) ValidationResult {
	// 获取JSON数据
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		return ValidationResult{
			Valid: false,
			Errors: []ValidationError{
				{
					Field:   "request",
					Message: "JSON格式错误",
				},
			},
		}
	}

	// 创建验证器并添加规则
	validator := NewValidator()
	for _, rule := range rules {
		validator.AddRule(rule)
	}

	return validator.Validate(data)
}


// ValidateUsernameAvailable 验证用户名是否可用（示例，需要数据库查询）
func ValidateUsernameAvailable(username string) *ValidationError {
	// 这里应该查询数据库检查用户名是否已存在
	// 为了示例，这里只做基本验证
	if len(username) < 3 {
		return &ValidationError{
			Field:   "username",
			Message: "用户名长度不能少于3个字符",
		}
	}

	if len(username) > 50 {
		return &ValidationError{
			Field:   "username",
			Message: "用户名长度不能超过50个字符",
		}
	}

	// 用户名只允许字母、数字和下划线
	usernamePattern := `^[a-zA-Z0-9_]+$`
	if matched, _ := regexp.MatchString(usernamePattern, username); !matched {
		return &ValidationError{
			Field:   "username",
			Message: "用户名只能包含字母、数字和下划线",
		}
	}

	return nil
}

// ValidateData 验证数据的快捷方法
func (v *Validator) ValidateData(data map[string]interface{}) ValidationResult {
	return v.Validate(data)
}

// 独立的验证函数

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) bool {
	// 更严格的邮箱验证，不允许连续的点
	if strings.Contains(email, "..") {
		return false
	}
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailPattern, email)
	return matched
}

// ValidatePhone 验证手机号格式
func ValidatePhone(phone string) bool {
	phonePattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(phonePattern, phone)
	return matched
}

// ValidateIDCard 验证身份证号码
func ValidateIDCard(idCard string) bool {
	if len(idCard) == 15 {
		// 15位身份证
		return regexp.MustCompile(`^\d{15}$`).MatchString(idCard)
	} else if len(idCard) == 18 {
		// 18位身份证
		if !regexp.MustCompile(`^\d{17}[\dXx]$`).MatchString(idCard) {
			return false
		}
		// 简化版校验码验证
		return true
	}
	return false
}

// ValidateDate 验证日期格式
func ValidateDate(dateStr, format string) bool {
	_, err := time.Parse(format, dateStr)
	return err == nil
}

// ValidateURL 验证URL格式
func ValidateURL(url string) bool {
	urlPattern := `^(https?|ftp)://[^\s/$.?#].[^\s]*$`
	matched, _ := regexp.MatchString(urlPattern, url)
	return matched
}

// ValidateJSON 验证JSON格式
func ValidateJSON(jsonStr string) bool {
	var result interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	return err == nil
}

// ValidatePassword 验证密码强度（返回bool）
func ValidatePassword(password string, minLen int) bool {
	if len(password) < minLen {
		return false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// ValidateUsername 验证用户名格式
func ValidateUsername(username string) bool {
	if len(username) < 3 || len(username) > 50 {
		return false
	}
	
	// 不能以数字、下划线或连字符开头
	if unicode.IsDigit(rune(username[0])) || username[0] == '_' || username[0] == '-' {
		return false
	}
	
	usernamePattern := `^[a-zA-Z][a-zA-Z0-9_-]*$`
	matched, _ := regexp.MatchString(usernamePattern, username)
	return matched
}

// ValidateIPAddress 验证IP地址
func ValidateIPAddress(ip string) bool {
	// IPv4
	ipv4Pattern := `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	if matched, _ := regexp.MatchString(ipv4Pattern, ip); matched {
		return true
	}
	
	// IPv6 (简化版)
	ipv6Pattern := `^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$|^::1$|^::$`
	if matched, _ := regexp.MatchString(ipv6Pattern, ip); matched {
		return true
	}
	
	// IPv6 压缩格式
	if strings.Contains(ip, "::") {
		parts := strings.Split(ip, "::")
		if len(parts) == 2 {
			return true // 简化处理，实际应该更严格
		}
	}
	
	return false
}

// SanitizeInput 清理输入，防止XSS
func SanitizeInput(input string) string {
	// HTML转义
	input = strings.ReplaceAll(input, "&", "&amp;")
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, "\"", "&#34;")
	input = strings.ReplaceAll(input, "'", "&#39;")
	return input
}

// TrimSpaces 去除多余空白字符
func TrimSpaces(input string) string {
	return strings.TrimSpace(input)
}

// ValidateRange 验证范围
func ValidateRange(value, min, max interface{}) bool {
	switch v := value.(type) {
	case int:
		minVal, minOk := min.(int)
		maxVal, maxOk := max.(int)
		if minOk && maxOk {
			return v >= minVal && v <= maxVal
		}
	case float64:
		minVal, minOk := min.(float64)
		maxVal, maxOk := max.(float64)
		if minOk && maxOk {
			return v >= minVal && v <= maxVal
		}
	case string:
		minVal, minOk := min.(int)
		maxVal, maxOk := max.(int)
		if minOk && maxOk {
			return len(v) >= minVal && len(v) <= maxVal
		}
	}
	return false
}