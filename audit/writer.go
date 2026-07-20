// Package audit 提供审计日志功能，用于 enhance 框架。
package audit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// consoleWriterImpl EventWriter 接口的控制台实现。
type consoleWriterImpl struct {
	mu     sync.Mutex
	output *os.File
}

// fileWriterImpl EventWriter 接口的文件实现。
type fileWriterImpl struct {
	mu     sync.Mutex
	file   *os.File
	writer *bufio.Writer
}

// NewConsoleWriter 创建控制台写入器。
func NewConsoleWriter() EventWriter {
	return &consoleWriterImpl{
		output: os.Stdout,
	}
}

func (w *consoleWriterImpl) Write(event Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("序列化事件失败: %w", err)
	}

	_, err = fmt.Fprintln(w.output, string(data))
	return err
}

func (w *consoleWriterImpl) Close() error {
	return nil
}

// NewFileWriter 创建文件写入器
//
// 打开或创建指定路径的文件，使用 4KB 缓冲区进行缓冲写入。
func NewFileWriter(filePath string) (EventWriter, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开文件 %s 失败: %w", filePath, err)
	}

	return &fileWriterImpl{
		file:   file,
		writer: bufio.NewWriterSize(file, 4096),
	}, nil
}

func (w *fileWriterImpl) Write(event Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("序列化事件失败: %w", err)
	}

	_, err = w.writer.Write(data)
	if err != nil {
		return err
	}

	if err := w.writer.WriteByte('\n'); err != nil {
		return err
	}

	return w.writer.Flush()
}

func (w *fileWriterImpl) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.writer.Flush(); err != nil {
		return fmt.Errorf("刷新缓冲区失败: %w", err)
	}

	if err := w.file.Close(); err != nil {
		return fmt.Errorf("关闭文件失败: %w", err)
	}
	return nil
}
