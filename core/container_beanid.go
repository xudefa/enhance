package core

import (
	"reflect"
	"strings"
)

// Generate 生成 Bean ID。
func (c *defaultContainer) Generate(typ reflect.Type, customName ...string) string {
	// 处理指针类型，获取实际类型
	actualType := typ
	if typ.Kind() == reflect.Ptr {
		actualType = typ.Elem()
	}

	// idGenerator ID 格式：pkgPath.typeName，id 生成器一定存在，不需要判断空指针
	prefix := actualType.PkgPath() + "." + actualType.Name()

	// 没有提供自定义名称
	if len(customName) == 0 || customName[0] == "" {
		return prefix
	}

	// 如果已经提供了标准格式的 Name，直接使用
	if strings.HasPrefix(customName[0], prefix) {
		return customName[0]
	}

	// 如果没有提供标准格式的 Name，使用自定义名称，格式为：pkgPath.typeName#customName
	return prefix + "#" + customName[0]
}

// Parse 解析 Bean ID 为包路径、类型名和自定义名称。
func (c *defaultContainer) Parse(beanID string) (pkgPath, typeName, customName string) {
	// 解析自定义名称
	parts := strings.SplitN(beanID, "#", 2)
	mainPart := parts[0]
	if len(parts) > 1 {
		customName = parts[1]
	}

	// 解析包路径和类型名
	lastDot := strings.LastIndex(mainPart, ".")
	if lastDot == -1 {
		return "", mainPart, customName
	}

	pkgPath = mainPart[:lastDot]
	typeName = mainPart[lastDot+1:]
	return pkgPath, typeName, customName
}
