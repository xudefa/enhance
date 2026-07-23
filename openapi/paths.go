package openapi

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
	Type any
}

// APISecurity API 安全注解
type APISecurity struct {
	// Name 安全方案名称
	Name string
	// Scopes 权限范围
	Scopes []string
}
