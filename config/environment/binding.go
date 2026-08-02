package environment

import (
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"unicode"
)

// Bind 将整个环境配置绑定到目标结构体
func (e *Environment) Bind(target any) error {
	val := reflectValueOf(target)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}
	return e.bindStruct(val.Elem(), "")
}

// BindKey 将指定键的值绑定到目标
func (e *Environment) BindKey(key string, target any) error {
	val := reflectValueOf(target)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}
	v, ok := e.GetProperty(key)
	if !ok {
		return fmt.Errorf("property not found: %s", key)
	}
	converted, err := globalTypeConverter.ConvertTo(v, val.Elem().Type())
	if err != nil {
		return fmt.Errorf("failed to convert property %q: %w", key, err)
	}
	val.Elem().Set(converted)
	return nil
}

// BindPrefix 将指定前缀的配置绑定到目标结构体
func (e *Environment) BindPrefix(prefix string, target any) error {
	val := reflectValueOf(target)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}
	return e.bindStruct(val.Elem(), prefix+".")
}

// Validate 验证配置，返回所有错误
func (e *Environment) Validate() []error {
	sources := e.GetPropertySources()
	errs := make([]error, 0, len(sources))
	for _, src := range sources {
		if v, ok := src.(interface{ Validate() []error }); ok {
			errs = append(errs, v.Validate()...)
		}
	}
	return errs
}

// ResolvePlaceholders 解析 ${...} 占位符（公有 API）
//
// 支持语法：
//   - ${key} — 引用配置参数 key
//   - ${key:defaultValue} — 引用 key，不存在时使用 defaultValue
func (e *Environment) ResolvePlaceholders(val string) string {
	return e.resolvePlaceholders(val, make(map[string]bool))
}

// parsePlaceholder 解析 ${...} 占位符内容
//
// 从 val[startIdx] 处的 '$' 开始解析，找到匹配的 '}'，
// 提取占位符内的 key 和 defaultValue。
//
// 参数:
//   - val: 原始字符串
//   - startIdx: '$' 字符的位置
//
// 返回:
//   - key: 占位符中的配置键名
//   - defaultValue: 冒号后的默认值（无默认值时返回空字符串）
//   - hasDefault: 是否存在默认值
//   - endIdx: 匹配的 '}' 的位置
//   - ok: 是否成功解析
func (e *Environment) parsePlaceholder(val string, startIdx int) (key, defaultValue string, hasDefault bool, endIdx int, ok bool) {
	depth := 1
	j := startIdx + 2
	for j < len(val) && depth > 0 {
		if val[j] == '$' && j+1 < len(val) && val[j+1] == '{' {
			depth++
			j += 2
			continue
		}
		if val[j] == '}' {
			depth--
			if depth == 0 {
				break
			}
		}
		j++
	}
	if depth != 0 {
		return "", "", false, 0, false
	}

	inner := val[startIdx+2 : j]

	keyEnd := -1
	nd := 0
	for k := 0; k < len(inner); k++ {
		if inner[k] == '$' && k+1 < len(inner) && inner[k+1] == '{' {
			nd++
			k++
			continue
		}
		if inner[k] == '}' {
			if nd > 0 {
				nd--
			}
			continue
		}
		if inner[k] == ':' && nd == 0 {
			keyEnd = k
			break
		}
	}

	if keyEnd >= 0 {
		return inner[:keyEnd], inner[keyEnd+1:], true, j, true
	}
	return inner, "", false, j, true
}

// resolvePlaceholders 内部占位符解析，带循环检测
//
// 支持语法：
//   - ${key} — 引用配置参数 key
//   - ${key:defaultValue} — 引用 key，不存在时使用 defaultValue
//     defaultValue 中支持嵌套 ${...} 占位符
func (e *Environment) resolvePlaceholders(val string, resolving map[string]bool) string {
	var result strings.Builder
	i := 0
	for i < len(val) {
		if val[i] == '$' && i+1 < len(val) && val[i+1] == '{' {
			key, defaultVal, hasDefault, j, ok := e.parsePlaceholder(val, i)
			if !ok {
				result.WriteByte(val[i])
				i++
				continue
			}

			if resolving[key] {
				slog.Warn("[environment] circular placeholder reference", "key", key)
				result.WriteString(val[i : j+1])
				i = j + 1
				continue
			}

			var replacement string
			if rawVal, rawOk := e.getRawProperty(key); rawOk {
				resolving[key] = true
				if s, ok := rawVal.(string); ok {
					replacement = e.resolvePlaceholders(s, resolving)
				} else {
					replacement = e.resolvePlaceholders(fmt.Sprintf("%v", rawVal), resolving)
				}
				delete(resolving, key)
			} else if hasDefault {
				resolving[key] = true
				replacement = e.resolvePlaceholders(defaultVal, resolving)
				delete(resolving, key)
			} else {
				replacement = val[i : j+1]
			}

			result.WriteString(replacement)
			i = j + 1
			continue
		}
		result.WriteByte(val[i])
		i++
	}
	return result.String()
}

func (e *Environment) bindStruct(val reflect.Value, prefix string) error {
	typ := val.Type()
	if typ.Kind() != reflect.Struct {
		return nil
	}
	for i := range typ.NumField() {
		field := typ.Field(i)
		fieldVal := val.Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		// 优先处理 value tag（支持 ${...} 占位符）
		if valueTag, ok := field.Tag.Lookup("value"); ok {
			if err := e.bindValueTag(fieldVal, valueTag); err != nil {
				return err
			}
			continue
		}

		// 确定配置键名
		key, hasExplicitKey := e.resolveConfigKey(field)
		fullKey := prefix + key
		if fieldVal.Kind() == reflect.Struct {
			// 嵌套结构体：如果字段有显式 key，检查嵌套字段是否也有显式 key
			// 如果嵌套字段有显式完整路径（如 db.url），则不添加父前缀
			if hasExplicitKey && hasNestedExplicitKeys(fieldVal) {
				// 嵌套字段已有完整路径，不添加父前缀
				if err := e.bindStruct(fieldVal, ""); err != nil {
					return err
				}
			} else if hasExplicitKey {
				if err := e.bindStruct(fieldVal, key+"."); err != nil {
					return err
				}
			} else {
				if err := e.bindStruct(fieldVal, fullKey+"."); err != nil {
					return err
				}
			}
			continue
		}
		if pv, ok := e.GetProperty(fullKey); ok {
			if err := setField(fieldVal, pv); err != nil {
				slog.Warn("[environment] set field failed", "key", fullKey, "error", err)
			}
		} else if defaultVal, ok := field.Tag.Lookup("default"); ok {
			if err := setField(fieldVal, defaultVal); err != nil {
				slog.Warn("[environment] set default failed", "key", fullKey, "error", err)
			}
		}
	}
	return nil
}

// bindValueTag 处理 value tag，支持 ${key} 和 ${key:defaultValue} 语法
func (e *Environment) bindValueTag(fieldVal reflect.Value, valueTag string) error {
	// 解析占位符
	resolved := e.ResolvePlaceholders(valueTag)

	// 将解析后的值设置到字段
	converted, err := globalTypeConverter.ConvertTo(resolved, fieldVal.Type())
	if err != nil {
		return fmt.Errorf("failed to convert value tag '%s' to %s: %w", valueTag, fieldVal.Type(), err)
	}
	fieldVal.Set(converted)
	return nil
}

// BindProperties 绑定配置属性到结构体（支持 value tag）
//
// 支持以下 tag：
//   - value:"${key}" — 从配置中获取 key 的值
//   - value:"${key:default}" — 从配置中获取 key 的值，不存在时使用默认值
//   - mapstructure:"key" — 从配置中获取 key 的值（无占位符）
//   - env:"KEY" — 从环境变量获取
//
// 示例：
//
//	type ServerConfig struct {
//	    Port    int           `value:"${server.port:8080}"`
//	    Host    string        `value:"${server.host:0.0.0.0}"`
//	    Timeout time.Duration `value:"${server.timeout:30s}"`
//	}
//
//	var config ServerConfig
//	env.BindProperties(&config)
func (e *Environment) BindProperties(target any) error {
	return e.Bind(target)
}

var globalTypeConverter = NewTypeConverter()

// toConfigKey 将 PascalCase 字段名转换为 config key（如 SSL → ssl, ServerConfig → server.config）
func toConfigKey(name string) string {
	var result strings.Builder
	for i, r := range name {
		if unicode.IsUpper(r) {
			// 在连续大写后接小写时插入分隔符（如 HTTPServer → http.server）
			if i > 0 {
				prev := rune(name[i-1])
				if !unicode.IsUpper(prev) || (i+1 < len(name) && unicode.IsLower(rune(name[i+1]))) {
					result.WriteRune('.')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func setField(fieldVal reflect.Value, val any) error {
	targetType := fieldVal.Type()
	converted, err := globalTypeConverter.ConvertTo(val, targetType)
	if err != nil {
		return fmt.Errorf("set field %s: %w", targetType, err)
	}
	fieldVal.Set(converted)
	return nil
}

func reflectValueOf(v any) reflect.Value {
	rv, ok := v.(reflect.Value)
	if ok {
		return rv
	}
	return reflect.ValueOf(v)
}

// resolveConfigKey 解析配置键名
func (e *Environment) resolveConfigKey(field reflect.StructField) (string, bool) {
	if tag, ok := field.Tag.Lookup("config"); ok {
		return tag, true
	}
	if tag, ok := field.Tag.Lookup("mapstructure"); ok {
		return tag, true
	}
	if tag, ok := field.Tag.Lookup("env"); ok {
		return tag, true
	}
	// 默认使用 snake_case 转换
	return toConfigKey(field.Name), false
}

// hasNestedExplicitKeys 检查嵌套结构体的所有导出字段是否都有显式 key
//
// 只有当全部导出字段都有显式 key 时才返回 true，否则该嵌套结构体不应跳过父前缀，
// 避免部分字段（无显式 key）因跳过前缀而解析到错误路径。
func hasNestedExplicitKeys(val reflect.Value) bool {
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}
		if !hasExplicitConfigKey(field) {
			return false
		}
	}
	return true
}

// hasExplicitConfigKey 检查字段是否带显式配置 key tag
func hasExplicitConfigKey(field reflect.StructField) bool {
	if _, hasConfig := field.Tag.Lookup("config"); hasConfig {
		return true
	}
	if _, hasMapstructure := field.Tag.Lookup("mapstructure"); hasMapstructure {
		return true
	}
	if _, hasEnv := field.Tag.Lookup("env"); hasEnv {
		return true
	}
	if _, hasValue := field.Tag.Lookup("value"); hasValue {
		return true
	}
	return false
}
