package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Parser Go 源码注解解析器
//
// 扫描 Go 源码文件，解析 @Aspect、@AopProxy、@ProxyInterface、@Before 等注解，
// 提取切面、代理和通知信息。
type Parser struct {
	fset       *token.FileSet            // 文件位置信息
	aspects    map[string]*AspectInfo    // 切面信息映射（按结构体名索引）
	proxies    map[string]*ProxyInfo     // 代理信息映射（按结构体名索引）
	interfaces map[string]*InterfaceInfo // 接口信息映射（按接口名索引）
	funcs      map[string]*AdviceInfo    // 独立函数通知映射（按函数名索引）
}

// NewParser 创建源码注解解析器
func NewParser() *Parser {
	return &Parser{
		fset:       token.NewFileSet(),
		aspects:    make(map[string]*AspectInfo),
		proxies:    make(map[string]*ProxyInfo),
		interfaces: make(map[string]*InterfaceInfo),
		funcs:      make(map[string]*AdviceInfo),
	}
}

// ParseDir 递归扫描目录中的 Go 源码文件并解析注解
//
// 跳过 _test.go 和 _aop.go 文件。
func (p *Parser) ParseDir(dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, "_aop.go") {
			return nil
		}
		return p.parseFile(path)
	})
}

// parseFile 解析单个 Go 源码文件
func (p *Parser) parseFile(filePath string) error {
	f, err := parser.ParseFile(p.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	pkgName := f.Name.Name

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			p.parseGenDecl(d, pkgName, filePath)
		case *ast.FuncDecl:
			p.parseFuncDecl(d, pkgName)
		}
	}

	// 解析接口定义中的方法
	for _, decl := range f.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if intf, ok := typeSpec.Type.(*ast.InterfaceType); ok {
						p.parseInterfaceMethods(intf, typeSpec.Name.Name, pkgName, filePath)
					}
				}
			}
		}
	}

	return nil
}

// parseGenDecl 解析类型声明，提取 @Aspect 和 @AopProxy 注解
func (p *Parser) parseGenDecl(decl *ast.GenDecl, pkgName, filePath string) {
	if decl.Tok != token.TYPE {
		return
	}

	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		typeName := typeSpec.Name.Name
		comments := decl.Doc

		if comments == nil {
			continue
		}

		if _, ok := typeSpec.Type.(*ast.StructType); ok {
			for _, comment := range comments.List {
				text := strings.TrimSpace(comment.Text)
				text = strings.TrimPrefix(text, "//")
				text = strings.TrimSpace(text)

				if strings.HasPrefix(text, "@Aspect") {
					p.parseAspectAnnotation(text, typeName, pkgName)
				} else if strings.HasPrefix(text, "@AopProxy") {
					p.parseProxyAnnotation(text, typeName, pkgName, filePath)
				}
			}
		} else if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
			for _, comment := range comments.List {
				text := strings.TrimSpace(comment.Text)
				text = strings.TrimPrefix(text, "//")
				text = strings.TrimSpace(text)

				if strings.HasPrefix(text, "@ProxyInterface") {
					p.parseInterfaceAnnotation(text, typeName, pkgName, filePath)
				}
			}
		}
	}
}

// GetAspects 获取所有解析到的切面信息
func (p *Parser) GetAspects() []AspectInfo {
	var result []AspectInfo
	for _, aspect := range p.aspects {
		result = append(result, *aspect)
	}
	return result
}

// GetProxies 获取所有解析到的代理信息
func (p *Parser) GetProxies() []ProxyInfo {
	var result []ProxyInfo
	for _, proxy := range p.proxies {
		result = append(result, *proxy)
	}
	return result
}

// GetFuncs 获取所有解析到的独立通知函数
func (p *Parser) GetFuncs() []AdviceInfo {
	var result []AdviceInfo
	for _, advice := range p.funcs {
		result = append(result, *advice)
	}
	return result
}

// GetInterfaces 获取所有解析到的接口信息
func (p *Parser) GetInterfaces() []InterfaceInfo {
	var result []InterfaceInfo
	for _, intf := range p.interfaces {
		result = append(result, *intf)
	}
	return result
}
