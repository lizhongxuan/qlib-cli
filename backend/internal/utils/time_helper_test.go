package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTimeHelper(t *testing.T) {
	helper := NewTimeHelper()
	assert.NotNil(t, helper)
}

func TestParseDate(t *testing.T) {
	helper := NewTimeHelper()

	tests := []struct {
		dateStr   string
		shouldErr bool
		expected  time.Time
	}{
		{
			dateStr:   "2023-12-25",
			shouldErr: false,
			expected:  time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			dateStr:   "2023-12-25 15:30:45",
			shouldErr: false,
			expected:  time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
		},
		{
			dateStr:   "12/25/2023",
			shouldErr: false,
			expected:  time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			dateStr:   "2023/12/25",
			shouldErr: false,
			expected:  time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			dateStr:   "2023年12月25日",
			shouldErr: false,
			expected:  time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			dateStr:   "invalid date",
			shouldErr: true,
		},
		{
			dateStr:   "",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.dateStr, func(t *testing.T) {
			result, err := helper.ParseDate(tt.dateStr)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Year(), result.Year())
				assert.Equal(t, tt.expected.Month(), result.Month())
				assert.Equal(t, tt.expected.Day(), result.Day())
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	helper := NewTimeHelper()
	testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)

	tests := []struct {
		format   string
		expected string
	}{
		{DateFormat, "2023-12-25"},
		{DateTimeFormat, "2023-12-25 15:30:45"},
		{ShortDateFormat, "12/25/2023"},
		{ChineseDateFormat, "2023年12月25日"},
		{TimeFormat, "15:30:45"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := helper.FormatDate(testTime, tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCurrentTime(t *testing.T) {
	helper := NewTimeHelper()
	
	before := time.Now()
	current := helper.GetCurrentTime()
	after := time.Now()

	assert.True(t, current.After(before.Add(-time.Second)))
	assert.True(t, current.Before(after.Add(time.Second)))
}

func TestGetCurrentTimeString(t *testing.T) {
	helper := NewTimeHelper()
	
	timeStr := helper.GetCurrentTimeString()
	assert.NotEmpty(t, timeStr)
	
	// 验证格式
	_, err := time.Parse(DateTimeFormat, timeStr)
	assert.NoError(t, err)
}

func TestAddDays(t *testing.T) {
	helper := NewTimeHelper()
	baseTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)

	tests := []struct {
		days     int
		expected time.Time
	}{
		{1, time.Date(2023, 12, 26, 15, 30, 45, 0, time.UTC)},
		{-1, time.Date(2023, 12, 24, 15, 30, 45, 0, time.UTC)},
		{7, time.Date(2024, 1, 1, 15, 30, 45, 0, time.UTC)},
		{0, baseTime},
		{365, time.Date(2024, 12, 24, 15, 30, 45, 0, time.UTC)}, // 2024年是闰年
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.days)), func(t *testing.T) {
			result := helper.AddDays(baseTime, tt.days)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAddMonths(t *testing.T) {
	helper := NewTimeHelper()
	baseTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)

	tests := []struct {
		months   int
		expected time.Time
	}{
		{1, time.Date(2024, 1, 25, 15, 30, 45, 0, time.UTC)},
		{-1, time.Date(2023, 11, 25, 15, 30, 45, 0, time.UTC)},
		{12, time.Date(2024, 12, 25, 15, 30, 45, 0, time.UTC)},
		{0, baseTime},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.months)), func(t *testing.T) {
			result := helper.AddMonths(baseTime, tt.months)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetStartOfDay(t *testing.T) {
	helper := NewTimeHelper()
	inputTime := time.Date(2023, 12, 25, 15, 30, 45, 123456789, time.UTC)
	expected := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)

	result := helper.GetStartOfDay(inputTime)
	assert.Equal(t, expected, result)
}

func TestGetEndOfDay(t *testing.T) {
	helper := NewTimeHelper()
	inputTime := time.Date(2023, 12, 25, 15, 30, 45, 123456789, time.UTC)
	expected := time.Date(2023, 12, 25, 23, 59, 59, 999999999, time.UTC)

	result := helper.GetEndOfDay(inputTime)
	assert.Equal(t, expected, result)
}

func TestGetStartOfWeek(t *testing.T) {
	helper := NewTimeHelper()
	// 2023-12-25 是星期一
	inputTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
	expected := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC) // 星期一

	result := helper.GetStartOfWeek(inputTime)
	assert.Equal(t, expected, result)

	// 测试星期三
	inputTime2 := time.Date(2023, 12, 27, 15, 30, 45, 0, time.UTC) // 星期三
	expected2 := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)     // 应该返回星期一

	result2 := helper.GetStartOfWeek(inputTime2)
	assert.Equal(t, expected2, result2)
}

func TestGetStartOfMonth(t *testing.T) {
	helper := NewTimeHelper()
	inputTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
	expected := time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC)

	result := helper.GetStartOfMonth(inputTime)
	assert.Equal(t, expected, result)
}

func TestGetEndOfMonth(t *testing.T) {
	helper := NewTimeHelper()
	inputTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
	expected := time.Date(2023, 12, 31, 23, 59, 59, 999999999, time.UTC)

	result := helper.GetEndOfMonth(inputTime)
	assert.Equal(t, expected, result)

	// 测试2月（非闰年）
	inputTime2 := time.Date(2023, 2, 15, 10, 0, 0, 0, time.UTC)
	expected2 := time.Date(2023, 2, 28, 23, 59, 59, 999999999, time.UTC)

	result2 := helper.GetEndOfMonth(inputTime2)
	assert.Equal(t, expected2, result2)

	// 测试2月（闰年）
	inputTime3 := time.Date(2024, 2, 15, 10, 0, 0, 0, time.UTC)
	expected3 := time.Date(2024, 2, 29, 23, 59, 59, 999999999, time.UTC)

	result3 := helper.GetEndOfMonth(inputTime3)
	assert.Equal(t, expected3, result3)
}

func TestDiffInDays(t *testing.T) {
	helper := NewTimeHelper()
	
	time1 := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)
	time2 := time.Date(2023, 12, 30, 0, 0, 0, 0, time.UTC)

	diff := helper.DiffInDays(time1, time2)
	assert.Equal(t, 5, diff)

	// 测试负数差值
	diff2 := helper.DiffInDays(time2, time1)
	assert.Equal(t, -5, diff2)

	// 测试同一天
	diff3 := helper.DiffInDays(time1, time1)
	assert.Equal(t, 0, diff3)
}

func TestDiffInHours(t *testing.T) {
	helper := NewTimeHelper()
	
	time1 := time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC)
	time2 := time.Date(2023, 12, 25, 15, 0, 0, 0, time.UTC)

	diff := helper.DiffInHours(time1, time2)
	assert.Equal(t, 5, diff)

	// 测试跨天
	time3 := time.Date(2023, 12, 26, 2, 0, 0, 0, time.UTC)
	diff2 := helper.DiffInHours(time1, time3)
	assert.Equal(t, 16, diff2)
}

func TestIsWeekend(t *testing.T) {
	helper := NewTimeHelper()

	// 星期六
	saturday := time.Date(2023, 12, 23, 0, 0, 0, 0, time.UTC)
	assert.True(t, helper.IsWeekend(saturday))

	// 星期日
	sunday := time.Date(2023, 12, 24, 0, 0, 0, 0, time.UTC)
	assert.True(t, helper.IsWeekend(sunday))

	// 星期一
	monday := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)
	assert.False(t, helper.IsWeekend(monday))

	// 星期五
	friday := time.Date(2023, 12, 22, 0, 0, 0, 0, time.UTC)
	assert.False(t, helper.IsWeekend(friday))
}

func TestIsToday(t *testing.T) {
	helper := NewTimeHelper()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 15, 30, 45, 0, now.Location())
	yesterday := now.AddDate(0, 0, -1)

	assert.True(t, helper.IsToday(today))
	assert.False(t, helper.IsToday(yesterday))
}

func TestGetAge(t *testing.T) {
	helper := NewTimeHelper()

	birthDate := time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC)
	referenceDate := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)

	age := helper.GetAge(birthDate, referenceDate)
	assert.Equal(t, 33, age)

	// 测试生日还没到的情况
	birthDate2 := time.Date(1990, 12, 26, 0, 0, 0, 0, time.UTC)
	age2 := helper.GetAge(birthDate2, referenceDate)
	assert.Equal(t, 32, age2)

	// 测试正好生日
	birthDate3 := time.Date(1990, 12, 25, 0, 0, 0, 0, time.UTC)
	age3 := helper.GetAge(birthDate3, referenceDate)
	assert.Equal(t, 33, age3)
}

func TestFormatDuration(t *testing.T) {
	helper := NewTimeHelper()

	tests := []struct {
		duration time.Duration
		expected string
	}{
		{time.Hour * 2, "2小时0分钟"},
		{time.Minute * 30, "30分钟"},
		{time.Second * 45, "45秒"},
		{time.Hour*2 + time.Minute*30 + time.Second*45, "2小时30分钟45秒"},
		{time.Hour * 25, "1天1小时0分钟"},
		{0, "0秒"},
	}

	for _, tt := range tests {
		t.Run(tt.duration.String(), func(t *testing.T) {
			result := helper.FormatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertTimezone(t *testing.T) {
	helper := NewTimeHelper()

	// 创建UTC时间
	utcTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)

	// 转换到东八区
	shanghaiTime, err := helper.ConvertTimezoneByName(utcTime, "Asia/Shanghai")
	assert.NoError(t, err)
	assert.Equal(t, 23, shanghaiTime.Hour()) // UTC+8

	// 转换到纽约时区
	nyTime, err := helper.ConvertTimezoneByName(utcTime, "America/New_York")
	assert.NoError(t, err)
	// 注意：需要考虑夏令时，12月25日应该是EST（UTC-5）
	assert.Equal(t, 10, nyTime.Hour()) // UTC-5

	// 测试无效时区
	_, err = helper.ConvertTimezoneByName(utcTime, "Invalid/Timezone")
	assert.Error(t, err)
}

func TestGetTimestamp(t *testing.T) {
	helper := NewTimeHelper()

	testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
	
	timestamp := helper.GetTimestamp(testTime)
	assert.Equal(t, testTime.Unix(), timestamp)
}

func TestFromTimestamp(t *testing.T) {
	helper := NewTimeHelper()

	// 创建一个时间并获取其时间戳，然后再转换回来测试
	originalTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
	timestamp := originalTime.Unix()

	result := helper.FromTimestamp(timestamp)
	// 转换为UTC时间进行比较
	assert.Equal(t, originalTime, result.UTC())
}