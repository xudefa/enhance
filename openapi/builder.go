package openapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

// RegisterController 注册控制器
func (b *DocumentBuilder) RegisterController(controller any) *DocumentBuilder {
	t := reflect.TypeOf(controller)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return b
	}

	// 提取基础路径
	basePath := b.extractBasePath(t)
	tagName := b.extractTagName(t)

	// 遍历方法
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		b.registerMethod(method, basePath, tagName)
	}

	return b
}

// AddPath 手动添加路径
func (b *DocumentBuilder) AddPath(path string, method string, operation OperationObject) *DocumentBuilder {
	pathItem, exists := b.doc.Paths[path]
	if !exists {
		pathItem = PathItem{}
	}

	switch strings.ToLower(method) {
	case "get":
		pathItem.Get = &operation
	case "post":
		pathItem.Post = &operation
	case "put":
		pathItem.Put = &operation
	case "delete":
		pathItem.Delete = &operation
	case "patch":
		pathItem.Patch = &operation
	}

	b.doc.Paths[path] = pathItem
	return b
}

// Build 构建文档
func (b *DocumentBuilder) Build() *OpenAPIDocument {
	// 排序路径
	sortedPaths := make(map[string]PathItem)
	keys := make([]string, 0, len(b.doc.Paths))
	for k := range b.doc.Paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sortedPaths[k] = b.doc.Paths[k]
	}
	b.doc.Paths = sortedPaths

	// 排序标签
	sort.Slice(b.doc.Tags, func(i, j int) bool {
		return b.doc.Tags[i].Name < b.doc.Tags[j].Name
	})

	return b.doc
}

// ToJSON 转换为 JSON
func (b *DocumentBuilder) ToJSON() (string, error) {
	doc := b.Build()
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToJSONBytes 转换为 JSON 字节
func (b *DocumentBuilder) ToJSONBytes() ([]byte, error) {
	doc := b.Build()
	return json.MarshalIndent(doc, "", "  ")
}

// SaveToFile 保存到文件
func (b *DocumentBuilder) SaveToFile(path string) error {
	data, err := b.ToJSONBytes()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ServeHTTP 实现 http.Handler，提供 OpenAPI JSON 端点
func (b *DocumentBuilder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data, err := b.ToJSONBytes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(data)
}

// extractBasePath 提取基础路径
func (b *DocumentBuilder) extractBasePath(t reflect.Type) string {
	// 从类型名称推断
	name := t.Name()
	name = strings.TrimSuffix(name, "Controller")
	name = strings.TrimSuffix(name, "Handler")

	if name == "" {
		return "/"
	}

	return "/" + strings.ToLower(name)
}

// extractTagName 提取标签名称
func (b *DocumentBuilder) extractTagName(t reflect.Type) string {
	name := t.Name()
	name = strings.TrimSuffix(name, "Controller")
	name = strings.TrimSuffix(name, "Handler")
	return name
}

// registerMethod 注册方法
func (b *DocumentBuilder) registerMethod(method reflect.Method, basePath, tagName string) {
	// 跳过未导出方法
	if !method.IsExported() {
		return
	}

	// 提取操作信息
	operation := b.extractOperation(method, tagName)

	// 提取路径
	path := b.extractPath(method, basePath)
	httpMethod := b.extractHTTPMethod(method)

	if path == "" || httpMethod == "" {
		return
	}

	// 添加到文档
	b.AddPath(path, httpMethod, operation)
}

// extractOperation 提取操作信息
func (b *DocumentBuilder) extractOperation(method reflect.Method, tagName string) OperationObject {
	operation := OperationObject{
		Summary:     method.Name,
		Description: "",
		OperationID: method.Name,
		Tags:        []string{tagName},
		Responses:   make(map[string]ResponseObject),
	}

	// 检查方法标签
	if tag, ok := reflect.StructTag("").Lookup("operation"); ok {
		parts := strings.Split(tag, ",")
		for _, part := range parts {
			if strings.HasPrefix(part, "summary=") {
				operation.Summary = strings.TrimPrefix(part, "summary=")
			} else if strings.HasPrefix(part, "description=") {
				operation.Description = strings.TrimPrefix(part, "description=")
			} else if strings.HasPrefix(part, "operationId=") {
				operation.OperationID = strings.TrimPrefix(part, "operationId=")
			} else if part == "deprecated" {
				operation.Deprecated = true
			}
		}
	}

	// 添加默认响应
	operation.Responses["200"] = ResponseObject{
		Description: "Successful operation",
	}

	return operation
}

// extractPath 提取路径
func (b *DocumentBuilder) extractPath(method reflect.Method, basePath string) string {
	// 检查方法标签
	if tag, ok := reflect.StructTag("").Lookup("path"); ok {
		return basePath + tag
	}

	// 从方法名推断
	name := method.Name
	name = strings.TrimPrefix(name, "Get")
	name = strings.TrimPrefix(name, "Post")
	name = strings.TrimPrefix(name, "Put")
	name = strings.TrimPrefix(name, "Delete")
	name = strings.TrimPrefix(name, "Patch")

	if name == "" {
		return basePath
	}

	return basePath + "/" + strings.ToLower(name)
}

// extractHTTPMethod 提取 HTTP 方法
func (b *DocumentBuilder) extractHTTPMethod(method reflect.Method) string {
	name := method.Name

	if strings.HasPrefix(name, "Get") || strings.HasPrefix(name, "Find") || strings.HasPrefix(name, "List") {
		return "GET"
	} else if strings.HasPrefix(name, "Post") || strings.HasPrefix(name, "Create") || strings.HasPrefix(name, "Add") {
		return "POST"
	} else if strings.HasPrefix(name, "Put") || strings.HasPrefix(name, "Update") {
		return "PUT"
	} else if strings.HasPrefix(name, "Delete") || strings.HasPrefix(name, "Remove") {
		return "DELETE"
	} else if strings.HasPrefix(name, "Patch") {
		return "PATCH"
	}

	return ""
}

// generateSchema 生成 Schema
func (b *DocumentBuilder) generateSchema(t reflect.Type) SchemaObject {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := SchemaObject{
		Type:       "object",
		Properties: make(map[string]SchemaObject),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldSchema := b.generateFieldSchema(field.Type)
		fieldSchema.Description = field.Tag.Get("description")

		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			name := parts[0]
			if name != "-" {
				schema.Properties[name] = fieldSchema
				if contains(parts, "omitempty") {
					continue
				}
				schema.Required = append(schema.Required, name)
				continue
			}
		}
		schema.Properties[field.Name] = fieldSchema
		schema.Required = append(schema.Required, field.Name)
	}

	return schema
}

// generateFieldSchema 生成字段 Schema
func (b *DocumentBuilder) generateFieldSchema(t reflect.Type) SchemaObject {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Bool:
		return SchemaObject{Type: "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return SchemaObject{Type: "integer", Format: "int64"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return SchemaObject{Type: "integer", Format: "int64"}
	case reflect.Float32, reflect.Float64:
		return SchemaObject{Type: "number", Format: "double"}
	case reflect.String:
		return SchemaObject{Type: "string"}
	case reflect.Slice, reflect.Array:
		return SchemaObject{
			Type:  "array",
			Items: &SchemaObject{Type: b.mapTypeToString(t.Elem())},
		}
	case reflect.Map:
		return SchemaObject{
			Type: "object",
			AdditionalProperties: &SchemaObject{
				Type: b.mapTypeToString(t.Elem()),
			},
		}
	case reflect.Struct:
		return SchemaObject{
			Type:       "object",
			Properties: make(map[string]SchemaObject),
		}
	default:
		return SchemaObject{Type: "object"}
	}
}

// mapTypeToString 映射类型到字符串
func (b *DocumentBuilder) mapTypeToString(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "string"
	default:
		return "object"
	}
}

// contains 检查切片是否包含元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
