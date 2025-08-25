package qlib

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// SyntaxValidator Qlib语法验证器
type SyntaxValidator struct {
	pythonPath string
	qlibPath   string
}

// NewSyntaxValidator 创建新的语法验证器实例
func NewSyntaxValidator(pythonPath, qlibPath string) *SyntaxValidator {
	if pythonPath == "" {
		pythonPath = "python3"
	}
	return &SyntaxValidator{
		pythonPath: pythonPath,
		qlibPath:   qlibPath,
	}
}

// ValidationResult 验证结果
type ValidationResult struct {
	IsValid       bool     `json:"is_valid"`
	ErrorMsg      string   `json:"error_msg"`
	Suggestions   []string `json:"suggestions"`
	ParsedAST     string   `json:"parsed_ast"`
	UsedFields    []string `json:"used_fields"`
	UsedFunctions []string `json:"used_functions"`
}

// Validate 验证因子表达式语法
func (v *SyntaxValidator) Validate(expression string) (*ValidationResult, error) {
	// 首先进行基础语法检查
	if err := v.basicSyntaxCheck(expression); err != nil {
		return &ValidationResult{
			IsValid:     false,
			ErrorMsg:    err.Error(),
			Suggestions: v.generateSuggestions(expression),
		}, nil
	}

	// 调用Python进行深度语法验证
	result, err := v.advancedSyntaxCheck(expression)
	if err != nil {
		return nil, fmt.Errorf("高级语法验证失败: %v", err)
	}

	// 提取使用的字段和函数
	result.UsedFields = v.extractFields(expression)
	result.UsedFunctions = v.extractFunctions(expression)

	// 生成建议
	if !result.IsValid {
		result.Suggestions = v.generateSuggestions(expression)
	}

	return result, nil
}

// basicSyntaxCheck 基础语法检查
func (v *SyntaxValidator) basicSyntaxCheck(expression string) error {
	if strings.TrimSpace(expression) == "" {
		return fmt.Errorf("表达式不能为空")
	}

	// 检查括号匹配
	if !v.checkParenthesesBalance(expression) {
		return fmt.Errorf("括号不匹配")
	}

	// 检查基本语法规则
	if err := v.checkBasicSyntaxRules(expression); err != nil {
		return err
	}

	return nil
}

// checkParenthesesBalance 检查括号是否匹配
func (v *SyntaxValidator) checkParenthesesBalance(expression string) bool {
	balance := 0
	for _, char := range expression {
		switch char {
		case '(':
			balance++
		case ')':
			balance--
			if balance < 0 {
				return false
			}
		}
	}
	return balance == 0
}

// checkBasicSyntaxRules 检查基本语法规则
func (v *SyntaxValidator) checkBasicSyntaxRules(expression string) error {
	// 检查是否包含无效字符
	invalidChars := []string{"；", "、"}
	for _, char := range invalidChars {
		if strings.Contains(expression, char) {
			return fmt.Errorf("包含无效字符: %s", char)
		}
	}

	// 检查操作符使用
	if strings.Contains(expression, "//") {
		return fmt.Errorf("请使用 / 而不是 // 进行除法运算")
	}

	// 检查字段名格式
	fieldPattern := regexp.MustCompile(`\$[a-zA-Z_][a-zA-Z0-9_]*`)
	fields := fieldPattern.FindAllString(expression, -1)
	validFields := map[string]bool{
		"$open": true, "$high": true, "$low": true, "$close": true,
		"$volume": true, "$factor": true, "$vwap": true, "$amount": true,
		"$pctchange": true, "$adjclose": true,
	}

	for _, field := range fields {
		if !validFields[field] {
			return fmt.Errorf("未知字段: %s", field)
		}
	}

	return nil
}

// advancedSyntaxCheck 高级语法检查（使用Python）
func (v *SyntaxValidator) advancedSyntaxCheck(expression string) (*ValidationResult, error) {
	scriptArgs := map[string]interface{}{
		"action":     "validate_syntax",
		"expression": expression,
	}

	result, err := v.executePythonScript(scriptArgs)
	if err != nil {
		return nil, err
	}

	validationResult := &ValidationResult{}
	if data, ok := result["data"].(map[string]interface{}); ok {
		if valid, ok := data["is_valid"].(bool); ok {
			validationResult.IsValid = valid
		}
		if errorMsg, ok := data["error_msg"].(string); ok {
			validationResult.ErrorMsg = errorMsg
		}
		if parsedAST, ok := data["parsed_ast"].(string); ok {
			validationResult.ParsedAST = parsedAST
		}
	}

	return validationResult, nil
}

// extractFields 提取表达式中使用的字段
func (v *SyntaxValidator) extractFields(expression string) []string {
	fieldPattern := regexp.MustCompile(`\$[a-zA-Z_][a-zA-Z0-9_]*`)
	fields := fieldPattern.FindAllString(expression, -1)
	
	// 去重
	fieldMap := make(map[string]bool)
	for _, field := range fields {
		fieldMap[field] = true
	}
	
	result := make([]string, 0, len(fieldMap))
	for field := range fieldMap {
		result = append(result, field)
	}
	
	return result
}

// extractFunctions 提取表达式中使用的函数
func (v *SyntaxValidator) extractFunctions(expression string) []string {
	functionPattern := regexp.MustCompile(`([A-Z][a-zA-Z0-9_]*)\s*\(`)
	matches := functionPattern.FindAllStringSubmatch(expression, -1)
	
	// 去重
	functionMap := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			functionMap[match[1]] = true
		}
	}
	
	result := make([]string, 0, len(functionMap))
	for function := range functionMap {
		result = append(result, function)
	}
	
	return result
}

// generateSuggestions 生成修复建议
func (v *SyntaxValidator) generateSuggestions(expression string) []string {
	suggestions := []string{}

	// 检查常见错误并给出建议
	if strings.Contains(expression, "//") {
		suggestions = append(suggestions, "将 // 替换为 / 进行除法运算")
	}

	if strings.Contains(expression, "；") {
		suggestions = append(suggestions, "将中文分号 ； 替换为英文分号 ;")
	}

	if !v.checkParenthesesBalance(expression) {
		suggestions = append(suggestions, "检查括号是否匹配")
	}

	// 检查字段使用
	fieldPattern := regexp.MustCompile(`\$[a-zA-Z_][a-zA-Z0-9_]*`)
	fields := fieldPattern.FindAllString(expression, -1)
	validFields := map[string]bool{
		"$open": true, "$high": true, "$low": true, "$close": true,
		"$volume": true, "$factor": true, "$vwap": true, "$amount": true,
	}

	for _, field := range fields {
		if !validFields[field] {
			suggestions = append(suggestions, fmt.Sprintf("字段 %s 可能不存在，请检查拼写", field))
		}
	}

	// 检查函数使用
	functionPattern := regexp.MustCompile(`([A-Z][a-zA-Z0-9_]*)\s*\(`)
	matches := functionPattern.FindAllStringSubmatch(expression, -1)
	validFunctions := map[string]bool{
		"Mean": true, "Std": true, "Corr": true, "Rank": true,
		"Ref": true, "Delta": true, "Sum": true, "Max": true,
		"Min": true, "Abs": true, "Log": true, "Sqrt": true,
	}

	for _, match := range matches {
		if len(match) > 1 {
			function := match[1]
			if !validFunctions[function] {
				suggestions = append(suggestions, fmt.Sprintf("函数 %s 可能不存在，请检查拼写或查阅文档", function))
			}
		}
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "请检查表达式语法是否正确")
	}

	return suggestions
}

// executePythonScript 执行Python脚本进行语法验证
func (v *SyntaxValidator) executePythonScript(args map[string]interface{}) (map[string]interface{}, error) {
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("序列化参数失败: %v", err)
	}

	pythonScript := `
import json
import sys
import re
import ast

def validate_syntax(expression):
    """验证Qlib因子表达式语法"""
    try:
        result = {
            "is_valid": True,
            "error_msg": "",
            "parsed_ast": ""
        }
        
        # 基础检查
        if not expression or len(expression.strip()) == 0:
            result["is_valid"] = False
            result["error_msg"] = "表达式不能为空"
            return result
            
        # 检查括号匹配
        if not check_parentheses_balance(expression):
            result["is_valid"] = False
            result["error_msg"] = "括号不匹配"
            return result
            
        # 检查基本语法
        syntax_error = check_basic_syntax(expression)
        if syntax_error:
            result["is_valid"] = False
            result["error_msg"] = syntax_error
            return result
            
        # 尝试解析表达式结构
        try:
            parsed_info = parse_expression_structure(expression)
            result["parsed_ast"] = json.dumps(parsed_info)
        except Exception as e:
            result["parsed_ast"] = f"解析失败: {str(e)}"
            
        return result
        
    except Exception as e:
        return {
            "is_valid": False,
            "error_msg": str(e),
            "parsed_ast": ""
        }

def check_parentheses_balance(expression):
    """检查括号平衡"""
    balance = 0
    for char in expression:
        if char == '(':
            balance += 1
        elif char == ')':
            balance -= 1
            if balance < 0:
                return False
    return balance == 0

def check_basic_syntax(expression):
    """检查基本语法"""
    # 检查无效字符
    invalid_chars = ['；', '、', '"', '"', ''', ''']
    for char in invalid_chars:
        if char in expression:
            return f"包含无效字符: {char}"
    
    # 检查字段格式
    field_pattern = re.compile(r'\$[a-zA-Z_][a-zA-Z0-9_]*')
    fields = field_pattern.findall(expression)
    valid_fields = {
        '$open', '$high', '$low', '$close', '$volume', 
        '$factor', '$vwap', '$amount', '$pctchange', '$adjclose'
    }
    
    for field in fields:
        if field not in valid_fields:
            return f"未知字段: {field}"
    
    # 检查函数格式
    function_pattern = re.compile(r'([A-Z][a-zA-Z0-9_]*)\s*\(')
    functions = [match for match in function_pattern.findall(expression)]
    valid_functions = {
        'Mean', 'Std', 'Corr', 'Rank', 'Ref', 'Delta', 'Sum', 
        'Max', 'Min', 'Abs', 'Log', 'Sqrt', 'Quantile', 'WMA'
    }
    
    for func in functions:
        if func not in valid_functions:
            return f"未知函数: {func}"
    
    return None

def parse_expression_structure(expression):
    """解析表达式结构"""
    structure = {
        "fields": [],
        "functions": [],
        "operators": [],
        "complexity": "simple"
    }
    
    # 提取字段
    field_pattern = re.compile(r'\$[a-zA-Z_][a-zA-Z0-9_]*')
    structure["fields"] = list(set(field_pattern.findall(expression)))
    
    # 提取函数
    function_pattern = re.compile(r'([A-Z][a-zA-Z0-9_]*)\s*\(')
    structure["functions"] = list(set(function_pattern.findall(expression)))
    
    # 提取操作符
    operator_pattern = re.compile(r'[+\-*/()><>=!]')
    structure["operators"] = list(set(operator_pattern.findall(expression)))
    
    # 评估复杂度
    if len(structure["functions"]) > 2 or len(structure["fields"]) > 3:
        structure["complexity"] = "complex"
    elif len(structure["functions"]) > 0:
        structure["complexity"] = "medium"
    
    return structure

def main():
    try:
        args_json = sys.stdin.read()
        args = json.loads(args_json)
        action = args.get('action')
        
        result = {"success": True, "data": None, "error": None}
        
        if action == "validate_syntax":
            expression = args.get('expression')
            validation_result = validate_syntax(expression)
            result["data"] = validation_result
        else:
            result["success"] = False
            result["error"] = f"Unknown action: {action}"
            
        print(json.dumps(result))
        
    except Exception as e:
        error_result = {
            "success": False,
            "error": str(e),
            "data": None
        }
        print(json.dumps(error_result))

if __name__ == "__main__":
    main()
`

	cmd := exec.Command(v.pythonPath, "-c", pythonScript)
	cmd.Stdin = strings.NewReader(string(argsJSON))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行Python脚本失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("解析Python输出失败: %v", err)
	}

	if success, ok := result["success"].(bool); !ok || !success {
		if errorMsg, ok := result["error"].(string); ok {
			return nil, fmt.Errorf("Python脚本执行失败: %s", errorMsg)
		}
		return nil, fmt.Errorf("Python脚本执行失败")
	}

	return result, nil
}

// ValidateAndSuggest 验证表达式并提供智能建议
func (v *SyntaxValidator) ValidateAndSuggest(expression string) (*ValidationResult, error) {
	result, err := v.Validate(expression)
	if err != nil {
		return nil, err
	}

	// 如果验证失败，尝试提供智能修复建议
	if !result.IsValid {
		result.Suggestions = append(result.Suggestions, v.generateIntelligentSuggestions(expression)...)
	}

	return result, nil
}

// generateIntelligentSuggestions 生成智能修复建议
func (v *SyntaxValidator) generateIntelligentSuggestions(expression string) []string {
	suggestions := []string{}

	// 常见错误修复建议
	commonFixes := map[string]string{
		"close":  "$close",
		"open":   "$open",
		"high":   "$high",
		"low":    "$low",
		"volume": "$volume",
		"mean":   "Mean",
		"std":    "Std",
		"rank":   "Rank",
	}

	for wrong, correct := range commonFixes {
		if strings.Contains(strings.ToLower(expression), wrong) && !strings.Contains(expression, correct) {
			suggestions = append(suggestions, fmt.Sprintf("是否想要使用 %s 代替 %s？", correct, wrong))
		}
	}

	// 函数参数建议
	if strings.Contains(expression, "Mean(") && !regexp.MustCompile(`Mean\([^,]+,\s*\d+\)`).MatchString(expression) {
		suggestions = append(suggestions, "Mean函数需要两个参数：Mean(data, window)")
	}

	if strings.Contains(expression, "Ref(") && !regexp.MustCompile(`Ref\([^,]+,\s*\d+\)`).MatchString(expression) {
		suggestions = append(suggestions, "Ref函数需要两个参数：Ref(data, period)")
	}

	return suggestions
}