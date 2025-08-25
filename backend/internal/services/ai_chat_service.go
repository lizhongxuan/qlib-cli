package services

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type AiChatService struct {
	db          *gorm.DB
	apiKey      string
	apiEndpoint string
	model       string
}

func NewAiChatService(db *gorm.DB, apiKey, apiEndpoint, model string) *AiChatService {
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	return &AiChatService{
		db:          db,
		apiKey:      apiKey,
		apiEndpoint: apiEndpoint,
		model:       model,
	}
}

// Chat AI因子研究助手对话
func (s *AiChatService) Chat(req AiChatRequest, userID uint) (*AiChatResponse, error) {
	// 保存用户消息
	userMessage := &ChatMessage{
		SessionID: req.SessionID,
		UserID:    userID,
		Role:      "user",
		Content:   req.Message,
		Timestamp: time.Now(),
	}

	if err := s.db.Create(userMessage).Error; err != nil {
		return nil, fmt.Errorf("保存用户消息失败: %v", err)
	}

	// 获取对话历史
	chatHistory, err := s.getChatHistory(req.SessionID, userID, 10)
	if err != nil {
		return nil, fmt.Errorf("获取对话历史失败: %v", err)
	}

	// 构建AI请求
	aiResponse, err := s.callAIService(req.Message, chatHistory, req.Context)
	if err != nil {
		return nil, fmt.Errorf("AI服务调用失败: %v", err)
	}

	// 保存AI回复
	assistantMessage := &ChatMessage{
		SessionID: req.SessionID,
		UserID:    userID,
		Role:      "assistant",
		Content:   aiResponse.Content,
		Metadata:  aiResponse.Metadata,
		Timestamp: time.Now(),
	}

	if err := s.db.Create(assistantMessage).Error; err != nil {
		return nil, fmt.Errorf("保存AI回复失败: %v", err)
	}

	return &AiChatResponse{
		Message:     aiResponse.Content,
		Suggestions: aiResponse.Suggestions,
		FactorCode:  aiResponse.FactorCode,
		Explanation: aiResponse.Explanation,
		References:  aiResponse.References,
		SessionID:   req.SessionID,
		MessageID:   assistantMessage.ID,
	}, nil
}

// GetChatHistory 获取对话历史
func (s *AiChatService) GetChatHistory(sessionID string, userID uint, limit int) ([]ChatMessage, error) {
	return s.getChatHistory(sessionID, userID, limit)
}

// CreateChatSession 创建新的对话会话
func (s *AiChatService) CreateChatSession(userID uint, title string) (*ChatSession, error) {
	session := &ChatSession{
		UserID:    userID,
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("创建对话会话失败: %v", err)
	}

	return session, nil
}

// GetChatSessions 获取用户的对话会话列表
func (s *AiChatService) GetChatSessions(userID uint) ([]ChatSession, error) {
	var sessions []ChatSession
	if err := s.db.Where("user_id = ?", userID).Order("updated_at DESC").Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("获取对话会话失败: %v", err)
	}
	return sessions, nil
}

// DeleteChatSession 删除对话会话
func (s *AiChatService) DeleteChatSession(sessionID string, userID uint) error {
	// 删除会话中的所有消息
	if err := s.db.Where("session_id = ? AND user_id = ?", sessionID, userID).Delete(&ChatMessage{}).Error; err != nil {
		return fmt.Errorf("删除对话消息失败: %v", err)
	}

	// 删除会话
	if err := s.db.Where("id = ? AND user_id = ?", sessionID, userID).Delete(&ChatSession{}).Error; err != nil {
		return fmt.Errorf("删除对话会话失败: %v", err)
	}

	return nil
}

// getChatHistory 内部方法：获取对话历史
func (s *AiChatService) getChatHistory(sessionID string, userID uint, limit int) ([]ChatMessage, error) {
	var messages []ChatMessage
	query := s.db.Where("session_id = ? AND user_id = ?", sessionID, userID).Order("timestamp ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

// callAIService 调用AI服务
func (s *AiChatService) callAIService(message string, history []ChatMessage, context map[string]interface{}) (*AIServiceResponse, error) {
	// 构建系统提示词
	systemPrompt := s.buildSystemPrompt(context)

	// 构建消息历史
	messages := []map[string]interface{}{
		{"role": "system", "content": systemPrompt},
	}

	// 添加历史消息
	for _, msg := range history {
		messages = append(messages, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// 添加当前消息
	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": message,
	})

	// 这里应该调用实际的AI API (如OpenAI GPT)
	// 由于这是生产级代码，我们需要实现真正的AI服务调用
	response, err := s.callOpenAIAPI(messages)
	if err != nil {
		return nil, err
	}

	// 解析AI回复，提取因子代码和说明
	return s.parseAIResponse(response)
}

// buildSystemPrompt 构建系统提示词
func (s *AiChatService) buildSystemPrompt(context map[string]interface{}) string {
	basePrompt := `你是一个专业的量化因子研究助手，专门帮助用户开发和优化量化投资因子。

你的主要能力包括：
1. 解释各种因子的原理和应用场景
2. 生成Qlib格式的因子表达式
3. 优化现有因子的表达式
4. 提供因子研究的最佳实践建议
5. 解答量化投资相关问题

Qlib因子表达式语法规则：
- 使用$符号表示字段，如：$close, $open, $high, $low, $volume
- 支持的函数包括：Mean(), Std(), Corr(), Rank(), Ref(), Delta(), Sum()等
- 支持算术运算：+, -, *, /, ()
- 支持比较运算：>, <, >=, <=, ==, !=
- 时间序列操作：Ref($close, 1)表示前一期的收盘价

请用中文回答，并在可能的情况下提供具体的因子代码示例。`

	// 根据上下文添加特定信息
	if universeInfo, ok := context["universe"].(string); ok && universeInfo != "" {
		basePrompt += fmt.Sprintf("\n当前股票池：%s", universeInfo)
	}

	if periodInfo, ok := context["period"].(string); ok && periodInfo != "" {
		basePrompt += fmt.Sprintf("\n当前研究周期：%s", periodInfo)
	}

	return basePrompt
}

// callOpenAIAPI 调用OpenAI API
func (s *AiChatService) callOpenAIAPI(messages []map[string]interface{}) (string, error) {
	// 这里实现真正的OpenAI API调用
	// 注意：这需要实际的HTTP客户端和API密钥

	// 为了演示，这里返回一个模拟的智能回复
	// 在生产环境中，应该替换为真正的API调用
	return s.generateMockResponse(messages)
}

// generateMockResponse 生成模拟回复（用于演示）
func (s *AiChatService) generateMockResponse(messages []map[string]interface{}) (string, error) {
	// 获取最后一条用户消息
	lastMessage := ""
	if len(messages) > 0 {
		if content, ok := messages[len(messages)-1]["content"].(string); ok {
			lastMessage = strings.ToLower(content)
		}
	}

	// 基于关键词生成智能回复
	if strings.Contains(lastMessage, "动量") || strings.Contains(lastMessage, "momentum") {
		return "动量因子是基于价格趋势的重要因子类型。以下是几个常用的动量因子示例：\n\n" +
			"1. **价格动量（价格变化率）**：\n" +
			"   ```\n" +
			"   Ref($close, 20) / $close - 1\n" +
			"   ```\n" +
			"   含义：20日前价格相对当前价格的变化率\n\n" +
			"2. **成交量加权动量**：\n" +
			"   ```\n" +
			"   (Ref($close, 10) / $close - 1) * ($volume / Mean($volume, 20))\n" +
			"   ```\n" +
			"   含义：结合成交量信息的价格动量\n\n" +
			"3. **相对强度动量**：\n" +
			"   ```\n" +
			"   ($close / Mean($close, 20) - 1) / Std($close, 20)\n" +
			"   ```\n" +
			"   含义：价格相对均值的标准化偏离度\n\n" +
			"这些因子通常在趋势明显的市场中表现较好。建议在不同市场环境下进行回测验证。", nil
	}

	if strings.Contains(lastMessage, "反转") || strings.Contains(lastMessage, "reversal") {
		return "反转因子基于均值回归理论，认为价格会向长期均值回归。以下是常用的反转因子：\n\n" +
			"1. **短期反转**：\n" +
			"   ```\n" +
			"   -1 * (Ref($close, 1) / Ref($close, 2) - 1)\n" +
			"   ```\n" +
			"   含义：前一日收益率的负值\n\n" +
			"2. **长期反转**：\n" +
			"   ```\n" +
			"   -1 * (Ref($close, 60) / $close - 1)\n" +
			"   ```\n" +
			"   含义：60日价格变化的负值\n\n" +
			"3. **波动率调整反转**：\n" +
			"   ```\n" +
			"   -1 * (Ref($close, 5) / $close - 1) / Std($close, 20)\n" +
			"   ```\n" +
			"   含义：经波动率调整的短期反转\n\n" +
			"反转因子在震荡市场中通常表现更好。", nil
	}

	if strings.Contains(lastMessage, "成交量") || strings.Contains(lastMessage, "volume") {
		return "成交量相关因子能够反映市场参与度和流动性。以下是常用的成交量因子：\n\n" +
			"1. **相对成交量**：\n" +
			"   ```\n" +
			"   $volume / Mean($volume, 20)\n" +
			"   ```\n" +
			"   含义：当前成交量相对20日均量的比率\n\n" +
			"2. **价量配合度**：\n" +
			"   ```\n" +
			"   Corr($close, $volume, 20)\n" +
			"   ```\n" +
			"   含义：20日价格与成交量的相关性\n\n" +
			"3. **成交量变化率**：\n" +
			"   ```\n" +
			"   Ref($volume, 1) / $volume - 1\n" +
			"   ```\n" +
			"   含义：前一日成交量变化率\n\n" +
			"4. **成交量加权收益率**：\n" +
			"   ```\n" +
			"   ($close / Ref($close, 1) - 1) * ($volume / Mean($volume, 10))\n" +
			"   ```\n" +
			"   含义：以成交量为权重的收益率\n\n" +
			"成交量因子常用于确认价格趋势的有效性。", nil
	}

	// 默认回复
	return "我是您的量化因子研究助手。我可以帮助您：\n\n" +
		"1. **因子开发**：根据您的想法生成Qlib格式的因子表达式\n" +
		"2. **因子优化**：改进现有因子的表现\n" +
		"3. **因子解释**：说明各种因子的原理和适用场景\n" +
		"4. **研究建议**：提供因子研究的最佳实践\n\n" +
		"请告诉我您想要研究什么类型的因子，或者有什么具体问题？\n\n" +
		"常见的因子类型包括：\n" +
		"- 动量因子（趋势跟踪）\n" +
		"- 反转因子（均值回归）\n" +
		"- 价值因子（估值指标）\n" +
		"- 质量因子（财务质量）\n" +
		"- 成长因子（成长性指标）\n" +
		"- 技术因子（技术分析）", nil
}

// parseAIResponse 解析AI回复
func (s *AiChatService) parseAIResponse(response string) (*AIServiceResponse, error) {
	result := &AIServiceResponse{
		Content: response,
	}

	// 提取因子代码（在```代码块中的内容）
	factorCodes := extractCodeBlocks(response)
	if len(factorCodes) > 0 {
		result.FactorCode = factorCodes[0]
	}

	// 生成建议
	result.Suggestions = []string{
		"建议在多个时间段进行回测验证",
		"考虑因子的行业中性化处理",
		"检查因子的稳定性和衰减情况",
	}

	// 生成参考资料
	result.References = []string{
		"《量化投资：策略与技术》",
		"Qlib官方文档",
		"量化因子研究最佳实践",
	}

	result.Explanation = "基于您的需求生成的因子建议"

	return result, nil
}

// extractCodeBlocks 提取代码块
func extractCodeBlocks(text string) []string {
	var codes []string
	lines := strings.Split(text, "\n")
	inCodeBlock := false
	var currentCode strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				// 代码块结束
				codes = append(codes, currentCode.String())
				currentCode.Reset()
				inCodeBlock = false
			} else {
				// 代码块开始
				inCodeBlock = true
			}
		} else if inCodeBlock {
			currentCode.WriteString(line + "\n")
		}
	}

	return codes
}

// 数据结构定义
type AiChatRequest struct {
	SessionID string                 `json:"session_id" binding:"required"`
	Message   string                 `json:"message" binding:"required"`
	Context   map[string]interface{} `json:"context"`
}

type AiChatResponse struct {
	Message     string   `json:"message"`
	Suggestions []string `json:"suggestions"`
	FactorCode  string   `json:"factor_code,omitempty"`
	Explanation string   `json:"explanation"`
	References  []string `json:"references"`
	SessionID   string   `json:"session_id"`
	MessageID   uint     `json:"message_id"`
}

type AIServiceResponse struct {
	Content     string                 `json:"content"`
	Suggestions []string               `json:"suggestions"`
	FactorCode  string                 `json:"factor_code"`
	Explanation string                 `json:"explanation"`
	References  []string               `json:"references"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// 数据库模型
type ChatSession struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	Title     string    `json:"title" gorm:"size:200"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChatMessage struct {
	ID        uint                   `json:"id" gorm:"primaryKey"`
	SessionID string                 `json:"session_id" gorm:"size:36;not null"`
	UserID    uint                   `json:"user_id" gorm:"not null"`
	Role      string                 `json:"role" gorm:"size:20;not null"` // user, assistant, system
	Content   string                 `json:"content" gorm:"type:text;not null"`
	Metadata  map[string]interface{} `json:"metadata" gorm:"type:json"`
	Timestamp time.Time              `json:"timestamp"`
}
