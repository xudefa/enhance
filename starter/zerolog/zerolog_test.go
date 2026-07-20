package zerolog

import (
	"bytes"
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/xudefa/enhance/log"
)

func TestZerologLoggerBasic(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.DebugLevel).With().Timestamp().Logger()

	logger := &ZerologLogger{
		logger:    zl,
		level:     zerolog.DebugLevel,
		format:    "json",
		addSource: false,
		output:    &buf,
	}

	ctx := context.Background()

	// 测试 Info 日志
	logger.Info(ctx, "测试信息", log.KeyValue{Key: "key1", Value: "value1"})
	if buf.Len() == 0 {
		t.Error("Info() 应该写入日志")
	}
	buf.Reset()

	// 测试 Debug 日志
	logger.Debug(ctx, "调试信息", log.KeyValue{Key: "key2", Value: "value2"})
	if buf.Len() == 0 {
		t.Error("Debug() 应该写入日志")
	}
	buf.Reset()

	// 测试 Warn 日志
	logger.Warn(ctx, "警告信息", log.KeyValue{Key: "key3", Value: "value3"})
	if buf.Len() == 0 {
		t.Error("Warn() 应该写入日志")
	}
	buf.Reset()

	// 测试 Error 日志
	logger.Error(ctx, "错误信息", log.KeyValue{Key: "key4", Value: "value4"})
	if buf.Len() == 0 {
		t.Error("Error() 应该写入日志")
	}
}

func TestZerologLoggerWith(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.DebugLevel).With().Timestamp().Logger()

	logger := &ZerologLogger{
		logger:    zl,
		level:     zerolog.DebugLevel,
		format:    "json",
		addSource: false,
		output:    &buf,
	}

	ctx := context.Background()

	// 测试 With 方法
	childLogger := logger.With(ctx, log.KeyValue{Key: "module", Value: "test"})
	childLogger.Info(ctx, "带额外字段的日志")

	if buf.Len() == 0 {
		t.Error("With().Info() 应该写入日志")
	}
}

func TestZerologLoggerSync(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.DebugLevel).With().Timestamp().Logger()

	logger := &ZerologLogger{
		logger:    zl,
		level:     zerolog.DebugLevel,
		format:    "json",
		addSource: false,
		output:    &buf,
	}

	// 测试 Sync 方法
	if err := logger.Sync(); err != nil {
		t.Errorf("Sync() error = %v", err)
	}
}

func TestZerologLoggerInterface(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.DebugLevel).With().Timestamp().Logger()

	var logger log.Logger = &ZerologLogger{
		logger:    zl,
		level:     zerolog.DebugLevel,
		format:    "json",
		addSource: false,
		output:    &buf,
	}

	// 验证实现了 log.Logger 接口
	if logger == nil {
		t.Error("ZerologLogger 应该实现 log.Logger 接口")
	}
}
