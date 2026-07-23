package openapi

import "reflect"

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
