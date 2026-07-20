package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

// OpenAPIDocument OpenAPI 3.0 文档
type OpenAPIDocument struct {
	OpenAPI    string              `json:"openapi"`
	Info       InfoObject          `json:"info"`
	Servers    []ServerObject      `json:"servers,omitempty"`
	Paths      map[string]PathItem `json:"paths"`
	Components *ComponentsObject   `json:"components,omitempty"`
	Tags       []TagObject         `json:"tags,omitempty"`
}

// InfoObject 文档信息
type InfoObject struct {
	Title          string         `json:"title"`
	Version        string         `json:"version"`
	Description    string         `json:"description,omitempty"`
	TermsOfService string         `json:"termsOfService,omitempty"`
	Contact        *ContactObject `json:"contact,omitempty"`
	License        *LicenseObject `json:"license,omitempty"`
}

// ContactObject 联系信息
type ContactObject struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// LicenseObject 许可证信息
type LicenseObject struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// ServerObject 服务器信息
type ServerObject struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

// ServerVariable 服务器变量
type ServerVariable struct {
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

// PathItem 路径项
type PathItem struct {
	Summary     string            `json:"summary,omitempty"`
	Description string            `json:"description,omitempty"`
	Get         *OperationObject  `json:"get,omitempty"`
	Post        *OperationObject  `json:"post,omitempty"`
	Put         *OperationObject  `json:"put,omitempty"`
	Delete      *OperationObject  `json:"delete,omitempty"`
	Patch       *OperationObject  `json:"patch,omitempty"`
	Parameters  []ParameterObject `json:"parameters,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
}

// OperationObject 操作对象
type OperationObject struct {
	Summary     string                    `json:"summary,omitempty"`
	Description string                    `json:"description,omitempty"`
	OperationID string                    `json:"operationId,omitempty"`
	Tags        []string                  `json:"tags,omitempty"`
	Parameters  []ParameterObject         `json:"parameters,omitempty"`
	RequestBody *RequestBodyObject        `json:"requestBody,omitempty"`
	Responses   map[string]ResponseObject `json:"responses"`
	Deprecated  bool                      `json:"deprecated,omitempty"`
	Security    []map[string][]string     `json:"security,omitempty"`
}

// ParameterObject 参数对象
type ParameterObject struct {
	Name        string        `json:"name"`
	In          string        `json:"in"` // query, path, header, cookie
	Description string        `json:"description,omitempty"`
	Required    bool          `json:"required,omitempty"`
	Schema      *SchemaObject `json:"schema,omitempty"`
	Example     any           `json:"example,omitempty"`
}

// RequestBodyObject 请求体对象
type RequestBodyObject struct {
	Description string                     `json:"description,omitempty"`
	Required    bool                       `json:"required,omitempty"`
	Content     map[string]MediaTypeObject `json:"content"`
}

// MediaTypeObject 媒体类型对象
type MediaTypeObject struct {
	Schema   *SchemaObject            `json:"schema,omitempty"`
	Example  any                      `json:"example,omitempty"`
	Examples map[string]ExampleObject `json:"examples,omitempty"`
}

// ExampleObject 示例对象
type ExampleObject struct {
	Summary       string `json:"summary,omitempty"`
	Description   string `json:"description,omitempty"`
	Value         any    `json:"value,omitempty"`
	ExternalValue string `json:"externalValue,omitempty"`
}

// ResponseObject 响应对象
type ResponseObject struct {
	Description string                     `json:"description"`
	Headers     map[string]HeaderObject    `json:"headers,omitempty"`
	Content     map[string]MediaTypeObject `json:"content,omitempty"`
	Links       map[string]LinkObject      `json:"links,omitempty"`
}

// HeaderObject 头部对象
type HeaderObject struct {
	Description string        `json:"description,omitempty"`
	Schema      *SchemaObject `json:"schema,omitempty"`
}

// LinkObject 链接对象
type LinkObject struct {
	OperationRef string            `json:"operationRef,omitempty"`
	OperationID  string            `json:"operationId,omitempty"`
	Parameters   map[string]string `json:"parameters,omitempty"`
}

// SchemaObject Schema 对象
type SchemaObject struct {
	Type                 string                  `json:"type,omitempty"`
	Format               string                  `json:"format,omitempty"`
	Description          string                  `json:"description,omitempty"`
	Properties           map[string]SchemaObject `json:"properties,omitempty"`
	Required             []string                `json:"required,omitempty"`
	Items                *SchemaObject           `json:"items,omitempty"`
	AdditionalProperties *SchemaObject           `json:"additionalProperties,omitempty"`
	Enum                 []string                `json:"enum,omitempty"`
	Default              any                     `json:"default,omitempty"`
	Example              any                     `json:"example,omitempty"`
	Minimum              *float64                `json:"minimum,omitempty"`
	Maximum              *float64                `json:"maximum,omitempty"`
	MinLength            *int                    `json:"minLength,omitempty"`
	MaxLength            *int                    `json:"maxLength,omitempty"`
	Pattern              string                  `json:"pattern,omitempty"`
	Nullable             bool                    `json:"nullable,omitempty"`
	ReadOnly             bool                    `json:"readOnly,omitempty"`
	WriteOnly            bool                    `json:"writeOnly,omitempty"`
}

// ComponentsObject 组件对象
type ComponentsObject struct {
	Schemas         map[string]SchemaObject         `json:"schemas,omitempty"`
	Responses       map[string]ResponseObject       `json:"responses,omitempty"`
	Parameters      map[string]ParameterObject      `json:"parameters,omitempty"`
	RequestBodies   map[string]RequestBodyObject    `json:"requestBodies,omitempty"`
	SecuritySchemes map[string]SecuritySchemeObject `json:"securitySchemes,omitempty"`
}

// SecuritySchemeObject 安全方案对象
type SecuritySchemeObject struct {
	Type             string            `json:"type"`
	Description      string            `json:"description,omitempty"`
	Name             string            `json:"name,omitempty"`
	In               string            `json:"in,omitempty"`
	Scheme           string            `json:"scheme,omitempty"`
	BearerFormat     string            `json:"bearerFormat,omitempty"`
	Flows            *OAuthFlowsObject `json:"flows,omitempty"`
	OpenIDConnectURL string            `json:"openIdConnectUrl,omitempty"`
}

// OAuthFlowsObject OAuth 流程对象
type OAuthFlowsObject struct {
	Implicit          *OAuthFlowObject `json:"implicit,omitempty"`
	Password          *OAuthFlowObject `json:"password,omitempty"`
	ClientCredentials *OAuthFlowObject `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlowObject `json:"authorizationCode,omitempty"`
}

// OAuthFlowObject OAuth 流程对象
type OAuthFlowObject struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// TagObject 标签对象
type TagObject struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// APITag API 标签注解
type APITag struct {
	// Name 标签名称
	Name string
	// Description 标签描述
	Description string
}

// APIOperation API 操作注解
type APIOperation struct {
	// Summary 操作摘要
	Summary string
	// Description 操作描述
	Description string
	// OperationID 操作 ID
	OperationID string
	// Tags 标签列表
	Tags []string
	// Deprecated 是否已弃用
	Deprecated bool
}

// APIParam API 参数注解
type APIParam struct {
	// Name 参数名称
	Name string
	// In 参数位置 (query, path, header, cookie)
	In string
	// Description 参数描述
	Description string
	// Required 是否必填
	Required bool
	// Example 示例值
	Example any
}

// APIResponse API 响应注解
type APIResponse struct {
	// StatusCode HTTP 状态码
	StatusCode int
	// Description 响应描述
	Description string
	// Type 响应类型
	Type reflect.Type
}

// APISecurity API 安全注解
type APISecurity struct {
	// Name 安全方案名称
	Name string
	// Scopes 权限范围
	Scopes []string
}

// DocumentBuilder 文档构建器
type DocumentBuilder struct {
	doc        *OpenAPIDocument
	schemas    map[string]reflect.Type
	registered map[string]bool
}

// NewDocument 创建新的 OpenAPI 文档
func NewDocument() *DocumentBuilder {
	return &DocumentBuilder{
		doc: &OpenAPIDocument{
			OpenAPI: "3.0.3",
			Info: InfoObject{
				Title:   "API Documentation",
				Version: "1.0.0",
			},
			Paths: make(map[string]PathItem),
		},
		schemas:    make(map[string]reflect.Type),
		registered: make(map[string]bool),
	}
}

// SetInfo 设置文档信息
func (b *DocumentBuilder) SetInfo(title, version, description string) *DocumentBuilder {
	b.doc.Info = InfoObject{
		Title:       title,
		Version:     version,
		Description: description,
	}
	return b
}

// SetContact 设置联系信息
func (b *DocumentBuilder) SetContact(name, url, email string) *DocumentBuilder {
	b.doc.Info.Contact = &ContactObject{
		Name:  name,
		URL:   url,
		Email: email,
	}
	return b
}

// SetLicense 设置许可证信息
func (b *DocumentBuilder) SetLicense(name, url string) *DocumentBuilder {
	b.doc.Info.License = &LicenseObject{
		Name: name,
		URL:  url,
	}
	return b
}

// SetTermsOfService 设置服务条款
func (b *DocumentBuilder) SetTermsOfService(terms string) *DocumentBuilder {
	b.doc.Info.TermsOfService = terms
	return b
}

// AddServer 添加服务器
func (b *DocumentBuilder) AddServer(url, description string) *DocumentBuilder {
	b.doc.Servers = append(b.doc.Servers, ServerObject{
		URL:         url,
		Description: description,
	})
	return b
}

// AddTag 添加标签
func (b *DocumentBuilder) AddTag(name, description string) *DocumentBuilder {
	b.doc.Tags = append(b.doc.Tags, TagObject{
		Name:        name,
		Description: description,
	})
	return b
}

// AddSecurityScheme 添加安全方案
func (b *DocumentBuilder) AddSecurityScheme(name string, scheme SecuritySchemeObject) *DocumentBuilder {
	if b.doc.Components == nil {
		b.doc.Components = &ComponentsObject{
			SecuritySchemes: make(map[string]SecuritySchemeObject),
		}
	}
	b.doc.Components.SecuritySchemes[name] = scheme
	return b
}

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

// AddSchema 添加 Schema
func (b *DocumentBuilder) AddSchema(name string, schema SchemaObject) *DocumentBuilder {
	if b.doc.Components == nil {
		b.doc.Components = &ComponentsObject{
			Schemas: make(map[string]SchemaObject),
		}
	}
	b.doc.Components.Schemas[name] = schema
	return b
}

// RegisterSchema 注册结构体 Schema
func (b *DocumentBuilder) RegisterSchema(name string, typ reflect.Type) *DocumentBuilder {
	if _, exists := b.schemas[name]; exists {
		return b
	}

	b.schemas[name] = typ
	schema := b.generateSchema(typ)
	b.AddSchema(name, schema)

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

// ServeSwaggerUI 提供 Swagger UI 服务
func ServeSwaggerUI(doc *DocumentBuilder, basePath string, port int) error {
	mux := http.NewServeMux()

	// OpenAPI JSON 端点
	mux.Handle(basePath+"/openapi.json", doc)

	// Swagger UI HTML
	mux.HandleFunc(basePath+"/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprint(w, swaggerUIHTML(basePath+"/openapi.json"))
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Swagger UI available at http://localhost%s%s/\n", addr, basePath)
	fmt.Printf("OpenAPI JSON at http://localhost%s%s/openapi.json\n", addr, basePath)

	return http.ListenAndServe(addr, mux)
}

// swaggerUIHTML Swagger UI HTML 模板
func swaggerUIHTML(openAPIURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '%s',
                dom_id: '#swagger-ui',
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
                layout: 'BaseLayout'
            });
        };
    </script>
</body>
</html>`, openAPIURL)
}
