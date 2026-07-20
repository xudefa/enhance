package security

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/xudefa/enhance/log"
)

// UsernamePasswordAuthenticationFilter 用户名密码认证过滤器
type UsernamePasswordAuthenticationFilter struct {
	loginProcessingURL    string
	defaultSuccessURL     string
	failureURL            string
	authenticationManager AuthenticationManager
	logger                log.Logger
}

// NewUsernamePasswordAuthenticationFilterWithDefaults 创建带默认值的用户名密码认证过滤器
func NewUsernamePasswordAuthenticationFilterWithDefaults(
	loginProcessingURL,
	defaultSuccessURL,
	failureURL string,
	authManager AuthenticationManager,
	logger log.Logger,
) *UsernamePasswordAuthenticationFilter {
	return &UsernamePasswordAuthenticationFilter{
		loginProcessingURL:    loginProcessingURL,
		defaultSuccessURL:     defaultSuccessURL,
		failureURL:            failureURL,
		authenticationManager: authManager,
		logger:                logger,
	}
}

// DoFilter 处理用户名密码认证
func (f *UsernamePasswordAuthenticationFilter) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error {
	if request.GetMethod() != "POST" {
		return chain.DoFilter(ctx, request, response)
	}

	uri := request.GetURI()
	if uri != f.loginProcessingURL {
		return chain.DoFilter(ctx, request, response)
	}

	username := request.GetHeader("username")
	password := request.GetHeader("password")

	if username == "" || password == "" {
		f.logger.Warn(ctx, "认证失败：缺少用户名或密码")
		response.SetStatusCode(401)
		if writeErr := response.Write([]byte("missing username or password")); writeErr != nil {
			f.logger.Error(ctx, "写入认证失败响应失败", log.KeyValue{Key: "error", Value: writeErr.Error()})
		}
		return nil
	}

	f.logger.Debug(ctx, "尝试认证用户", log.KeyValue{Key: "username", Value: username})
	authToken := NewUsernamePasswordAuthenticationToken(username, password)
	authenticated, err := f.authenticationManager.Authenticate(ctx, authToken)

	if err != nil {
		f.logger.Warn(ctx, "用户认证失败", log.KeyValue{Key: "username", Value: username}, log.KeyValue{Key: "error", Value: err.Error()})
		response.SetStatusCode(401)
		response.SetHeader("Location", f.failureURL)
		if writeErr := response.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, err))); writeErr != nil {
			f.logger.Error(ctx, "写入认证错误响应失败", log.KeyValue{Key: "error", Value: writeErr.Error()})
		}
		return nil
	}

	f.logger.Info(ctx, "用户认证成功", log.KeyValue{Key: "username", Value: username})
	SetAuthentication(authenticated)

	response.SetStatusCode(302)
	response.SetHeader("Location", f.defaultSuccessURL)
	if writeErr := response.Write([]byte(`{"message":"login success"}`)); writeErr != nil {
		f.logger.Error(ctx, "写入登录成功响应失败", log.KeyValue{Key: "error", Value: writeErr.Error()})
	}

	return nil
}

// BasicAuthenticationEntryPointWithRealm Basic认证入口点（带Realm）
type BasicAuthenticationEntryPointWithRealm struct {
	realmName string
	logger    log.Logger
}

// NewBasicAuthenticationEntryPointWithRealm 创建带Realm的Basic认证入口点
func NewBasicAuthenticationEntryPointWithRealm(realmName string, logger log.Logger) *BasicAuthenticationEntryPointWithRealm {
	return &BasicAuthenticationEntryPointWithRealm{
		realmName: realmName,
		logger:    logger,
	}
}

// Commence 返回Basic认证失败响应
func (e *BasicAuthenticationEntryPointWithRealm) Commence(ctx context.Context, request SecurityRequest, response SecurityResponse, err error) error {
	e.logger.Debug(ctx, "Basic 认证入口点被触发", log.KeyValue{Key: "realm", Value: e.realmName})
	response.SetStatusCode(401)
	response.SetHeader("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, e.realmName))
	if writeErr := response.Write([]byte("Authentication required")); writeErr != nil {
		e.logger.Error(ctx, "写入认证响应失败", log.KeyValue{Key: "error", Value: writeErr.Error()})
	}

	if err == nil {
		return ErrBadCredentials
	}
	return err
}

// BasicAuthenticationFilterWithRealm 带Realm的Basic认证过滤器
type BasicAuthenticationFilterWithRealm struct {
	authenticationManager AuthenticationManager
	entryPoint            AuthenticationEntryPoint
	logger                log.Logger
}

// NewBasicAuthenticationFilterWithRealm 创建带Realm的Basic认证过滤器
func NewBasicAuthenticationFilterWithRealm(authManager AuthenticationManager, realmName string, logger log.Logger) *BasicAuthenticationFilterWithRealm {
	entryPoint := NewBasicAuthenticationEntryPointWithRealm(realmName, logger)
	return &BasicAuthenticationFilterWithRealm{
		authenticationManager: authManager,
		entryPoint:            entryPoint,
		logger:                logger,
	}
}

// DoFilter 处理Basic认证
func (f *BasicAuthenticationFilterWithRealm) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error {
	authHeader := request.GetHeader("Authorization")
	if authHeader == "" {
		f.logger.Debug(ctx, "请求缺少 Authorization 头")
		return f.entryPoint.Commence(ctx, request, response, ErrBadCredentials)
	}

	const prefix = "Basic "
	if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
		f.logger.Debug(ctx, "Authorization 头格式不正确，非 Basic 认证")
		return f.entryPoint.Commence(ctx, request, response, ErrBadCredentials)
	}

	encoded := authHeader[len(prefix):]
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		f.logger.Warn(ctx, "Basic 认证凭据解码失败", log.KeyValue{Key: "error", Value: err.Error()})
		return f.entryPoint.Commence(ctx, request, response, ErrBadCredentials)
	}

	credentials := string(decoded)
	sepIndex := -1
	for i, c := range credentials {
		if c == ':' {
			sepIndex = i
			break
		}
	}

	if sepIndex == -1 {
		f.logger.Warn(ctx, "Basic 认证凭据格式错误，缺少分隔符")
		return f.entryPoint.Commence(ctx, request, response, ErrBadCredentials)
	}

	username := credentials[:sepIndex]
	password := credentials[sepIndex+1:]

	f.logger.Debug(ctx, "尝试 Basic 认证", log.KeyValue{Key: "username", Value: username})
	authToken := NewUsernamePasswordAuthenticationToken(username, password)
	authenticated, err := f.authenticationManager.Authenticate(ctx, authToken)

	if err != nil {
		f.logger.Warn(ctx, "Basic 认证失败", log.KeyValue{Key: "username", Value: username}, log.KeyValue{Key: "error", Value: err.Error()})
		return f.entryPoint.Commence(ctx, request, response, err)
	}

	f.logger.Info(ctx, "Basic 认证成功", log.KeyValue{Key: "username", Value: username})
	SetAuthentication(authenticated)

	return chain.DoFilter(ctx, request, response)
}
