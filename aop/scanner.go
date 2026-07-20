package aop

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// AopBeanScanner AOP Bean扫描器
//
// 扫描并注册带有AOP注解的Bean
type AopBeanScanner struct {
	mu        sync.RWMutex
	container *AopContainer
	basePath  string
	enabled   bool
}

// NewAopBeanScanner 创建AOP Bean扫描器
func NewAopBeanScanner(container *AopContainer) *AopBeanScanner {
	if container == nil {
		container = CreateAopContainer()
	}
	return &AopBeanScanner{
		container: container,
		enabled:   true,
	}
}

// Scan 扫描指定路径
func (s *AopBeanScanner) Scan(basePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.basePath = basePath

	if !s.enabled {
		return nil
	}

	// 检查路径是否存在
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return fmt.Errorf("aop scanner: path %s does not exist", basePath)
	}

	fset := token.NewFileSet()

	// 遍历目录查找 Go 文件
	err := filepath.WalkDir(basePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		// 排除测试文件、生成的 AOP 文件和以 _ 开头的文件
		if !strings.HasSuffix(name, "_test.go") &&
			!strings.HasSuffix(name, "_aop.go") &&
			!strings.HasPrefix(name, "_") &&
			strings.HasSuffix(name, ".go") {

			// 解析文件
			file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				// 解析失败时跳过该文件
				return nil
			}

			// 扫描带有 @AopProxy 注解的类型并注册
			if err := s.scanAopProxy(file, path); err != nil {
				return fmt.Errorf("aop scanner: failed to scan %s: %w", path, err)
			}
		}
		return nil
	})

	return err
}

// scanAopProxy 扫描文件中带有 @AopProxy 注解的类型
func (s *AopBeanScanner) scanAopProxy(file *ast.File, filePath string) error {
	// 遍历所有类型声明
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			// 检查注释中是否包含 @AopProxy 注解
			if typeSpec.Doc == nil {
				continue
			}

			commentText := typeSpec.Doc.Text()
			if strings.Contains(commentText, "@AopProxy") || strings.Contains(commentText, "@aopProxy") {
				// 找到带有 @AopProxy 注解的类型，注册到容器中
				typeName := typeSpec.Name.Name
				s.container.registerProxyType(typeName, filePath)
			}
		}
	}

	return nil
}

// Enable 启用扫描器
func (s *AopBeanScanner) Enable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = true
}

// Disable 禁用扫描器
func (s *AopBeanScanner) Disable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = false
}

// IsEnabled 检查是否启用
func (s *AopBeanScanner) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// GetContainer 获取容器
func (s *AopBeanScanner) GetContainer() *AopContainer {
	return s.container
}

// GlobalAopBeanScanner 全局AOP Bean扫描器
var GlobalAopBeanScanner = NewAopBeanScanner(nil)

// ScanAopBeans 扫描AOP Bean
func ScanAopBeans(basePath string) error {
	return GlobalAopBeanScanner.Scan(basePath)
}

// AutoScan 自动扫描
func AutoScan() error {
	return GlobalAopBeanScanner.Scan(".")
}
