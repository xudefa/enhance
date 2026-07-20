// Package metadata 提供元数据管理功能，用于 enhance 框架。
package metadata

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// tagAnnotationResolverImpl TagAnnotationResolver 接口的默认实现。
type tagAnnotationResolverImpl struct {
	mu      sync.RWMutex
	cache   sync.Map
	tagName string
}

// NewTagAnnotationResolver 创建 TagAnnotationResolver。
//
// 参数:
//   - tagName: tag 名称，默认为 "metadata"
//
// 返回值:
//   - TagAnnotationResolver: 解析器实例
func NewTagAnnotationResolver(tagName string) TagAnnotationResolver {
	if tagName == "" {
		tagName = "metadata"
	}
	return &tagAnnotationResolverImpl{
		tagName: tagName,
	}
}

// ResolveAnnotations 解析类型的所有注解
//
// 参数:
//   - t: 反射类型
//
// 返回值:
//   - []Annotation: 注解列表
func (r *tagAnnotationResolverImpl) ResolveAnnotations(t reflect.Type) []Annotation {
	// 检查缓存
	if cached, ok := r.cache.Load(t); ok {
		return cached.([]Annotation)
	}

	var annotations []Annotation

	// 如果是指针类型，获取其指向的类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 只处理结构体类型
	if t.Kind() != reflect.Struct {
		return annotations
	}

	// 遍历所有字段
	for i := range t.NumField() {
		field := t.Field(i)

		// 跳过未导出字段
		if !field.IsExported() {
			continue
		}

		// 获取 tag 值
		tagValue := field.Tag.Get(r.tagName)
		if tagValue == "" {
			continue
		}

		// 解析 tag 值
		ann := r.parseTagValue(tagValue, field)
		if ann.Name != "" {
			annotations = append(annotations, ann)
		}
	}

	// 缓存结果
	r.cache.Store(t, annotations)

	return annotations
}

// parseTagValue 解析 tag 值
//
// 支持格式:
//   - "name:attr1=val1,attr2=val2"
//   - "name" (仅有注解名称)
//   - "name:attr1=val1" (单个属性)
//
// 参数:
//   - tagValue: tag 值字符串
//   - field: 结构体字段
//
// 返回值:
//   - Annotation: 解析后的注解
func (r *tagAnnotationResolverImpl) parseTagValue(tagValue string, field reflect.StructField) Annotation {
	ann := Annotation{
		Attributes: make(map[string]any),
	}

	// 分割注解名称和属性
	parts := strings.SplitN(tagValue, ":", 2)
	ann.Name = strings.TrimSpace(parts[0])

	if len(parts) < 2 || parts[1] == "" {
		return ann
	}

	// 解析属性
	attrStr := strings.TrimSpace(parts[1])
	attrs := r.parseAttributes(attrStr, field.Type)

	// 合并属性
	for k, v := range attrs {
		ann.Attributes[k] = v
	}

	return ann
}

// parseAttributes 解析属性字符串
//
// 支持格式:
//   - "key1=value1,key2=value2"
//   - "key1=val1"
//   - "required=true,minLength=1,maxLength=100"
//
// 参数:
//   - attrStr: 属性字符串
//   - fieldType: 字段类型（用于类型推断）
//
// 返回值:
//   - map[string]any: 属性映射
func (r *tagAnnotationResolverImpl) parseAttributes(attrStr string, fieldType reflect.Type) map[string]any {
	attrs := make(map[string]any)

	// 分割多个属性（支持逗号分隔）
	pairs := splitAttributes(attrStr)

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		// 分割键值对
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			// 没有值的键，视为布尔值 true
			attrs[strings.TrimSpace(kv[0])] = true
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		// 尝试类型转换
		attrs[key] = r.convertValue(value)
	}

	return attrs
}

// splitAttributes 分割属性字符串，支持引号内的逗号
func splitAttributes(attrStr string) []string {
	var result []string
	var current strings.Builder
	inQuotes := false

	for _, ch := range attrStr {
		switch ch {
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(ch)
		case ',':
			if inQuotes {
				current.WriteRune(ch)
				continue
			}
			result = append(result, current.String())
			current.Reset()
		default:
			current.WriteRune(ch)
		}
	}

	// 添加最后一个属性
	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// convertValue 转换字符串值为适当的类型
//
// 支持的类型:
//   - bool: "true", "false"
//   - int: "123"
//   - float64: "3.14"
//   - string: 其他值
//
// 参数:
//   - value: 字符串值
//
// 返回值:
//   - any: 转换后的值
func (r *tagAnnotationResolverImpl) convertValue(value string) any {
	// 移除引号
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}

	// 尝试布尔值
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}

	// 尝试整数
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}

	// 尝试浮点数
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}

	// 返回字符串
	return value
}

// GetAnnotation 获取指定名称的注解
//
// 参数:
//   - target: 目标类型或实例
//   - name: 注解名称
//
// 返回值:
//   - Annotation: 注解，未找到时返回空注解
func (r *tagAnnotationResolverImpl) GetAnnotation(target any, name string) Annotation {
	t := reflect.TypeOf(target)
	annotations := r.ResolveAnnotations(t)

	for _, ann := range annotations {
		if ann.Name == name {
			return ann
		}
	}

	return Annotation{}
}

// HasAnnotation 检查是否存在指定注解
//
// 参数:
//   - target: 目标类型或实例
//   - name: 注解名称
//
// 返回值:
//   - bool: 是否存在
func (r *tagAnnotationResolverImpl) HasAnnotation(target any, name string) bool {
	t := reflect.TypeOf(target)
	annotations := r.ResolveAnnotations(t)

	for _, ann := range annotations {
		if ann.Name == name {
			return true
		}
	}

	return false
}

// GetAnnotations 获取所有注解
//
// 参数:
//   - target: 目标类型或实例
//
// 返回值:
//   - []Annotation: 注解列表
func (r *tagAnnotationResolverImpl) GetAnnotations(target any) []Annotation {
	t := reflect.TypeOf(target)
	return r.ResolveAnnotations(t)
}

// GetFieldAnnotations 获取指定字段的所有注解
//
// 参数:
//   - target: 目标类型或实例
//   - fieldName: 字段名称
//
// 返回值:
//   - []Annotation: 注解列表
//   - error: 字段不存在时返回错误
func (r *tagAnnotationResolverImpl) GetFieldAnnotations(target any, fieldName string) ([]Annotation, error) {
	t := reflect.TypeOf(target)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("target must be a struct or a pointer to a struct")
	}

	field, ok := t.FieldByName(fieldName)
	if !ok {
		return nil, fmt.Errorf("field %s does not exist", fieldName)
	}

	tagValue := field.Tag.Get(r.tagName)
	if tagValue == "" {
		return nil, nil
	}

	ann := r.parseTagValue(tagValue, field)
	if ann.Name == "" {
		return nil, nil
	}

	return []Annotation{ann}, nil
}

// GetFieldAnnotation 获取指定字段的指定注解
//
// 参数:
//   - target: 目标类型或实例
//   - fieldName: 字段名称
//   - annotationName: 注解名称
//
// 返回值:
//   - Annotation: 注解
//   - error: 字段不存在时返回错误
func (r *tagAnnotationResolverImpl) GetFieldAnnotation(target any, fieldName, annotationName string) (Annotation, error) {
	annotations, err := r.GetFieldAnnotations(target, fieldName)
	if err != nil {
		return Annotation{}, err
	}

	for _, ann := range annotations {
		if ann.Name == annotationName {
			return ann, nil
		}
	}

	return Annotation{}, nil
}

// GetStringAttribute 获取字符串类型的属性值
//
// 参数:
//   - ann: 注解
//   - key: 属性键
//
// 返回值:
//   - string: 属性值
//   - bool: 是否存在
func GetStringAttribute(ann Annotation, key string) (string, bool) {
	val, ok := ann.Attributes[key]
	if !ok {
		return "", false
	}

	if str, ok := val.(string); ok {
		return str, true
	}

	return fmt.Sprintf("%v", val), true
}

// GetIntAttribute 获取整数类型的属性值
//
// 参数:
//   - ann: 注解
//   - key: 属性键
//
// 返回值:
//   - int: 属性值
//   - bool: 是否存在且类型匹配
func GetIntAttribute(ann Annotation, key string) (int, bool) {
	val, ok := ann.Attributes[key]
	if !ok {
		return 0, false
	}

	if i, ok := val.(int); ok {
		return i, true
	}

	// 尝试从 float64 转换
	if f, ok := val.(float64); ok {
		return int(f), true
	}

	return 0, false
}

// GetBoolAttribute 获取布尔类型的属性值
//
// 参数:
//   - ann: 注解
//   - key: 属性键
//
// 返回值:
//   - bool: 属性值
//   - bool: 是否存在
func GetBoolAttribute(ann Annotation, key string) (bool, bool) {
	val, ok := ann.Attributes[key]
	if !ok {
		return false, false
	}

	if b, ok := val.(bool); ok {
		return b, true
	}

	return false, false
}
