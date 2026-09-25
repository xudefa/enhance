// Package server 提供 HTTP 服务器功能，用于 enhance 框架。
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/xudefa/enhance/log"
)

// buildURL 构建完整 URL。
func (c *NetClient) buildURL(path string, query map[string][]string) string {
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		base := strings.TrimSuffix(c.baseURL, "/")
		path = strings.TrimPrefix(path, "/")
		path = base + "/" + path
	}

	if len(query) > 0 {
		separator := "?"
		if strings.Contains(path, "?") {
			separator = "&"
		}
		var sb strings.Builder
		first := true
		for k, v := range query {
			for _, val := range v {
				if !first {
					sb.WriteString("&")
				}
				sb.WriteString(url.QueryEscape(k))
				sb.WriteString("=")
				sb.WriteString(url.QueryEscape(val))
				first = false
			}
		}
		path = path + separator + sb.String()
	}

	return path
}

// httpCall 描述一次 HTTP 调用的方法、路径与请求体。
type httpCall struct {
	method string
	path   string
	body   any
}

// buildRequest 构建 HTTP 请求。
func (c *NetClient) buildRequest(ctx context.Context, call httpCall, cfg *HTTPRequest) (*http.Request, error) {
	reqBody, contentType, err := marshalRequestBody(call.body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, call.method, c.buildURL(call.path, cfg.Query), reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	c.mu.RLock()
	req.Header = c.headers.Clone()
	c.mu.RUnlock()

	if cfg.ContentType == "" {
		req.Header.Set("Content-Type", contentType)
	} else {
		req.Header.Set("Content-Type", cfg.ContentType)
	}

	applyRequestHeaders(req, cfg)

	return req, nil
}

// marshalRequestBody 根据 body 类型序列化为请求体并确定 Content-Type。
func marshalRequestBody(body any) (io.Reader, string, error) {
	if body == nil {
		return nil, "application/json", nil
	}
	switch bodyValue := body.(type) {
	case string:
		return strings.NewReader(bodyValue), "text/plain", nil
	case []byte:
		return bytes.NewReader(bodyValue), "application/octet-stream", nil
	case map[string][]string:
		return marshalFormBody(bodyValue), "application/x-www-form-urlencoded", nil
	default:
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, "", fmt.Errorf("marshal body failed: %w", err)
		}
		return bytes.NewReader(jsonBytes), "application/json", nil
	}
}

// marshalFormBody 将表单字段编码为 URL-encoded 请求体。
func marshalFormBody(values map[string][]string) io.Reader {
	var sb strings.Builder
	first := true
	for k, vals := range values {
		for _, val := range vals {
			if !first {
				sb.WriteString("&")
			}
			sb.WriteString(url.QueryEscape(k))
			sb.WriteString("=")
			sb.WriteString(url.QueryEscape(val))
			first = false
		}
	}
	return strings.NewReader(sb.String())
}

// applyRequestHeaders 将 HTTPRequest 配置中的 headers/auth 写入请求。
func applyRequestHeaders(req *http.Request, cfg *HTTPRequest) {
	for key, values := range cfg.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	if cfg.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.AuthToken)
	}
	if cfg.BasicAuth.Username != "" || cfg.BasicAuth.Password != "" {
		req.SetBasicAuth(cfg.BasicAuth.Username, cfg.BasicAuth.Password)
	}
}

// Get 发送 GET 请求。
func (c *NetClient) Get(ctx context.Context, path string, opts ...RequestOption) (*HTTPResponse, error) {
	return c.do(ctx, httpCall{method: "GET", path: path}, opts...)
}

// Head 发送 HEAD 请求。
func (c *NetClient) Head(ctx context.Context, path string, opts ...RequestOption) (*HTTPResponse, error) {
	return c.do(ctx, httpCall{method: "HEAD", path: path}, opts...)
}

// Post 发送 POST 请求。
func (c *NetClient) Post(ctx context.Context, path string, body any, opts ...RequestOption) (*HTTPResponse, error) {
	return c.do(ctx, httpCall{method: "POST", path: path, body: body}, opts...)
}

// Put 发送 PUT 请求。
func (c *NetClient) Put(ctx context.Context, path string, body any, opts ...RequestOption) (*HTTPResponse, error) {
	return c.do(ctx, httpCall{method: "PUT", path: path, body: body}, opts...)
}

// Patch 发送 PATCH 请求。
func (c *NetClient) Patch(ctx context.Context, path string, body any, opts ...RequestOption) (*HTTPResponse, error) {
	return c.do(ctx, httpCall{method: "PATCH", path: path, body: body}, opts...)
}

// Delete 发送 DELETE 请求。
func (c *NetClient) Delete(ctx context.Context, path string, opts ...RequestOption) (*HTTPResponse, error) {
	return c.do(ctx, httpCall{method: "DELETE", path: path}, opts...)
}

// Options 发送 OPTIONS 请求。
func (c *NetClient) Options(ctx context.Context, path string, opts ...RequestOption) (*HTTPResponse, error) {
	return c.do(ctx, httpCall{method: "OPTIONS", path: path}, opts...)
}

// do 执行 HTTP 请求并返回响应。
func (c *NetClient) do(ctx context.Context, call httpCall, opts ...RequestOption) (*HTTPResponse, error) {
	cfg := &HTTPRequest{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	req, err := c.buildRequest(ctx, call, cfg)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	return c.Do(ctx, req)
}

// Do 执行自定义 HTTP 请求并返回响应。
func (c *NetClient) Do(ctx context.Context, request any) (*HTTPResponse, error) {
	if c.httpClient == nil {
		return nil, fmt.Errorf("httpClient is nil")
	}
	req, ok := request.(*http.Request)
	if !ok {
		return nil, fmt.Errorf("invalid request type, expected *http.Request")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	body, err := readResponseBody(ctx, resp, c.logger)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	httpResp := &HTTPResponse{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       body,
	}

	c.mu.RLock()
	middleware := make([]ClientMiddlewareFunc, len(c.middleware))
	copy(middleware, c.middleware)
	c.mu.RUnlock()

	if err := applyClientMiddlewares(req, httpResp, middleware); err != nil {
		return nil, fmt.Errorf("apply client middlewares: %w", err)
	}

	return httpResp, nil
}

// readResponseBody 读取响应体，检查大小限制并关闭 Body。
func readResponseBody(ctx context.Context, resp *http.Response, logger log.Logger) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(resp.Body, DefaultMaxResponseBodySize+1))
	if err != nil {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Error(ctx, "close response body failed",
				log.KeyValue{Key: "close_error", Value: closeErr.Error()},
				log.KeyValue{Key: "read_error", Value: err.Error()},
			)
		}
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if len(body) > DefaultMaxResponseBodySize {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Error(ctx, "close response body failed",
				log.KeyValue{Key: "close_error", Value: closeErr.Error()},
			)
		}
		return nil, fmt.Errorf("response body too large: max %d bytes", DefaultMaxResponseBodySize)
	}

	if err := resp.Body.Close(); err != nil {
		return nil, fmt.Errorf("close response body failed: %w", err)
	}

	return body, nil
}

// applyClientMiddlewares 对响应依次执行客户端中间件。
func applyClientMiddlewares(req *http.Request, resp *HTTPResponse, middlewares []ClientMiddlewareFunc) error {
	for _, m := range middlewares {
		if err := m(req, resp); err != nil {
			return fmt.Errorf("execute client middleware: %w", err)
		}
	}
	return nil
}

// Unmarshal 反序列化 JSON 数据到指定目标。
//
// 参数:
//   - target: 指向目标结构体的指针
//
// 返回:
//   - error: 反序列化错误
func (r *HTTPResponse) Unmarshal(target any) error {
	if len(r.Body) == 0 {
		return nil
	}
	return json.Unmarshal(r.Body, target)
}
