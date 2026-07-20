// Package server 提供 HTTP 服务器功能，用于 enhance 框架。
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		q := ""
		for k, v := range query {
			for _, val := range v {
				if q != "" {
					q += "&"
				}
				q += k + "=" + val
			}
		}
		path = path + separator + q
	}

	return path
}

// buildRequest 构建 HTTP 请求。
func (c *NetClient) buildRequest(ctx context.Context, method, path string, body any, cfg *HTTPRequest) (*http.Request, error) {
	var reqBody io.Reader
	contentType := "application/json"

	if body != nil {
		switch v := body.(type) {
		case string:
			reqBody = strings.NewReader(v)
			contentType = "text/plain"
		case []byte:
			reqBody = bytes.NewReader(v)
			contentType = "application/octet-stream"
		case map[string][]string:
			q := ""
			for k, vals := range v {
				for _, val := range vals {
					if q != "" {
						q += "&"
					}
					q += k + "=" + val
				}
			}
			reqBody = strings.NewReader(q)
			contentType = "application/x-www-form-urlencoded"
		default:
			data, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("marshal body failed: %w", err)
			}
			reqBody = bytes.NewReader(data)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.buildURL(path, cfg.Query), reqBody)
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

	return req, nil
}

// Get 发送 GET 请求。
func (c *NetClient) Get(ctx context.Context, path string, opts ...RequestOption) (*HTTPResponse, error) {
	return c.do(ctx, "GET", path, nil, opts...)
}

// Head 发送 HEAD 请求。
func (c *NetClient) Head(ctx context.Context, path string, opts ...RequestOption) (*HTTPResponse, error) {
	return c.do(ctx, "HEAD", path, nil, opts...)
}

// Post 发送 POST 请求。
func (c *NetClient) Post(ctx context.Context, path string, body any, opts ...RequestOption) (*HTTPResponse, error) {
	return c.do(ctx, "POST", path, body, opts...)
}

// Put 发送 PUT 请求。
func (c *NetClient) Put(ctx context.Context, path string, body any, opts ...RequestOption) (*HTTPResponse, error) {
	return c.do(ctx, "PUT", path, body, opts...)
}

// Patch 发送 PATCH 请求。
func (c *NetClient) Patch(ctx context.Context, path string, body any, opts ...RequestOption) (*HTTPResponse, error) {
	return c.do(ctx, "PATCH", path, body, opts...)
}

// Delete 发送 DELETE 请求。
func (c *NetClient) Delete(ctx context.Context, path string, opts ...RequestOption) (*HTTPResponse, error) {
	return c.do(ctx, "DELETE", path, nil, opts...)
}

// Options 发送 OPTIONS 请求。
func (c *NetClient) Options(ctx context.Context, path string, opts ...RequestOption) (*HTTPResponse, error) {
	return c.do(ctx, "OPTIONS", path, nil, opts...)
}

// do 执行 HTTP 请求并返回响应。
func (c *NetClient) do(ctx context.Context, method, path string, body any, opts ...RequestOption) (*HTTPResponse, error) {
	cfg := &HTTPRequest{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	req, err := c.buildRequest(ctx, method, path, body, cfg)
	if err != nil {
		return nil, err
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.mu.RLock()
			if len(c.middleware) > 0 {
				c.logger.Error(ctx, "close response body failed",
					log.KeyValue{Key: "close_error", Value: closeErr.Error()},
					log.KeyValue{Key: "read_error", Value: err.Error()},
				)
			}
			c.mu.RUnlock()
		}
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if err := resp.Body.Close(); err != nil {
		return nil, fmt.Errorf("close response body failed: %w", err)
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

	for _, m := range middleware {
		if err := m(req, httpResp); err != nil {
			return nil, err
		}
	}

	return httpResp, nil
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
