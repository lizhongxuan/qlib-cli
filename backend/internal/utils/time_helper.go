package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DateFormat 常用日期格式
const (
	DateFormat      = "2006-01-02"
	TimeFormat      = "15:04:05"
	DateTimeFormat  = "2006-01-02 15:04:05"
	ISO8601Format   = "2006-01-02T15:04:05Z07:00"
	ShortDateFormat = "01/02/2006"
	ChineseDateFormat = "2006年01月02日"
)

// TimeHelper 时间工具
type TimeHelper struct{}

// NewTimeHelper 创建时间工具
func NewTimeHelper() *TimeHelper {
	return &TimeHelper{}
}

// ParseDate 解析日期字符串
func (th *TimeHelper) ParseDate(dateStr string) (time.Time, error) {
	// 尝试多种格式
	formats := []string{
		DateFormat,
		DateTimeFormat,
		ISO8601Format,
		ShortDateFormat,
		ChineseDateFormat,
		time.RFC3339,
		"2006/01/02",
		"2006-1-2",
		"2006年1月2日",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析日期格式: %s", dateStr)
}

// FormatDate 格式化日期
func (th *TimeHelper) FormatDate(t time.Time, format string) string {
	if format == "" {
		format = DateFormat
	}
	return t.Format(format)
}

// GetDateRange 获取日期范围
func (th *TimeHelper) GetDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	start, err := th.ParseDate(startDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("解析开始日期失败: %w", err)
	}

	end, err := th.ParseDate(endDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("解析结束日期失败: %w", err)
	}

	if start.After(end) {
		return time.Time{}, time.Time{}, fmt.Errorf("开始日期不能晚于结束日期")
	}

	return start, end, nil
}

// GetTradingDays 获取交易日列表（排除周末和节假日）
func (th *TimeHelper) GetTradingDays(startDate, endDate time.Time) []time.Time {
	var tradingDays []time.Time
	
	current := startDate
	for !current.After(endDate) {
		// 排除周末
		if current.Weekday() != time.Saturday && current.Weekday() != time.Sunday {
			// 这里可以进一步排除节假日
			if !th.isHoliday(current) {
				tradingDays = append(tradingDays, current)
			}
		}
		current = current.AddDate(0, 0, 1)
	}

	return tradingDays
}

// isHoliday 检查是否为节假日（简化版本）
func (th *TimeHelper) isHoliday(date time.Time) bool {
	// 这里应该实现具体的节假日判断逻辑
	// 可以基于固定节假日或从API获取
	
	// 简单的固定节假日判断
	holidays := []string{
		"01-01", // 元旦
		"05-01", // 劳动节
		"10-01", // 国庆节
		"10-02",
		"10-03",
	}

	dateStr := date.Format("01-02")
	for _, holiday := range holidays {
		if dateStr == holiday {
			return true
		}
	}

	return false
}

// GetQuarter 获取季度
func (th *TimeHelper) GetQuarter(t time.Time) int {
	month := int(t.Month())
	return (month-1)/3 + 1
}

// GetQuarterRange 获取季度日期范围
func (th *TimeHelper) GetQuarterRange(year, quarter int) (time.Time, time.Time) {
	var startMonth, endMonth int
	
	switch quarter {
	case 1:
		startMonth, endMonth = 1, 3
	case 2:
		startMonth, endMonth = 4, 6
	case 3:
		startMonth, endMonth = 7, 9
	case 4:
		startMonth, endMonth = 10, 12
	default:
		quarter = 1
		startMonth, endMonth = 1, 3
	}

	startDate := time.Date(year, time.Month(startMonth), 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, time.Month(endMonth+1), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)

	return startDate, endDate
}

// GetMonthRange 获取月份日期范围
func (th *TimeHelper) GetMonthRange(year, month int) (time.Time, time.Time) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)
	return startDate, endDate
}

// GetWeekRange 获取周日期范围（周一到周日）
func (th *TimeHelper) GetWeekRange(t time.Time) (time.Time, time.Time) {
	weekday := int(t.Weekday())
	if weekday == 0 { // 周日
		weekday = 7
	}

	startDate := t.AddDate(0, 0, 1-weekday)
	endDate := startDate.AddDate(0, 0, 6)

	return startDate, endDate
}

// GetYearRange 获取年份日期范围
func (th *TimeHelper) GetYearRange(year int) (time.Time, time.Time) {
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 999999999, time.UTC)
	return startDate, endDate
}

// CalculateDuration 计算时间间隔
func (th *TimeHelper) CalculateDuration(start, end time.Time) time.Duration {
	return end.Sub(start)
}

// FormatDuration 格式化时间间隔
func (th *TimeHelper) FormatDuration(duration time.Duration) string {
	if duration == 0 {
		return "0秒"
	}

	totalSeconds := int(duration.Seconds())
	days := totalSeconds / (24 * 3600)
	hours := (totalSeconds % (24 * 3600)) / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if days > 0 {
		return fmt.Sprintf("%d天%d小时%d分钟", days, hours, minutes)
	} else if hours > 0 {
		if minutes > 0 || seconds > 0 {
			return fmt.Sprintf("%d小时%d分钟%d秒", hours, minutes, seconds)
		}
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	} else if minutes > 0 {
		if seconds > 0 {
			return fmt.Sprintf("%d分钟%d秒", minutes, seconds)
		}
		return fmt.Sprintf("%d分钟", minutes)
	} else {
		return fmt.Sprintf("%d秒", seconds)
	}
}

// GetCurrentTime 获取当前时间
func (th *TimeHelper) GetCurrentTime() time.Time {
	return time.Now()
}

// GetCurrentTimeString 获取当前时间字符串
func (th *TimeHelper) GetCurrentTimeString() string {
	return time.Now().Format(DateTimeFormat)
}

// AddDays 添加天数
func (th *TimeHelper) AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

// AddMonths 添加月数
func (th *TimeHelper) AddMonths(t time.Time, months int) time.Time {
	return t.AddDate(0, months, 0)
}

// GetStartOfDay 获取一天的开始时间
func (th *TimeHelper) GetStartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// GetEndOfDay 获取一天的结束时间
func (th *TimeHelper) GetEndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// GetStartOfWeek 获取一周的开始时间（周一）
func (th *TimeHelper) GetStartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 { // 周日
		weekday = 7
	}
	return th.GetStartOfDay(t.AddDate(0, 0, 1-weekday))
}

// GetStartOfMonth 获取一月的开始时间
func (th *TimeHelper) GetStartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// GetEndOfMonth 获取一月的结束时间
func (th *TimeHelper) GetEndOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location()).AddDate(0, 0, -1).Add(time.Hour*23 + time.Minute*59 + time.Second*59 + time.Nanosecond*999999999)
}

// DiffInDays 计算天数差
func (th *TimeHelper) DiffInDays(t1, t2 time.Time) int {
	return int(t2.Sub(t1).Hours() / 24)
}

// DiffInHours 计算小时差
func (th *TimeHelper) DiffInHours(t1, t2 time.Time) int {
	return int(t2.Sub(t1).Hours())
}

// IsWeekend 检查是否为周末
func (th *TimeHelper) IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// GetAge 计算年龄（支持参考时间）
func (th *TimeHelper) GetAge(birthDate, referenceDate time.Time) int {
	age := referenceDate.Year() - birthDate.Year()
	
	// 如果还没到今年的生日，年龄减1
	if referenceDate.Month() < birthDate.Month() || 
		(referenceDate.Month() == birthDate.Month() && referenceDate.Day() < birthDate.Day()) {
		age--
	}
	
	return age
}

// GetTimestamp 获取时间戳
func (th *TimeHelper) GetTimestamp(t time.Time) int64 {
	return t.Unix()
}

// FromTimestamp 从时间戳创建时间
func (th *TimeHelper) FromTimestamp(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

// ConvertTimezoneByName 转换时区（使用字符串）
func (th *TimeHelper) ConvertTimezoneByName(t time.Time, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}
	return t.In(loc), nil
}

// IsToday 检查是否为今天
func (th *TimeHelper) IsToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

// IsYesterday 检查是否为昨天
func (th *TimeHelper) IsYesterday(t time.Time) bool {
	yesterday := time.Now().AddDate(0, 0, -1)
	return t.Year() == yesterday.Year() && t.Month() == yesterday.Month() && t.Day() == yesterday.Day()
}

// IsThisWeek 检查是否为本周
func (th *TimeHelper) IsThisWeek(t time.Time) bool {
	now := time.Now()
	startWeek, endWeek := th.GetWeekRange(now)
	return !t.Before(startWeek) && !t.After(endWeek)
}

// IsThisMonth 检查是否为本月
func (th *TimeHelper) IsThisMonth(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month()
}

// IsThisYear 检查是否为今年
func (th *TimeHelper) IsThisYear(t time.Time) bool {
	return t.Year() == time.Now().Year()
}

// GetRelativeTime 获取相对时间描述
func (th *TimeHelper) GetRelativeTime(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	if duration < 0 {
		duration = -duration
		if duration < time.Minute {
			return "几秒后"
		}
		if duration < time.Hour {
			return fmt.Sprintf("%d分钟后", int(duration.Minutes()))
		}
		if duration < 24*time.Hour {
			return fmt.Sprintf("%d小时后", int(duration.Hours()))
		}
		return fmt.Sprintf("%d天后", int(duration.Hours()/24))
	}

	if duration < time.Minute {
		return "刚刚"
	}
	if duration < time.Hour {
		return fmt.Sprintf("%d分钟前", int(duration.Minutes()))
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%d小时前", int(duration.Hours()))
	}
	if duration < 7*24*time.Hour {
		return fmt.Sprintf("%d天前", int(duration.Hours()/24))
	}
	if duration < 30*24*time.Hour {
		return fmt.Sprintf("%d周前", int(duration.Hours()/(7*24)))
	}
	if duration < 365*24*time.Hour {
		return fmt.Sprintf("%d个月前", int(duration.Hours()/(30*24)))
	}
	return fmt.Sprintf("%d年前", int(duration.Hours()/(365*24)))
}

// GetNextTradingDay 获取下一个交易日
func (th *TimeHelper) GetNextTradingDay(t time.Time) time.Time {
	next := t.AddDate(0, 0, 1)
	for next.Weekday() == time.Saturday || next.Weekday() == time.Sunday || th.isHoliday(next) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

// GetPreviousTradingDay 获取上一个交易日
func (th *TimeHelper) GetPreviousTradingDay(t time.Time) time.Time {
	prev := t.AddDate(0, 0, -1)
	for prev.Weekday() == time.Saturday || prev.Weekday() == time.Sunday || th.isHoliday(prev) {
		prev = prev.AddDate(0, 0, -1)
	}
	return prev
}

// Timezone related functions
var (
	// 常用时区
	UTC       = time.UTC
	Shanghai  = time.FixedZone("CST", 8*3600)  // 北京时间
	NewYork   = time.FixedZone("EST", -5*3600) // 纽约时间
	London    = time.FixedZone("GMT", 0)       // 伦敦时间
	Tokyo     = time.FixedZone("JST", 9*3600)  // 东京时间
)

// ConvertTimezone 转换时区（使用Location）
func (th *TimeHelper) ConvertTimezone(t time.Time, to *time.Location) time.Time {
	return t.In(to)
}

// GetMarketTime 获取市场时间
func (th *TimeHelper) GetMarketTime(market string) time.Time {
	now := time.Now()
	
	switch strings.ToLower(market) {
	case "cn", "china", "shanghai":
		return now.In(Shanghai)
	case "us", "usa", "newyork":
		return now.In(NewYork)
	case "uk", "london":
		return now.In(London)
	case "jp", "japan", "tokyo":
		return now.In(Tokyo)
	default:
		return now.In(UTC)
	}
}

// IsMarketOpen 检查市场是否开盘（简化版本）
func (th *TimeHelper) IsMarketOpen(market string) bool {
	marketTime := th.GetMarketTime(market)
	
	// 检查是否为交易日
	if marketTime.Weekday() == time.Saturday || marketTime.Weekday() == time.Sunday {
		return false
	}
	
	if th.isHoliday(marketTime) {
		return false
	}
	
	hour := marketTime.Hour()
	minute := marketTime.Minute()
	
	switch strings.ToLower(market) {
	case "cn", "china", "shanghai":
		// 中国股市: 9:30-11:30, 13:00-15:00
		return (hour == 9 && minute >= 30) || hour == 10 || hour == 11 && minute <= 30 ||
			   hour == 13 || hour == 14 || (hour == 15 && minute == 0)
	case "us", "usa", "newyork":
		// 美国股市: 9:30-16:00 (EST)
		return (hour == 9 && minute >= 30) || (hour >= 10 && hour <= 15) || (hour == 16 && minute == 0)
	default:
		return false
	}
}

// ParseDurationString 解析时间间隔字符串
func (th *TimeHelper) ParseDurationString(durationStr string) (time.Duration, error) {
	durationStr = strings.TrimSpace(durationStr)
	
	// 支持中文时间单位
	replacements := map[string]string{
		"年": "y",
		"月": "M",
		"周": "w",
		"天": "d",
		"日": "d",
		"时": "h",
		"小时": "h",
		"分": "m",
		"分钟": "m",
		"秒": "s",
	}
	
	for old, new := range replacements {
		durationStr = strings.ReplaceAll(durationStr, old, new)
	}
	
	// 处理年、月、周、天等Go不直接支持的单位
	if strings.Contains(durationStr, "y") {
		parts := strings.Split(durationStr, "y")
		if len(parts) == 2 {
			years, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return 0, fmt.Errorf("无效的年份数值: %s", parts[0])
			}
			return time.Duration(years) * 365 * 24 * time.Hour, nil
		}
	}
	
	if strings.Contains(durationStr, "M") {
		parts := strings.Split(durationStr, "M")
		if len(parts) == 2 {
			months, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return 0, fmt.Errorf("无效的月份数值: %s", parts[0])
			}
			return time.Duration(months) * 30 * 24 * time.Hour, nil
		}
	}
	
	if strings.Contains(durationStr, "w") {
		parts := strings.Split(durationStr, "w")
		if len(parts) == 2 {
			weeks, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return 0, fmt.Errorf("无效的周数值: %s", parts[0])
			}
			return time.Duration(weeks) * 7 * 24 * time.Hour, nil
		}
	}
	
	if strings.Contains(durationStr, "d") {
		parts := strings.Split(durationStr, "d")
		if len(parts) == 2 {
			days, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return 0, fmt.Errorf("无效的天数值: %s", parts[0])
			}
			return time.Duration(days) * 24 * time.Hour, nil
		}
	}
	
	// 使用Go标准解析
	return time.ParseDuration(durationStr)
}