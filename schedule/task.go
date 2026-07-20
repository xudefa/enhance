// Package schedule 提供定时任务调度功能，用于 enhance 框架。
package schedule

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// functionTask 基于函数的任务实现。
type functionTask struct {
	name      string
	cron      string
	executeFn func(ctx context.Context) error
}

// NewTask 创建基于函数的任务实例。
func NewTask(name, cron string, fn func(ctx context.Context) error) Task {
	return &functionTask{
		name:      name,
		cron:      cron,
		executeFn: fn,
	}
}

func (t *functionTask) Name() string {
	return t.name
}

func (t *functionTask) Cron() string {
	return t.cron
}

func (t *functionTask) Execute(ctx context.Context) error {
	return t.executeFn(ctx)
}

// fixedDelayTask 固定延迟任务实现。
type fixedDelayTask struct {
	name      string
	delay     time.Duration
	executeFn func(ctx context.Context) error
}

// NewFixedDelayTask 创建固定延迟任务。
//
// 延迟时间从上次任务完成时开始计算。
func NewFixedDelayTask(name string, delay time.Duration, fn func(ctx context.Context) error) Task {
	return &fixedDelayTask{
		name:      name,
		delay:     delay,
		executeFn: fn,
	}
}

func (t *fixedDelayTask) Name() string {
	return t.name
}

func (t *fixedDelayTask) Cron() string {
	return fmt.Sprintf("@fixed-delay(%s)", t.delay)
}

func (t *fixedDelayTask) Execute(ctx context.Context) error {
	return t.executeFn(ctx)
}

// FixedDelay 获取固定延迟时间。
func (t *fixedDelayTask) FixedDelay() time.Duration {
	return t.delay
}

// fixedRateTask 固定频率任务实现。
type fixedRateTask struct {
	name      string
	interval  time.Duration
	executeFn func(ctx context.Context) error
}

// NewFixedRateTask 创建固定频率任务。
//
// 从任务开始执行时计算下一次执行时间。
func NewFixedRateTask(name string, interval time.Duration, fn func(ctx context.Context) error) Task {
	return &fixedRateTask{
		name:      name,
		interval:  interval,
		executeFn: fn,
	}
}

func (t *fixedRateTask) Name() string {
	return t.name
}

func (t *fixedRateTask) Cron() string {
	return fmt.Sprintf("@fixed-rate(%s)", t.interval)
}

func (t *fixedRateTask) Execute(ctx context.Context) error {
	return t.executeFn(ctx)
}

// Interval 获取执行间隔。
func (t *fixedRateTask) Interval() time.Duration {
	return t.interval
}

// CronExpression Cron 表达式解析器。
//
// 支持 6 字段 Spring 风格 Cron 表达式：秒 分 时 日 月 周
// 使用位图编码实现高效的下次执行时间计算。
type CronExpression struct {
	second     uint64
	minute     uint64
	hour       uint64
	dayOfMonth uint64
	month      uint64
	dayOfWeek  uint64
}

// ParseCronExpression 解析 Cron 表达式字符串。
//
// 支持格式：秒 分 时 日 月 周
// 示例：
//   - "0 */5 * * * *" : 每 5 分钟
//   - "0 0 */1 * * *" : 每小时
//   - "0 0 0 * * *" : 每天零点
//   - "0 0 0 * * MON-FRI" : 工作日零点
func ParseCronExpression(expr string) (*CronExpression, error) {
	fields := strings.Fields(expr)
	if len(fields) != 6 {
		return nil, fmt.Errorf("invalid cron expression: expected 6 fields, got %d", len(fields))
	}

	ce := &CronExpression{}

	var err error
	ce.second, err = parseField(fields[0], 0, 59, dayOfWeekMap)
	if err != nil {
		return nil, fmt.Errorf("invalid second field: %w", err)
	}

	ce.minute, err = parseField(fields[1], 0, 59, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid minute field: %w", err)
	}

	ce.hour, err = parseField(fields[2], 0, 23, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid hour field: %w", err)
	}

	ce.dayOfMonth, err = parseField(fields[3], 1, 31, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid day of month field: %w", err)
	}

	ce.month, err = parseField(fields[4], 1, 12, monthMap)
	if err != nil {
		return nil, fmt.Errorf("invalid month field: %w", err)
	}

	ce.dayOfWeek, err = parseField(fields[5], 0, 6, dayOfWeekMap)
	if err != nil {
		return nil, fmt.Errorf("invalid day of week field: %w", err)
	}

	return ce, nil
}

// Next 计算给定时间之后的下次执行时间。
//
// 使用逐字段匹配算法，从当前时间开始向后搜索最近的匹配时间。
func (ce *CronExpression) Next(from time.Time) time.Time {
	year := from.Year()
	month := int(from.Month())
	day := from.Day()
	hour := from.Hour()
	minute := from.Minute()
	second := from.Second() + 1

	for range 4 * 366 * 24 * 60 {
		if !hasBit(ce.second, second) {
			second++
			if second >= 60 {
				second = 0
				minute++
			}
			continue
		}

		if !hasBit(ce.minute, minute) {
			minute++
			second = 0
			if minute >= 60 {
				minute = 0
				hour++
			}
			continue
		}

		if !hasBit(ce.hour, hour) {
			hour++
			minute = 0
			second = 0
			if hour >= 24 {
				hour = 0
				day++
			}
			continue
		}

		if !hasBit(ce.dayOfMonth, day) || !hasBit(ce.month, month) {
			day++
			hour = 0
			minute = 0
			second = 0
			if day > 31 {
				day = 1
				month++
				if month > 12 {
					month = 1
					year++
				}
			}
			continue
		}

		t := time.Date(year, time.Month(month), day, hour, minute, second, 0, from.Location())
		if !hasBit(ce.dayOfWeek, int(t.Weekday())) {
			day++
			hour = 0
			minute = 0
			second = 0
			if day > 31 {
				day = 1
				month++
				if month > 12 {
					month = 1
					year++
				}
			}
			continue
		}

		return t
	}

	return time.Time{}
}

var (
	monthMap = map[string]int{
		"JAN": 1, "FEB": 2, "MAR": 3, "APR": 4,
		"MAY": 5, "JUN": 6, "JUL": 7, "AUG": 8,
		"SEP": 9, "OCT": 10, "NOV": 11, "DEC": 12,
	}

	dayOfWeekMap = map[string]int{
		"SUN": 0, "MON": 1, "TUE": 2, "WED": 3,
		"THU": 4, "FRI": 5, "SAT": 6,
	}
)

func parseField(field string, min, max int, names map[string]int) (uint64, error) {
	if field == "*" {
		return setAllBits(min, max), nil
	}

	var bitmap uint64
	parts := strings.Split(field, ",")

	for _, part := range parts {
		if err := parsePart(part, min, max, names, &bitmap); err != nil {
			return 0, err
		}
	}

	return bitmap, nil
}

func parsePart(part string, min, max int, names map[string]int, bitmap *uint64) error {
	if strings.Contains(part, "/") {
		return parseStep(part, min, max, names, bitmap)
	}

	if strings.Contains(part, "-") {
		return parseRange(part, min, max, names, bitmap)
	}

	val, err := parseValue(part, names)
	if err != nil {
		return err
	}

	if val < min || val > max {
		return fmt.Errorf("value %d out of range [%d, %d]", val, min, max)
	}

	*bitmap = setBit(*bitmap, val)
	return nil
}

func parseStep(part string, min, max int, names map[string]int, bitmap *uint64) error {
	parts := strings.Split(part, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid step format: %s", part)
	}

	start := min
	if parts[0] != "*" {
		// 支持范围+步长格式，如 "0-30/5"
		if strings.Contains(parts[0], "-") {
			rangeParts := strings.Split(parts[0], "-")
			if len(rangeParts) != 2 {
				return fmt.Errorf("invalid range format: %s", parts[0])
			}

			rangeStart, err := parseValue(rangeParts[0], names)
			if err != nil {
				return err
			}

			rangeEnd, err := parseValue(rangeParts[1], names)
			if err != nil {
				return err
			}

			step, err := strconv.Atoi(parts[1])
			if err != nil {
				return fmt.Errorf("invalid step value: %w", err)
			}

			if step <= 0 {
				return fmt.Errorf("step value must be positive")
			}

			for i := rangeStart; i <= rangeEnd; i += step {
				*bitmap = setBit(*bitmap, i)
			}

			return nil
		}

		val, err := parseValue(parts[0], names)
		if err != nil {
			return err
		}
		start = val
	}

	step, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid step value: %w", err)
	}

	if step <= 0 {
		return fmt.Errorf("step value must be positive")
	}

	for i := start; i <= max; i += step {
		*bitmap = setBit(*bitmap, i)
	}

	return nil
}

func parseRange(part string, min, max int, names map[string]int, bitmap *uint64) error {
	parts := strings.Split(part, "-")
	if len(parts) != 2 {
		return fmt.Errorf("invalid range format: %s", part)
	}

	start, err := parseValue(parts[0], names)
	if err != nil {
		return err
	}

	end, err := parseValue(parts[1], names)
	if err != nil {
		return err
	}

	if start < min || end > max {
		return fmt.Errorf("range [%d, %d] out of bounds [%d, %d]", start, end, min, max)
	}

	if start > end {
		return fmt.Errorf("invalid range: start %d > end %d", start, end)
	}

	for i := start; i <= end; i++ {
		*bitmap = setBit(*bitmap, i)
	}

	return nil
}

func parseValue(s string, names map[string]int) (int, error) {
	if names != nil {
		if val, ok := names[strings.ToUpper(s)]; ok {
			return val, nil
		}
	}

	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid value: %s", s)
	}

	return val, nil
}

func setBit(bitmap uint64, pos int) uint64 {
	return bitmap | (1 << pos)
}

func hasBit(bitmap uint64, pos int) bool {
	return (bitmap & (1 << pos)) != 0
}

func setAllBits(min, max int) uint64 {
	var bitmap uint64
	for i := min; i <= max; i++ {
		bitmap = setBit(bitmap, i)
	}
	return bitmap
}
