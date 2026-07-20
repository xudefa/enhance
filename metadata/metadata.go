// Package metadata 提供元数据管理功能，用于 enhance 框架。
package metadata

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// ToJSON 将元数据转换为 JSON 字符串。
func (m *ConfigurationMetadata) ToJSON() (string, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ==================== MetadataGenerator 实现 ====================

// metadataGeneratorImpl MetadataGenerator 接口的默认实现。
type metadataGeneratorImpl struct {
	groups     []GroupMetadata
	properties []PropertyMetadata
	hints      []HintMetadata
}

// NewMetadataGenerator 创建元数据生成器。
func NewMetadataGenerator() MetadataGenerator {
	return &metadataGeneratorImpl{}
}

// Register 注册配置结构体。
func (g *metadataGeneratorImpl) Register(config any) MetadataGenerator {
	t := reflect.TypeOf(config)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return g
	}

	groupName := g.extractGroupName(t.Name())

	group := GroupMetadata{
		Name:       groupName,
		Type:       t.Name(),
		SourceType: t.String(),
	}
	g.groups = append(g.groups, group)

	g.extractProperties(t, groupName)

	return g
}

// WithHint 添加配置提示。
func (g *metadataGeneratorImpl) WithHint(propertyName string, values []HintValue) MetadataGenerator {
	hint := HintMetadata{
		Name:   propertyName,
		Values: values,
	}
	g.hints = append(g.hints, hint)
	return g
}

// WithHintProvider 添加提示提供者。
func (g *metadataGeneratorImpl) WithHintProvider(propertyName string, providerName string, params map[string]string) MetadataGenerator {
	var hint *HintMetadata
	for i := range g.hints {
		if g.hints[i].Name == propertyName {
			hint = &g.hints[i]
			break
		}
	}

	if hint == nil {
		g.hints = append(g.hints, HintMetadata{Name: propertyName})
		hint = &g.hints[len(g.hints)-1]
	}

	hint.Providers = append(hint.Providers, HintProvider{
		Name:       providerName,
		Parameters: params,
	})

	return g
}

// Generate 生成配置元数据。
func (g *metadataGeneratorImpl) Generate() *ConfigurationMetadata {
	sort.Slice(g.groups, func(i, j int) bool {
		return g.groups[i].Name < g.groups[j].Name
	})
	sort.Slice(g.properties, func(i, j int) bool {
		return g.properties[i].Name < g.properties[j].Name
	})
	sort.Slice(g.hints, func(i, j int) bool {
		return g.hints[i].Name < g.hints[j].Name
	})

	return &ConfigurationMetadata{
		Groups:     g.groups,
		Properties: g.properties,
		Hints:      g.hints,
	}
}

// extractProperties 提取结构体的属性元数据。
func (g *metadataGeneratorImpl) extractProperties(t reflect.Type, groupName string) {
	for i := range t.NumField() {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			g.extractProperties(field.Type, groupName)
			continue
		}

		configTag := field.Tag.Get("config")
		if configTag == "" {
			configTag = groupName + "." + g.camelToKebab(field.Name)
		}

		description := field.Tag.Get("description")
		defaultValue := field.Tag.Get("default")
		required := field.Tag.Get("required") == "true"
		secret := field.Tag.Get("secret") == "true"
		deprecated := field.Tag.Get("deprecated") == "true"
		deprecationReason := field.Tag.Get("deprecationReason")

		if strings.Contains(strings.ToLower(field.Name), "password") ||
			strings.Contains(strings.ToLower(field.Name), "secret") ||
			strings.Contains(strings.ToLower(field.Name), "token") {
			secret = true
		}

		prop := PropertyMetadata{
			Name:              configTag,
			Type:              g.mapType(field.Type),
			Description:       description,
			DefaultValue:      defaultValue,
			SourceType:        t.Name(),
			Required:          required,
			Secret:            secret,
			Deprecated:        deprecated,
			DeprecationReason: deprecationReason,
		}

		g.properties = append(g.properties, prop)
	}
}

// extractGroupName 从结构体名称提取组名（驼峰转小写）。
func (g *metadataGeneratorImpl) extractGroupName(structName string) string {
	if len(structName) == 0 {
		return ""
	}

	name := strings.TrimSuffix(structName, "Config")
	name = strings.TrimSuffix(name, "Properties")

	return g.camelToKebab(name)
}

// camelToKebab 驼峰命名转短横线命名。
func (g *metadataGeneratorImpl) camelToKebab(s string) string {
	var result strings.Builder
	for i, c := range s {
		if i > 0 && c >= 'A' && c <= 'Z' {
			prevRune := rune(s[i-1])
			if prevRune >= 'a' && prevRune <= 'z' {
				result.WriteByte('-')
			} else if i+1 < len(s) && s[i+1] >= 'a' && s[i+1] <= 'z' {
				result.WriteByte('-')
			}
		}
		result.WriteRune(c)
	}
	return strings.ToLower(result.String())
}

// mapType 映射 Go 类型到配置类型字符串。
func (g *metadataGeneratorImpl) mapType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "java.lang.Boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "java.lang.Integer"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "java.lang.Integer"
	case reflect.Float32, reflect.Float64:
		return "java.lang.Float"
	case reflect.String:
		return "java.lang.String"
	case reflect.Slice, reflect.Array:
		return "java.util.List<" + g.mapType(t.Elem()) + ">"
	case reflect.Map:
		return "java.util.Map<java.lang.String, " + g.mapType(t.Elem()) + ">"
	case reflect.Struct:
		return t.Name()
	case reflect.Ptr:
		return g.mapType(t.Elem())
	default:
		return t.String()
	}
}

// ==================== PropertyIndex 实现 ====================

// propertyIndexImpl PropertyIndex 接口的默认实现。
type propertyIndexImpl struct {
	properties map[string]PropertyMetadata
}

// NewPropertyIndex 创建属性索引。
func NewPropertyIndex(metadata *ConfigurationMetadata) PropertyIndex {
	index := &propertyIndexImpl{
		properties: make(map[string]PropertyMetadata),
	}

	for _, prop := range metadata.Properties {
		index.properties[prop.Name] = prop
	}

	return index
}

// Get 获取属性元数据。
func (idx *propertyIndexImpl) Get(name string) (PropertyMetadata, bool) {
	prop, ok := idx.properties[name]
	return prop, ok
}

// Has 检查属性是否存在。
func (idx *propertyIndexImpl) Has(name string) bool {
	_, ok := idx.properties[name]
	return ok
}

// GetAll 获取所有属性。
func (idx *propertyIndexImpl) GetAll() []PropertyMetadata {
	var props []PropertyMetadata
	for _, prop := range idx.properties {
		props = append(props, prop)
	}

	sort.Slice(props, func(i, j int) bool {
		return props[i].Name < props[j].Name
	})

	return props
}

// GetByPrefix 按前缀获取属性。
func (idx *propertyIndexImpl) GetByPrefix(prefix string) []PropertyMetadata {
	var props []PropertyMetadata
	for _, prop := range idx.properties {
		if strings.HasPrefix(prop.Name, prefix) {
			props = append(props, prop)
		}
	}

	sort.Slice(props, func(i, j int) bool {
		return props[i].Name < props[j].Name
	})

	return props
}

// ==================== 工具函数 ====================

// GenerateFromStruct 从结构体类型直接生成元数据。
func GenerateFromStruct(configs ...any) (*ConfigurationMetadata, error) {
	gen := NewMetadataGenerator()
	for _, config := range configs {
		gen.Register(config)
	}
	return gen.Generate(), nil
}

// GenerateJSON 从结构体生成 JSON 格式的元数据。
func GenerateJSON(configs ...any) (string, error) {
	metadata, err := GenerateFromStruct(configs...)
	if err != nil {
		return "", err
	}
	return metadata.ToJSON()
}

// ValidateProperty 验证属性值。
func ValidateProperty(name string, value string, metadata PropertyMetadata) error {
	if metadata.Required && value == "" {
		return fmt.Errorf("property %q is required", name)
	}

	if metadata.Secret && value != "" {
		return fmt.Errorf("property %q contains sensitive information and should not be exposed", name)
	}

	return nil
}
