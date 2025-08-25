package services

import (
	"encoding/json"
	"fmt"
	"time"

	"qlib-backend/internal/models"

	"gorm.io/gorm"
)

// UIConfigService 界面配置服务
type UIConfigService struct {
	db *gorm.DB
}

// LayoutConfig 布局配置
type LayoutConfig struct {
	UserID      uint                   `json:"user_id"`
	ConfigType  string                 `json:"config_type"`  // default, custom
	Platform    string                 `json:"platform"`     // web, mobile, tablet
	Theme       string                 `json:"theme"`        // light, dark
	Layout      *LayoutStructure       `json:"layout"`
	Widgets     []WidgetConfig         `json:"widgets"`
	Settings    *UISettings            `json:"settings"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// LayoutStructure 布局结构
type LayoutStructure struct {
	Type        string           `json:"type"`         // grid, flex, absolute
	Columns     int              `json:"columns"`      // 网格列数
	Rows        int              `json:"rows"`         // 网格行数
	Areas       []LayoutArea     `json:"areas"`        // 布局区域
	Responsive  bool             `json:"responsive"`   // 是否响应式
	Breakpoints map[string]int   `json:"breakpoints"`  // 断点配置
}

// LayoutArea 布局区域
type LayoutArea struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Position AreaPosition      `json:"position"`
	Size     AreaSize          `json:"size"`
	Style    map[string]string `json:"style"`
	Widget   string            `json:"widget"`    // 绑定的组件ID
}

// AreaPosition 区域位置
type AreaPosition struct {
	X      int `json:"x"`       // 起始列
	Y      int `json:"y"`       // 起始行
	ZIndex int `json:"z_index"` // 层级
}

// AreaSize 区域大小
type AreaSize struct {
	Width     int    `json:"width"`      // 列数/像素
	Height    int    `json:"height"`     // 行数/像素
	MinWidth  int    `json:"min_width"`  // 最小宽度
	MinHeight int    `json:"min_height"` // 最小高度
	MaxWidth  int    `json:"max_width"`  // 最大宽度
	MaxHeight int    `json:"max_height"` // 最大高度
	Unit      string `json:"unit"`       // px, %, grid
}

// WidgetConfig 组件配置
type WidgetConfig struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`         // chart, table, card, form
	Title       string                 `json:"title"`
	Area        string                 `json:"area"`         // 绑定的布局区域ID
	Visible     bool                   `json:"visible"`
	Resizable   bool                   `json:"resizable"`
	Draggable   bool                   `json:"draggable"`
	Config      map[string]interface{} `json:"config"`       // 组件特定配置
	DataSource  string                 `json:"data_source"`  // 数据源
	Refresh     *RefreshConfig         `json:"refresh"`      // 刷新配置
	Permissions []string               `json:"permissions"`  // 权限要求
}

// RefreshConfig 刷新配置
type RefreshConfig struct {
	Auto     bool `json:"auto"`     // 自动刷新
	Interval int  `json:"interval"` // 刷新间隔(秒)
	Manual   bool `json:"manual"`   // 手动刷新
}

// UISettings 界面设置
type UISettings struct {
	Theme           string            `json:"theme"`            // 主题
	Language        string            `json:"language"`         // 语言
	FontSize        string            `json:"font_size"`        // 字体大小
	Sidebar         *SidebarConfig    `json:"sidebar"`          // 侧边栏配置
	Header          *HeaderConfig     `json:"header"`           // 头部配置
	Footer          *FooterConfig     `json:"footer"`           // 底部配置
	Animation       bool              `json:"animation"`        // 动画效果
	Sound           bool              `json:"sound"`            // 声音提示
	Notifications   *NotifySettings   `json:"notifications"`    // 通知设置
	Shortcuts       map[string]string `json:"shortcuts"`        // 快捷键
	AutoSave        bool              `json:"auto_save"`        // 自动保存
	SaveInterval    int               `json:"save_interval"`    // 保存间隔(秒)
}

// SidebarConfig 侧边栏配置
type SidebarConfig struct {
	Visible   bool   `json:"visible"`
	Collapsed bool   `json:"collapsed"`
	Position  string `json:"position"` // left, right
	Width     int    `json:"width"`
}

// HeaderConfig 头部配置
type HeaderConfig struct {
	Visible bool     `json:"visible"`
	Height  int      `json:"height"`
	Items   []string `json:"items"` // 显示的项目
}

// FooterConfig 底部配置
type FooterConfig struct {
	Visible bool     `json:"visible"`
	Height  int      `json:"height"`
	Items   []string `json:"items"` // 显示的项目
}

// NotifySettings 通知设置
type NotifySettings struct {
	Desktop  bool `json:"desktop"`  // 桌面通知
	Sound    bool `json:"sound"`    // 声音通知
	Popup    bool `json:"popup"`    // 弹窗通知
	Email    bool `json:"email"`    // 邮件通知
	Position string `json:"position"` // 通知位置
}

// UIConfigTemplate 界面配置模板
type UIConfigTemplate struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Category    string        `json:"category"`    // dashboard, analysis, research
	Platform    string        `json:"platform"`    // web, mobile, tablet
	Config      *LayoutConfig `json:"config"`
	Preview     string        `json:"preview"`     // 预览图URL
	Popular     bool          `json:"popular"`     // 是否热门
	CreatedAt   time.Time     `json:"created_at"`
}

// NewUIConfigService 创建新的界面配置服务
func NewUIConfigService(db *gorm.DB) *UIConfigService {
	return &UIConfigService{
		db: db,
	}
}

// GetLayoutConfig 获取布局配置
func (ucs *UIConfigService) GetLayoutConfig(userID uint, configType, platform, theme string) (*LayoutConfig, error) {
	// 先尝试获取用户自定义配置
	var config models.UIConfig
	err := ucs.db.Where("user_id = ? AND config_type = ? AND platform = ?", userID, configType, platform).First(&config).Error
	
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("获取用户配置失败: %v", err)
	}
	
	// 如果没有找到用户配置，返回默认配置
	if err == gorm.ErrRecordNotFound {
		return ucs.getDefaultLayoutConfig(platform, theme), nil
	}
	
	// 解析用户配置
	var layoutConfig LayoutConfig
	if err := json.Unmarshal([]byte(config.ConfigData), &layoutConfig); err != nil {
		return nil, fmt.Errorf("解析用户配置失败: %v", err)
	}
	
	layoutConfig.UserID = userID
	layoutConfig.UpdatedAt = config.UpdatedAt
	
	return &layoutConfig, nil
}

// SaveLayoutConfig 保存布局配置
func (ucs *UIConfigService) SaveLayoutConfig(userID uint, config *LayoutConfig) error {
	// 序列化配置
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}
	
	// 保存或更新配置
	uiConfig := models.UIConfig{
		UserID:     userID,
		ConfigType: config.ConfigType,
		Platform:   config.Platform,
		ConfigData: string(configJSON),
	}
	
	// 尝试更新现有配置
	result := ucs.db.Where("user_id = ? AND config_type = ? AND platform = ?", userID, config.ConfigType, config.Platform).Updates(&uiConfig)
	
	// 如果没有更新任何记录，则创建新记录
	if result.RowsAffected == 0 {
		if err := ucs.db.Create(&uiConfig).Error; err != nil {
			return fmt.Errorf("创建配置失败: %v", err)
		}
	}
	
	return nil
}

// GetConfigTemplates 获取配置模板
func (ucs *UIConfigService) GetConfigTemplates(platform, category string) ([]UIConfigTemplate, error) {
	templates := ucs.getBuiltinTemplates()
	
	// 过滤模板
	var filtered []UIConfigTemplate
	for _, template := range templates {
		if platform != "" && template.Platform != platform {
			continue
		}
		if category != "" && template.Category != category {
			continue
		}
		filtered = append(filtered, template)
	}
	
	return filtered, nil
}

// getDefaultLayoutConfig 获取默认布局配置
func (ucs *UIConfigService) getDefaultLayoutConfig(platform, theme string) *LayoutConfig {
	config := &LayoutConfig{
		ConfigType: "default",
		Platform:   platform,
		Theme:      theme,
		Layout: &LayoutStructure{
			Type:    "grid",
			Columns: 12,
			Rows:    8,
			Areas: []LayoutArea{
				{
					ID:   "header",
					Name: "头部区域",
					Position: AreaPosition{X: 0, Y: 0, ZIndex: 100},
					Size:     AreaSize{Width: 12, Height: 1, Unit: "grid"},
					Widget:   "header",
				},
				{
					ID:   "sidebar",
					Name: "侧边栏",
					Position: AreaPosition{X: 0, Y: 1, ZIndex: 10},
					Size:     AreaSize{Width: 2, Height: 6, Unit: "grid"},
					Widget:   "navigation",
				},
				{
					ID:   "main",
					Name: "主内容区",
					Position: AreaPosition{X: 2, Y: 1, ZIndex: 1},
					Size:     AreaSize{Width: 10, Height: 6, Unit: "grid"},
					Widget:   "content",
				},
				{
					ID:   "footer",
					Name: "底部区域",
					Position: AreaPosition{X: 0, Y: 7, ZIndex: 100},
					Size:     AreaSize{Width: 12, Height: 1, Unit: "grid"},
					Widget:   "footer",
				},
			},
			Responsive: true,
			Breakpoints: map[string]int{
				"mobile": 768,
				"tablet": 1024,
				"desktop": 1200,
			},
		},
		Widgets: []WidgetConfig{
			{
				ID:       "header",
				Type:     "header",
				Title:    "顶部导航",
				Area:     "header",
				Visible:  true,
				Config:   map[string]interface{}{"showLogo": true, "showUser": true},
			},
			{
				ID:       "navigation",
				Type:     "navigation",
				Title:    "导航菜单",
				Area:     "sidebar",
				Visible:  true,
				Config:   map[string]interface{}{"collapsed": false},
			},
			{
				ID:       "content",
				Type:     "content",
				Title:    "主内容",
				Area:     "main",
				Visible:  true,
				Config:   map[string]interface{}{"padding": "20px"},
			},
			{
				ID:       "footer",
				Type:     "footer",
				Title:    "底部信息",
				Area:     "footer",
				Visible:  true,
				Config:   map[string]interface{}{"showCopyright": true},
			},
		},
		Settings: &UISettings{
			Theme:     theme,
			Language:  "zh",
			FontSize:  "medium",
			Animation: true,
			Sound:     false,
			Sidebar: &SidebarConfig{
				Visible:   true,
				Collapsed: false,
				Position:  "left",
				Width:     240,
			},
			Header: &HeaderConfig{
				Visible: true,
				Height:  60,
				Items:   []string{"logo", "navigation", "user", "notifications"},
			},
			Footer: &FooterConfig{
				Visible: true,
				Height:  40,
				Items:   []string{"copyright", "version"},
			},
			Notifications: &NotifySettings{
				Desktop:  true,
				Sound:    false,
				Popup:    true,
				Email:    false,
				Position: "top-right",
			},
			AutoSave:     true,
			SaveInterval: 300, // 5分钟
		},
		UpdatedAt: time.Now(),
	}
	
	return config
}

// getBuiltinTemplates 获取内置模板
func (ucs *UIConfigService) getBuiltinTemplates() []UIConfigTemplate {
	return []UIConfigTemplate{
		{
			ID:          "dashboard_default",
			Name:        "默认仪表板",
			Description: "适合数据展示和监控的默认布局",
			Category:    "dashboard",
			Platform:    "web",
			Config:      ucs.getDefaultLayoutConfig("web", "light"),
			Popular:     true,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "analysis_workspace",
			Name:        "分析工作台",
			Description: "专为数据分析设计的布局模板",
			Category:    "analysis",
			Platform:    "web",
			Config:      ucs.getAnalysisLayoutConfig(),
			Popular:     true,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "research_lab",
			Name:        "研究实验室",
			Description: "适合因子研究和策略开发的布局",
			Category:    "research",
			Platform:    "web",
			Config:      ucs.getResearchLayoutConfig(),
			Popular:     false,
			CreatedAt:   time.Now(),
		},
	}
}

// getAnalysisLayoutConfig 获取分析布局配置
func (ucs *UIConfigService) getAnalysisLayoutConfig() *LayoutConfig {
	config := ucs.getDefaultLayoutConfig("web", "light")
	config.ConfigType = "analysis"
	
	// 调整布局区域
	config.Layout.Areas = []LayoutArea{
		{
			ID:   "header",
			Name: "头部工具栏",
			Position: AreaPosition{X: 0, Y: 0, ZIndex: 100},
			Size:     AreaSize{Width: 12, Height: 1, Unit: "grid"},
			Widget:   "analysis_header",
		},
		{
			ID:   "sidebar",
			Name: "工具面板",
			Position: AreaPosition{X: 0, Y: 1, ZIndex: 10},
			Size:     AreaSize{Width: 3, Height: 7, Unit: "grid"},
			Widget:   "tools_panel",
		},
		{
			ID:   "chart_area",
			Name: "图表区域",
			Position: AreaPosition{X: 3, Y: 1, ZIndex: 1},
			Size:     AreaSize{Width: 6, Height: 4, Unit: "grid"},
			Widget:   "analysis_chart",
		},
		{
			ID:   "data_table",
			Name: "数据表格",
			Position: AreaPosition{X: 3, Y: 5, ZIndex: 1},
			Size:     AreaSize{Width: 6, Height: 3, Unit: "grid"},
			Widget:   "data_grid",
		},
		{
			ID:   "properties",
			Name: "属性面板",
			Position: AreaPosition{X: 9, Y: 1, ZIndex: 10},
			Size:     AreaSize{Width: 3, Height: 7, Unit: "grid"},
			Widget:   "properties_panel",
		},
	}
	
	return config
}

// getResearchLayoutConfig 获取研究布局配置
func (ucs *UIConfigService) getResearchLayoutConfig() *LayoutConfig {
	config := ucs.getDefaultLayoutConfig("web", "dark")
	config.ConfigType = "research"
	config.Theme = "dark"
	
	// 调整为研究模式布局
	config.Layout.Areas = []LayoutArea{
		{
			ID:   "header",
			Name: "研究工具栏",
			Position: AreaPosition{X: 0, Y: 0, ZIndex: 100},
			Size:     AreaSize{Width: 12, Height: 1, Unit: "grid"},
			Widget:   "research_header",
		},
		{
			ID:   "factor_tree",
			Name: "因子树",
			Position: AreaPosition{X: 0, Y: 1, ZIndex: 10},
			Size:     AreaSize{Width: 2, Height: 7, Unit: "grid"},
			Widget:   "factor_tree",
		},
		{
			ID:   "editor",
			Name: "代码编辑器",
			Position: AreaPosition{X: 2, Y: 1, ZIndex: 1},
			Size:     AreaSize{Width: 5, Height: 4, Unit: "grid"},
			Widget:   "code_editor",
		},
		{
			ID:   "preview",
			Name: "预览窗口",
			Position: AreaPosition{X: 7, Y: 1, ZIndex: 1},
			Size:     AreaSize{Width: 5, Height: 4, Unit: "grid"},
			Widget:   "result_preview",
		},
		{
			ID:   "console",
			Name: "控制台",
			Position: AreaPosition{X: 2, Y: 5, ZIndex: 1},
			Size:     AreaSize{Width: 10, Height: 3, Unit: "grid"},
			Widget:   "console",
		},
	}
	
	return config
}