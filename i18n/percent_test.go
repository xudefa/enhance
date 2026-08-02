package i18n

import (
	"testing"
)

// TestFormatMessage_LiteralPercent 验证包含字面量 % 的消息不被破坏（回归测试）。
//
// 背景：formatMessage 直接使用 fmt.Sprintf(msg, args...)，当消息包含字面量 %
// （如 "折扣 50% 使用"）时会被 Sprintf 解释为格式动词，产生 %!(NOVERB) 乱码。
func TestFormatMessage_LiteralPercent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		msg  string
		args []any
		want string
	}{
		{"literal percent mid", "折扣 50% 使用", []any{"unused"}, "折扣 50% 使用"},
		{"literal percent trailing", "完成度 80%", []any{"unused"}, "完成度 80%"},
		{"no args keeps literal percent", "折扣 50% 使用", nil, "折扣 50% 使用"},
		{"standard verbs", "Hello, %s!", []any{"World"}, "Hello, World!"},
		{"mixed verbs", "Welcome, %s! You have %d messages.", []any{"Alice", 5}, "Welcome, Alice! You have 5 messages."},
		{"mixed literal and verb", "价格 %d 元 (折扣 50%)", []any{100}, "价格 100 元 (折扣 50%)"},
		{"width verb", "%5s", []any{"ab"}, "   ab"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := formatMessage(tt.msg, tt.args); got != tt.want {
				t.Errorf("formatMessage(%q, %v) = %q, want %q", tt.msg, tt.args, got, tt.want)
			}
		})
	}
}

// TestStaticMessageSource_GetMessage_LiteralPercent 通过公开 API 验证字面量 % 消息。
func TestStaticMessageSource_GetMessage_LiteralPercent(t *testing.T) {
	t.Parallel()
	src := NewStaticMessageSource()
	src.AddMessage("discount", "折扣 50% 使用")

	got := src.GetMessage("discount", "unused")
	if got != "折扣 50% 使用" {
		t.Errorf("GetMessage() = %q, want %q", got, "折扣 50% 使用")
	}
}
