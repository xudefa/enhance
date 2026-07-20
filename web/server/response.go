// Package server 提供 HTTP 服务器功能，用于 enhance 框架。
package server

// IsSuccess 判断响应是否为成功状态码 (2xx)。
func (r *HTTPResponse) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsRedirect 判断响应是否为重定向状态码 (3xx)。
func (r *HTTPResponse) IsRedirect() bool {
	return r.StatusCode >= 300 && r.StatusCode < 400
}

// IsClientError 判断响应是否为客户端错误状态码 (4xx)。
func (r *HTTPResponse) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError 判断响应是否为服务端错误状态码 (5xx)。
func (r *HTTPResponse) IsServerError() bool {
	return r.StatusCode >= 500 && r.StatusCode < 600
}

// Bind 绑定响应体到目标结构体。
func (r *HTTPResponse) Bind(v any) error {
	if len(r.Body) == 0 {
		return nil
	}
	return r.Unmarshal(v)
}

// String 获取响应体字符串。
func (r *HTTPResponse) String() string {
	return string(r.Body)
}
