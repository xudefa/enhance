package openapi

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
