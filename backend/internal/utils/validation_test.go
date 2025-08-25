package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	
	assert.NotNil(t, validator)
	assert.NotNil(t, validator.rules)
	assert.Len(t, validator.rules, 0)
}

func TestAddRule(t *testing.T) {
	validator := NewValidator()
	
	rule := ValidationRule{
		Field:    "email",
		Required: true,
		Type:     "email",
		Message:  "请输入有效的邮箱地址",
	}
	
	validator.AddRule(rule)
	
	assert.Len(t, validator.rules, 1)
	assert.Equal(t, "email", validator.rules[0].Field)
	assert.True(t, validator.rules[0].Required)
	assert.Equal(t, "email", validator.rules[0].Type)
}

func TestValidateData(t *testing.T) {
	validator := NewValidator()
	
	// 添加验证规则
	validator.AddRule(ValidationRule{
		Field:    "name",
		Required: true,
		Type:     "string",
		Min:      2,
		Max:      50,
		Message:  "姓名长度必须在2-50个字符之间",
	})
	
	validator.AddRule(ValidationRule{
		Field:    "age",
		Required: true,
		Type:     "int",
		Min:      0,
		Max:      150,
		Message:  "年龄必须在0-150之间",
	})
	
	validator.AddRule(ValidationRule{
		Field:    "email",
		Required: true,
		Type:     "email",
		Message:  "请输入有效的邮箱地址",
	})

	// 测试有效数据
	validData := map[string]interface{}{
		"name":  "张三",
		"age":   25,
		"email": "zhangsan@example.com",
	}
	
	result := validator.ValidateData(validData)
	assert.True(t, result.Valid)
	assert.Len(t, result.Errors, 0)

	// 测试无效数据 - 缺少必填字段
	invalidData1 := map[string]interface{}{
		"age":   25,
		"email": "zhangsan@example.com",
	}
	
	result = validator.ValidateData(invalidData1)
	assert.False(t, result.Valid)
	assert.Greater(t, len(result.Errors), 0)
	
	// 找到name字段的错误
	var nameError *ValidationError
	for i := range result.Errors {
		if result.Errors[i].Field == "name" {
			nameError = &result.Errors[i]
			break
		}
	}
	assert.NotNil(t, nameError)
	// 检查错误消息是否包含必填字段相关内容
	assert.True(t, nameError.Message == "姓名长度必须在2-50个字符之间" || nameError.Message == "name是必填字段")

	// 测试无效数据 - 类型错误
	invalidData2 := map[string]interface{}{
		"name":  "张三",
		"age":   "invalid_age", // 应该是int，但提供的是string
		"email": "zhangsan@example.com",
	}
	
	result = validator.ValidateData(invalidData2)
	assert.False(t, result.Valid)
	assert.Greater(t, len(result.Errors), 0)

	// 测试无效数据 - 范围错误
	invalidData3 := map[string]interface{}{
		"name":  "A", // 太短
		"age":   200, // 超出范围
		"email": "invalid_email", // 无效邮箱格式
	}
	
	result = validator.ValidateData(invalidData3)
	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 3) // 应该有3个错误
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"user.name@domain.co.uk", true},
		{"user+tag@example.org", true},
		{"invalid_email", false},
		{"@example.com", false},
		{"test@", false},
		{"", false},
		{"test..test@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := ValidateEmail(tt.email)
			assert.Equal(t, tt.valid, result, "Email: %s", tt.email)
		})
	}
}

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		phone string
		valid bool
	}{
		{"13812345678", true},
		{"15987654321", true},
		{"18000000000", true},
		{"12345678901", false}, // 不是有效的手机号段
		{"1381234567", false},  // 太短
		{"138123456789", false}, // 太长
		{"abcdefghijk", false}, // 非数字
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.phone, func(t *testing.T) {
			result := ValidatePhone(tt.phone)
			assert.Equal(t, tt.valid, result, "Phone: %s", tt.phone)
		})
	}
}

func TestValidateIDCard(t *testing.T) {
	tests := []struct {
		idCard string
		valid  bool
	}{
		{"110101199003077477", true}, // 有效的18位身份证
		{"11010119900307747X", true}, // 末位为X的身份证
		{"110101900307123", true},    // 有效的15位身份证
		{"123456789012345678", false}, // 无效的18位
		{"12345678901234567X", false}, // 无效的校验位
		{"12345678901234", false},     // 无效的15位
		{"abcd", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.idCard, func(t *testing.T) {
			result := ValidateIDCard(tt.idCard)
			assert.Equal(t, tt.valid, result, "ID Card: %s", tt.idCard)
		})
	}
}

func TestValidateDate(t *testing.T) {
	tests := []struct {
		date   string
		format string
		valid  bool
	}{
		{"2023-12-25", "2006-01-02", true},
		{"2023/12/25", "2006/01/02", true},
		{"25-12-2023", "02-01-2006", true},
		{"2023-13-25", "2006-01-02", false}, // 无效月份
		{"2023-12-32", "2006-01-02", false}, // 无效日期
		{"invalid", "2006-01-02", false},
		{"", "2006-01-02", false},
	}

	for _, tt := range tests {
		t.Run(tt.date, func(t *testing.T) {
			result := ValidateDate(tt.date, tt.format)
			assert.Equal(t, tt.valid, result, "Date: %s, Format: %s", tt.date, tt.format)
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		url   string
		valid bool
	}{
		{"https://www.example.com", true},
		{"http://example.com/path", true},
		{"ftp://ftp.example.com", true},
		{"https://subdomain.example.co.uk/path?query=value", true},
		{"invalid_url", false},
		{"http://", false},
		{"://example.com", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := ValidateURL(tt.url)
			assert.Equal(t, tt.valid, result, "URL: %s", tt.url)
		})
	}
}

func TestValidateJSON(t *testing.T) {
	tests := []struct {
		json  string
		valid bool
	}{
		{`{"key": "value"}`, true},
		{`{"number": 123, "array": [1,2,3]}`, true},
		{`[]`, true},
		{`"string"`, true},
		{`123`, true},
		{`true`, true},
		{`null`, true},
		{`{invalid json}`, false},
		{`{"key": }`, false},
		{`{key: "value"}`, false}, // 键没有引号
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.json, func(t *testing.T) {
			result := ValidateJSON(tt.json)
			assert.Equal(t, tt.valid, result, "JSON: %s", tt.json)
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password string
		minLen   int
		valid    bool
	}{
		{"StrongPass123!", 8, true},
		{"weakpass", 8, false}, // 缺少数字和大写字母
		{"WEAKPASS123", 8, false}, // 缺少小写字母
		{"WeakPass", 8, false}, // 缺少数字
		{"weak123", 8, false}, // 缺少大写字母
		{"Short1!", 8, false}, // 太短
		{"VeryStrongPassword123!", 12, true},
		{"", 8, false},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			result := ValidatePassword(tt.password, tt.minLen)
			assert.Equal(t, tt.valid, result, "Password: %s, MinLen: %d", tt.password, tt.minLen)
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		username string
		valid    bool
	}{
		{"user123", true},
		{"test_user", true},
		{"user-name", true},
		{"User123", true},
		{"123user", false}, // 不能以数字开头
		{"_user", false}, // 不能以下划线开头
		{"-user", false}, // 不能以连字符开头
		{"user@123", false}, // 包含特殊字符
		{"us", false}, // 太短
		{"verylongusernamethatexceedsthelimitofcharactersandmorethan50", false}, // 太长，59个字符
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.username, func(t *testing.T) {
			result := ValidateUsername(tt.username)
			assert.Equal(t, tt.valid, result, "Username: %s", tt.username)
		})
	}
}

func TestValidateIPAddress(t *testing.T) {
	tests := []struct {
		ip    string
		valid bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"127.0.0.1", true},
		{"255.255.255.255", true},
		{"0.0.0.0", true},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", true}, // IPv6
		{"::1", true}, // IPv6 loopback
		{"256.1.1.1", false}, // 超出范围
		{"192.168.1", false}, // 不完整
		{"192.168.1.1.1", false}, // 太多段
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := ValidateIPAddress(tt.ip)
			assert.Equal(t, tt.valid, result, "IP: %s", tt.ip)
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"normal text", "normal text"},
		{"text with & symbols", "text with &amp; symbols"},
		{"quotes \" and '", "quotes &#34; and &#39;"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrimSpaces(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello world  ", "hello world"},
		{"\t\n  text  \r\n", "text"},
		{"no spaces", "no spaces"},
		{"   ", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := TrimSpaces(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateRange(t *testing.T) {
	// 测试整数范围
	assert.True(t, ValidateRange(5, 1, 10))
	assert.True(t, ValidateRange(1, 1, 10)) // 边界值
	assert.True(t, ValidateRange(10, 1, 10)) // 边界值
	assert.False(t, ValidateRange(0, 1, 10))
	assert.False(t, ValidateRange(11, 1, 10))

	// 测试浮点数范围
	assert.True(t, ValidateRange(5.5, 1.0, 10.0))
	assert.False(t, ValidateRange(0.5, 1.0, 10.0))
	assert.False(t, ValidateRange(10.5, 1.0, 10.0))

	// 测试字符串长度范围
	assert.True(t, ValidateRange("hello", 3, 10)) // 长度为5
	assert.False(t, ValidateRange("hi", 3, 10)) // 长度为2，小于最小值
	assert.False(t, ValidateRange("this is too long", 3, 10)) // 长度超过最大值
}