package utils

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// YAMLGenerator YAML生成器
type YAMLGenerator struct {
	indent string
}

// NewYAMLGenerator 创建新的YAML生成器
func NewYAMLGenerator() *YAMLGenerator {
	return &YAMLGenerator{
		indent: "  ", // 使用2个空格缩进
	}
}

// Generate 生成YAML字符串
func (yg *YAMLGenerator) Generate(data interface{}) (string, error) {
	var builder strings.Builder
	
	// 添加YAML头部注释
	builder.WriteString("# Qlib工作流配置文件\n")
	builder.WriteString("# 由Qlib可视化平台自动生成\n")
	builder.WriteString("# https://github.com/microsoft/qlib\n\n")
	
	err := yg.writeValue(&builder, data, 0)
	if err != nil {
		return "", err
	}
	
	return builder.String(), nil
}

// writeValue 写入值
func (yg *YAMLGenerator) writeValue(builder *strings.Builder, value interface{}, depth int) error {
	if value == nil {
		builder.WriteString("null")
		return nil
	}
	
	rv := reflect.ValueOf(value)
	rt := reflect.TypeOf(value)
	
	switch rv.Kind() {
	case reflect.String:
		yg.writeString(builder, rv.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		builder.WriteString(strconv.FormatInt(rv.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		builder.WriteString(strconv.FormatUint(rv.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		builder.WriteString(strconv.FormatFloat(rv.Float(), 'f', -1, 64))
	case reflect.Bool:
		builder.WriteString(strconv.FormatBool(rv.Bool()))
	case reflect.Slice, reflect.Array:
		return yg.writeArray(builder, rv, depth)
	case reflect.Map:
		return yg.writeMap(builder, rv, depth)
	case reflect.Struct:
		return yg.writeStruct(builder, rv, rt, depth)
	case reflect.Interface:
		return yg.writeValue(builder, rv.Elem().Interface(), depth)
	case reflect.Ptr:
		if rv.IsNil() {
			builder.WriteString("null")
		} else {
			return yg.writeValue(builder, rv.Elem().Interface(), depth)
		}
	default:
		return fmt.Errorf("unsupported type: %s", rt.String())
	}
	
	return nil
}

// writeString 写入字符串
func (yg *YAMLGenerator) writeString(builder *strings.Builder, value string) {
	if yg.needsQuotes(value) {
		builder.WriteString(fmt.Sprintf(`"%s"`, yg.escapeString(value)))
	} else {
		builder.WriteString(value)
	}
}

// needsQuotes 检查字符串是否需要引号
func (yg *YAMLGenerator) needsQuotes(value string) bool {
	if value == "" {
		return true
	}
	
	// 包含特殊字符需要引号
	specialChars := []string{":", "{", "}", "[", "]", ",", "&", "*", "#", "?", "|", "-", "<", ">", "=", "!", "%", "@", "`"}
	for _, char := range specialChars {
		if strings.Contains(value, char) {
			return true
		}
	}
	
	// 数字字符串需要引号
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return true
	}
	
	// 布尔值字符串需要引号
	if value == "true" || value == "false" || value == "yes" || value == "no" || value == "on" || value == "off" {
		return true
	}
	
	// 以空格开头或结尾需要引号
	if strings.HasPrefix(value, " ") || strings.HasSuffix(value, " ") {
		return true
	}
	
	return false
}

// escapeString 转义字符串
func (yg *YAMLGenerator) escapeString(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\r", "\\r")
	value = strings.ReplaceAll(value, "\t", "\\t")
	return value
}

// writeArray 写入数组
func (yg *YAMLGenerator) writeArray(builder *strings.Builder, rv reflect.Value, depth int) error {
	length := rv.Len()
	
	if length == 0 {
		builder.WriteString("[]")
		return nil
	}
	
	// 检查是否是简单类型的数组，可以写成一行
	if yg.isSimpleArray(rv) && length <= 5 {
		builder.WriteString("[")
		for i := 0; i < length; i++ {
			if i > 0 {
				builder.WriteString(", ")
			}
			err := yg.writeValue(builder, rv.Index(i).Interface(), depth)
			if err != nil {
				return err
			}
		}
		builder.WriteString("]")
		return nil
	}
	
	// 多行数组格式
	for i := 0; i < length; i++ {
		builder.WriteString("\n")
		builder.WriteString(strings.Repeat(yg.indent, depth))
		builder.WriteString("- ")
		
		item := rv.Index(i).Interface()
		if yg.isComplexType(item) {
			// 复杂类型需要缩进
			err := yg.writeValue(builder, item, depth+1)
			if err != nil {
				return err
			}
		} else {
			// 简单类型直接写在同一行
			err := yg.writeValue(builder, item, depth+1)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}

// writeMap 写入映射
func (yg *YAMLGenerator) writeMap(builder *strings.Builder, rv reflect.Value, depth int) error {
	keys := rv.MapKeys()
	if len(keys) == 0 {
		builder.WriteString("{}")
		return nil
	}
	
	// 对键进行排序
	keyStrings := make([]string, len(keys))
	for i, key := range keys {
		keyStrings[i] = fmt.Sprintf("%v", key.Interface())
	}
	sort.Strings(keyStrings)
	
	for i, keyStr := range keyStrings {
		if i > 0 || depth > 0 {
			builder.WriteString("\n")
			builder.WriteString(strings.Repeat(yg.indent, depth))
		}
		
		// 写入键
		yg.writeString(builder, keyStr)
		builder.WriteString(": ")
		
		// 找到对应的值
		var value interface{}
		for _, key := range keys {
			if fmt.Sprintf("%v", key.Interface()) == keyStr {
				value = rv.MapIndex(key).Interface()
				break
			}
		}
		
		// 写入值
		if yg.isComplexType(value) {
			err := yg.writeValue(builder, value, depth+1)
			if err != nil {
				return err
			}
		} else {
			err := yg.writeValue(builder, value, depth)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}

// writeStruct 写入结构体
func (yg *YAMLGenerator) writeStruct(builder *strings.Builder, rv reflect.Value, rt reflect.Type, depth int) error {
	numField := rv.NumField()
	fieldCount := 0
	
	for i := 0; i < numField; i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)
		
		// 跳过未导出的字段
		if !fieldValue.CanInterface() {
			continue
		}
		
		// 跳过零值字段（可选）
		if yg.isZeroValue(fieldValue) {
			continue
		}
		
		// 获取字段名
		fieldName := yg.getFieldName(field)
		if fieldName == "" || fieldName == "-" {
			continue
		}
		
		if fieldCount > 0 || depth > 0 {
			builder.WriteString("\n")
			builder.WriteString(strings.Repeat(yg.indent, depth))
		}
		
		// 写入字段名
		yg.writeString(builder, fieldName)
		builder.WriteString(": ")
		
		// 写入字段值
		if yg.isComplexType(fieldValue.Interface()) {
			err := yg.writeValue(builder, fieldValue.Interface(), depth+1)
			if err != nil {
				return err
			}
		} else {
			err := yg.writeValue(builder, fieldValue.Interface(), depth)
			if err != nil {
				return err
			}
		}
		
		fieldCount++
	}
	
	return nil
}

// isSimpleArray 检查是否是简单类型数组
func (yg *YAMLGenerator) isSimpleArray(rv reflect.Value) bool {
	if rv.Len() == 0 {
		return true
	}
	
	for i := 0; i < rv.Len(); i++ {
		if yg.isComplexType(rv.Index(i).Interface()) {
			return false
		}
	}
	
	return true
}

// isComplexType 检查是否是复杂类型
func (yg *YAMLGenerator) isComplexType(value interface{}) bool {
	if value == nil {
		return false
	}
	
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Map, reflect.Struct, reflect.Slice, reflect.Array:
		// 空的集合类型不算复杂
		if (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) && rv.Len() == 0 {
			return false
		}
		if rv.Kind() == reflect.Map && rv.Len() == 0 {
			return false
		}
		return true
	case reflect.Interface, reflect.Ptr:
		if rv.IsNil() {
			return false
		}
		return yg.isComplexType(rv.Elem().Interface())
	default:
		return false
	}
}

// isZeroValue 检查是否是零值
func (yg *YAMLGenerator) isZeroValue(rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.String:
		return rv.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Slice, reflect.Array:
		return rv.Len() == 0
	case reflect.Map:
		return rv.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return rv.IsNil()
	default:
		return false
	}
}

// getFieldName 获取字段名
func (yg *YAMLGenerator) getFieldName(field reflect.StructField) string {
	// 优先使用yaml标签
	if yamlTag := field.Tag.Get("yaml"); yamlTag != "" {
		parts := strings.Split(yamlTag, ",")
		if parts[0] != "" {
			return parts[0]
		}
	}
	
	// 使用json标签
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] != "" && parts[0] != "-" {
			return parts[0]
		}
	}
	
	// 使用字段名的小写形式
	return strings.ToLower(field.Name)
}

// GenerateWorkflowYAML 生成工作流专用的YAML格式
func (yg *YAMLGenerator) GenerateWorkflowYAML(config map[string]interface{}) (string, error) {
	var builder strings.Builder
	
	// YAML头部
	builder.WriteString("# Qlib工作流配置文件\n")
	builder.WriteString("# Generated by Qlib Visualization Platform\n")
	builder.WriteString("# Visit: https://github.com/microsoft/qlib\n\n")
	
	// 基本信息
	if name, ok := config["name"].(string); ok {
		builder.WriteString(fmt.Sprintf("name: %s\n", name))
	}
	if desc, ok := config["description"].(string); ok && desc != "" {
		builder.WriteString(fmt.Sprintf("description: %s\n", desc))
	}
	if version, ok := config["version"].(string); ok {
		builder.WriteString(fmt.Sprintf("version: %s\n", version))
	}
	
	builder.WriteString("\n")
	
	// 全局配置
	if globalConfig, ok := config["config"].(map[string]interface{}); ok && len(globalConfig) > 0 {
		builder.WriteString("config:\n")
		err := yg.writeValue(&builder, globalConfig, 1)
		if err != nil {
			return "", err
		}
		builder.WriteString("\n\n")
	}
	
	// 工作流步骤
	if steps, ok := config["steps"].([]interface{}); ok && len(steps) > 0 {
		builder.WriteString("workflow:\n")
		builder.WriteString("  steps:\n")
		
		for i, step := range steps {
			stepMap, ok := step.(map[string]interface{})
			if !ok {
				continue
			}
			
			builder.WriteString(fmt.Sprintf("    - name: %s\n", stepMap["name"]))
			builder.WriteString(fmt.Sprintf("      type: %s\n", stepMap["type"]))
			
			if desc, exists := stepMap["description"].(string); exists && desc != "" {
				builder.WriteString(fmt.Sprintf("      description: %s\n", desc))
			}
			
			if enabled, exists := stepMap["enabled"].(bool); exists {
				builder.WriteString(fmt.Sprintf("      enabled: %t\n", enabled))
			}
			
			if required, exists := stepMap["required"].(bool); exists {
				builder.WriteString(fmt.Sprintf("      required: %t\n", required))
			}
			
			if deps, exists := stepMap["dependencies"].([]interface{}); exists && len(deps) > 0 {
				builder.WriteString("      dependencies:\n")
				for _, dep := range deps {
					builder.WriteString(fmt.Sprintf("        - %s\n", dep))
				}
			}
			
			if stepConfig, exists := stepMap["config"].(map[string]interface{}); exists && len(stepConfig) > 0 {
				builder.WriteString("      config:\n")
				for key, value := range stepConfig {
					builder.WriteString(fmt.Sprintf("        %s: ", key))
					err := yg.writeValue(&builder, value, 0)
					if err != nil {
						return "", err
					}
					builder.WriteString("\n")
				}
			}
			
			if i < len(steps)-1 {
				builder.WriteString("\n")
			}
		}
	}
	
	// 元数据
	if metadata, ok := config["metadata"].(map[string]interface{}); ok && len(metadata) > 0 {
		builder.WriteString("\n# Metadata\n")
		builder.WriteString("metadata:\n")
		err := yg.writeValue(&builder, metadata, 1)
		if err != nil {
			return "", err
		}
	}
	
	return builder.String(), nil
}