# security 包 — 安全认证与授权

> **所属层级**: Infrastructure Layer  
> **设计理念**: 接口设计，灵活集成  
> **设计灵感**: Spring Security

## 概述

`security` 包提供全面的安全认证和授权功能，采用接口设计，方便快速集成。支持认证、授权、密码编码、过滤器链、Web 表达式等特性。

### 核心功能

| 功能 | 说明 |
|------|------|
| **认证** | 支持用户名密码、JWT、Basic 等多种认证方式 |
| **授权** | 基于角色和权限的访问控制 |
| **密码编码** | BCrypt、SHA256 等多种密码编码器 |
| **过滤器链** | 可配置的安全过滤器链 |
| **Web 表达式** | hasRole、hasAuthority 等表达式支持 |
| **用户管理** | 内存用户、DAO 用户详情服务 |
| **CORS** | 跨域资源共享支持 |
| **限流** | 请求速率限制 |

---

## 核心接口

### Authentication 接口

```go
type Authentication interface {
    GetPrincipal() any
    GetCredentials() any
    GetAuthorities() []string
    IsAuthenticated() bool
}
```

### AuthenticationManager 接口

```go
type AuthenticationManager interface {
    Authenticate(authentication Authentication) (Authentication, error)
}
```

### UserDetailsService 接口

```go
type UserDetailsService interface {
    LoadUserByUsername(username string) (UserDetails, error)
}
```

### PasswordEncoder 接口

```go
type PasswordEncoder interface {
    Encode(rawPassword string) string
    Matches(rawPassword, encodedPassword string) bool
}
```

---

## 快速开始

### 基本配置

```go
package main

import (
    "net/http"
    "github.com/xudefa/enhance/security"
)

func main() {
    // 创建用户详情服务
    userDetailsService := security.NewInMemoryUserDetailsService()
    userDetailsService.CreateUser("admin", "hashed_password", []string{"ROLE_ADMIN", "ROLE_USER"})

    // 创建密码编码器
    passwordEncoder := security.NewBCryptPasswordEncoder()
    
    // 创建认证提供者和管理器
    authProvider := security.NewDaoAuthenticationProvider(userDetailsService, passwordEncoder)
    authManager := security.NewProviderManager(authProvider)

    // 配置 HTTP 安全
    sec := security.NewHttpSecurity()
    chain, err := sec.
        AuthenticationManager(authManager).
        UserDetailsService(userDetailsService).
        PasswordEncoder(passwordEncoder).
        Anonymous().
        AuthorizeRequests(func(reg security.AuthorizeRequests) {
            reg.AntMatchers("/public/**").PermitAll()
            reg.AntMatchers("/admin/**").HasRole("ADMIN")
            reg.AnyRequest().Authenticated()
        }).
        Build()
    if err != nil {
        panic(err)
    }

    // 集成 HTTP 服务器
    handler := security.NewSecurityFilterChainHandler(chain, yourAppHandler)
    http.ListenAndServe(":8080", handler)
}
```

---

## 认证授权执行流程

### 完整请求处理流程（从进入到出去）

当一个HTTP请求到达时，安全模块按以下流程处理：

```
HTTP请求 → SecurityFilterChainHandler → 过滤器链 → 业务Handler → HTTP响应
```

#### 详细执行步骤

**阶段1：请求进入（SecurityFilterChainHandler.ServeHTTP）**

1. `SecurityFilterChainHandler.ServeHTTP(w, r)` - 接收HTTP请求
2. `NewHttpRequestAdapter(r)` - 将http.Request适配为SecurityRequest
3. `NewHttpResponseAdapter(w)` - 将http.ResponseWriter适配为SecurityResponse
4. `securityFilterChain.DoFilter(ctx, request, response)` - 开始执行过滤器链

**阶段2：过滤器链执行（FilterChainProxy）**

5. `FilterChainProxy.DoFilter()` - 调用`doFilterInternal(ctx, request, response, 0)`
6. 按索引顺序执行每个过滤器，每个过滤器通过`chain.DoFilter()`触发下一个

**阶段3：各过滤器执行（按配置顺序）**

7. **AuthContextFilter** - 安全上下文处理
   - `DoFilter()` - 保存并清除全局认证信息
   - 执行后续过滤器链
   - 请求结束后恢复认证信息并清理

8. **AnonymousAuthenticationFilter** - 匿名认证（如启用）
   - `DoFilter()` - 检查是否已认证
   - 未认证则创建`AnonymousAuthenticationToken`并设置到全局上下文
   - 继续执行后续过滤器链

9. **CORS过滤器** - 跨域处理（如启用）
   - `CorsFilter.DoFilter()` - 检查Origin头
   - 设置CORS响应头
   - 如果是OPTIONS预检请求，直接返回204

10. **限流过滤器** - 请求限流（如启用）
    - `RateLimitFilter.DoFilter()` - 检查令牌桶
    - 超过限流则返回429
    - 否则继续执行后续过滤器链

11. **认证过滤器** - 身份验证（根据配置选择）
    
    **Basic认证**：
    - `BasicAuthenticationFilter.DoFilter()` - 检查Authorization头
    - 提取用户名密码，创建`UsernamePasswordAuthenticationToken`
    - 调用`AuthenticationManager.Authenticate()`进行认证
    - 认证成功后设置到全局安全上下文
    
    **表单登录认证**：
    - `UsernamePasswordAuthenticationFilter.DoFilter()` - 检查POST请求和登录URL
    - 提取用户名密码，创建认证令牌
    - 调用`AuthenticationManager.Authenticate()`进行认证
    - 认证成功后重定向到成功URL
    
    **JWT认证**（jwt模块提供）：
    - 从请求头提取JWT Token
    - 验证Token有效性
    - 认证成功后设置到全局安全上下文

12. **ExceptionTranslationFilter** - 异常处理
    - `DoFilter()` - 执行后续过滤器链并捕获异常
    - 如果抛出`ErrAccessDenied`：
      - 未认证用户：调用`AuthenticationEntryPoint.Commence()`返回401或重定向
      - 已认证用户：调用`AccessDeniedHandler.Handle()`返回403

13. **FilterSecurityInterceptor** - 授权拦截
    - `DoFilter()` - 执行访问决策检查
    - `SecurityMetadataSource.GetAttributes()` - 获取URL对应的权限表达式
    - 如果已处理过该请求，直接放行
    - `AccessDecisionManager.Decide()` - 进行权限决策
    - 决策通过则继续，否则抛出`ErrAccessDenied`

**阶段4：权限决策流程（AccessDecisionManager）**

14. `AffirmativeBased.Decide()` - 访问决策（默认使用肯定优先策略）
15. 遍历所有投票者（Voter）进行投票：
    - `WebExpressionVoter.Vote()` - 解析Web表达式（permitAll、hasRole等）
    - `AuthenticatedVoter.Vote()` - 检查认证级别
    - `RoleVoter.Vote()` - 检查角色权限
16. 根据投票结果决定是否允许访问

**阶段5：请求处理完成**

17. 所有过滤器执行完毕，调用`DefaultSecurityFilterChain.DoFilter()`
18. 返回到`SecurityFilterChainHandler.ServeHTTP()`
19. 检查响应状态码，如果>=400则直接返回
20. 从请求属性中获取认证信息：`request.GetAttribute("security.currentAuthentication")`
21. 将认证信息注入request context：`ContextWithAuthentication(r.Context(), auth)`
22. 调用业务Handler：`nextHandler.ServeHTTP(w, r)`

**阶段6：业务处理**

23. 业务Handler从context获取认证信息：`GetAuthenticationFromContext(ctx)`
24. 执行业务逻辑
25. 返回HTTP响应

#### 认证流程详解

**Basic认证流程**：
```
请求携带Authorization: Basic base64(username:password)
  ↓
BasicAuthenticationFilter.DoFilter()
  ↓
提取用户名密码 → NewUsernamePasswordAuthenticationToken(username, password)
  ↓
AuthenticationManager.Authenticate(token)
  ↓
ProviderManager.Authenticate() → 遍历所有Provider
  ↓
DaoAuthenticationProvider.Authenticate()
  ↓
UserDetailsService.LoadUserByUsername() → 加载用户详情
  ↓
PasswordEncoder.Matches() → 验证密码
  ↓
检查用户状态（Enabled、AccountNonLocked等）
  ↓
认证成功 → NewAuthenticatedUsernamePasswordAuthenticationToken(user, authorities)
  ↓
SetAuthentication(authenticated) → 设置到全局安全上下文
```

**表单登录认证流程**：
```
POST /login 携带username和password
  ↓
UsernamePasswordAuthenticationFilter.DoFilter()
  ↓
提取用户名密码 → NewUsernamePasswordAuthenticationToken(username, password)
  ↓
AuthenticationManager.Authenticate(token)
  ↓
（后续流程同Basic认证）
  ↓
认证成功 → 302重定向到defaultSuccessUrl
```

#### 授权流程详解

**URL权限检查流程**：
```
FilterSecurityInterceptor.DoFilter()
  ↓
SecurityMetadataSource.GetAttributes(request)
  ↓
根据URL匹配规则获取权限表达式
  例如："/admin/**" → ["hasRole('ADMIN')"]
  ↓
AccessDecisionManager.Decide(auth, request, attributes)
  ↓
遍历所有Voter进行投票：
  - WebExpressionVoter: 解析"hasRole('ADMIN')"表达式
    ↓
    检查用户authorities是否包含"ROLE_ADMIN"
    ↓
    返回ACCESS_GRANTED或ACCESS_DENIED
  ↓
根据投票策略决定：
  - AffirmativeBased: 有一个GRANTED就通过
  - UnanimousBased: 所有都不DENIED才通过
  - ConsensusBased: 多数票决定
  ↓
通过 → 继续执行后续过滤器
拒绝 → 抛出ErrAccessDenied异常
```

#### 过滤器执行顺序示例

默认过滤器链顺序（从上到下执行）：
1. `AuthContextFilter` - 安全上下文管理
2. `AnonymousAuthenticationFilter` - 匿名认证
3. `CsrfFilter`（如启用）- CSRF防护
4. `LogoutFilter`（如配置）- 登出处理
5. `UsernamePasswordAuthenticationFilter`（如启用）- 表单登录
6. `BasicAuthenticationFilter`（如启用）- Basic认证
7. `ExceptionTranslationFilter` - 异常转换
8. `FilterSecurityInterceptor` - 授权拦截
9. `DefaultSecurityFilterChain` - 默认链（放行到业务Handler）

---

## API 参考

### 认证组件

| 组件 | 说明 |
|------|------|
| `Authentication` | 认证信息接口，包含 principal、credentials、authorities |
| `AuthenticationManager` | 认证管理器，处理认证请求 |
| `AuthenticationProvider` | 认证提供者接口，定义认证逻辑 |
| `UsernamePasswordAuthenticationToken` | 用户名密码认证令牌 |

### 授权组件

| 组件 | 说明 |
|------|------|
| `AccessDecisionManager` | 访问决策管理器，决定资源访问权限 |
| `AccessDecisionVoter` | 访问决策投票者，投票决定访问权限 |
| `WebExpressionVoter` | Web 表达式投票者，支持 hasRole、hasAuthority 等 |
| `RoleVoter` | 角色投票者，检查角色权限 |
| `AuthenticatedVoter` | 认证投票者，检查认证级别 |

### 访问决策策略

| 策略 | 说明 |
|------|------|
| `AffirmativeBased` | 肯定优先，只要有投票者授予权限就通过 |
| `UnanimousBased` | 一致通过，所有投票者都不拒绝才通过 |
| `ConsensusBased` | 多数通过，多数投票者授予权限才通过 |

### 过滤器

| 过滤器 | 说明 |
|--------|------|
| `FilterChainProxy` | 过滤器链代理 |
| `AuthContextFilter` | 安全上下文过滤器 |
| `AnonymousAuthenticationFilter` | 匿名认证过滤器 |
| `ExceptionTranslationFilter` | 异常转换过滤器 |
| `FilterSecurityInterceptor` | 过滤器安全拦截器 |
| `BasicAuthenticationFilter` | Basic 认证过滤器 |
| `JwtAuthenticationFilter` | JWT 认证过滤器（jwt 模块提供） |

#### 过滤器执行顺序

1. CORS 过滤器（如启用）
2. 限流过滤器（如启用）
3. JWT 认证过滤器（如启用 jwt 模块）
4. AuthContext 过滤器
5. Anonymous 认证过滤器
6. ExceptionTranslation 过滤器
7. FilterSecurityInterceptor

### 密码编码器

| 编码器 | 说明 |
|--------|------|
| `NoOpPasswordEncoder` | 不编码，用于开发测试 |
| `BCryptPasswordEncoder` | SHA256 密码编码器 |
| `StandardPasswordEncoder` | 标准密码编码器，支持密钥加盐 |
| `DelegatingPasswordEncoder` | 委托密码编码器，支持多编码器 |

### Web 表达式

| 表达式 | 说明 |
|-------|------|
| `permitAll` | 允许所有人访问 |
| `denyAll` | 拒绝所有人访问 |
| `authenticated` | 仅允许已认证用户 |
| `hasRole('ROLE')` | 检查是否具有指定角色 |
| `hasAnyRole('ROLE1','ROLE2')` | 检查是否具有任一角色 |
| `hasAuthority('AUTHORITY')` | 检查是否具有指定权限 |
| `hasAnyAuthority('AUTH1','AUTH2')` | 检查是否具有任一权限 |

### 安全上下文

```go
// 线程安全的认证信息存储
ctx := security.GetSecurityContext()
ctx.SetAuthentication(auth)
auth := ctx.Authentication()
ctx.ClearAuthentication()

// 全局认证快捷方法
security.SetAuthentication(auth)
security.GetAuthentication()
security.ClearAuthentication()
```

---

## 使用示例

### URL 权限配置

```go
security.AuthorizeRequests(func(reg security.AuthorizeRequests) {
    reg.AntMatchers("/public/**").PermitAll()
    reg.AntMatchers("/admin/**").HasRole("ADMIN")
    reg.AntMatchers("/user/**").HasAnyRole("ADMIN", "USER")
    reg.AntMatchers("/api/**").HasAuthority("ACCESS_API")
    reg.AnyRequest().Authenticated()
})
```

### 错误处理

```go
// 错误类型
security.ErrAuthenticationFailed  // 认证失败
security.ErrAccessDenied          // 访问被拒绝
security.ErrUserNotFound          // 用户不存在
security.ErrBadCredentials        // 凭证错误
```

### 自动配置

通过 `condition.OnProperty("security.enabled", "true")` 控制，启用时自动配置以下组件：

```yaml
# application.yaml
security:
  enabled: true
  cors:
    enabled: false
    allowed-origins: "*"
    allowed-methods: "GET,POST,PUT,DELETE,OPTIONS"
    allowed-headers: "Content-Type,Authorization,X-Requested-With"
    allow-credentials: false
    max-age: 3600
  rate-limit:
    enabled: false
    rate: 100          # 每秒请求数
    burst: 200         # 突发请求数
    exclude-paths: "/health,/actuator/health"
  rules: "/public/**->permitAll,/admin/**->hasRole('ADMIN'),/api/**->authenticated"
  login-url: "/login"
```

#### 自动配置注册的 Bean

| Bean ID | 说明 | 是否可自定义 |
|---------|------|-------------|
| `userDetailsService` | 用户详情服务 | 是 |
| `passwordEncoder` | 密码编码器（BCrypt） | 是 |
| `authenticationManager` | 认证管理器 | 否 |
| `securityFilterChain` | 安全过滤器链 | 是 |

#### 配置选项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `security.enabled` | 是否启用安全模块 | `false` |
| `security.cors.enabled` | 是否启用 CORS | `false` |
| `security.cors.allowed-origins` | 允许的源 | `*` |
| `security.cors.allowed-methods` | 允许的方法 | `GET,POST,PUT,DELETE,OPTIONS` |
| `security.cors.allowed-headers` | 允许的请求头 | `Content-Type,Authorization,X-Requested-With` |
| `security.cors.allow-credentials` | 是否允许凭证 | `false` |
| `security.cors.max-age` | 预检请求缓存时间（秒） | `3600` |
| `security.rate-limit.enabled` | 是否启用限流 | `false` |
| `security.rate-limit.rate` | 限流速率（每秒） | `100` |
| `security.rate-limit.burst` | 突发容量 | `200` |
| `security.rate-limit.exclude-paths` | 排除路径 | `/health,/actuator/health` |
| `security.rules` | URL 安全规则 | 空 |
| `security.login-url` | 登录页面 URL | `/login` |

### 与 JWT 模块集成

当同时启用安全模块和 JWT 模块时，JWT 认证过滤器会自动集成到安全过滤器链中：

```yaml
# application.yaml
security:
  enabled: true
jwt:
  enabled: true
  secret-key: my-secret-key-123456
  exclude-paths: /login,/register
```

**集成效果**：
- JWT 过滤器会自动添加到安全过滤器链中
- 支持配置排除路径（不需要 JWT 认证的路径）
- 认证成功后，用户信息自动设置到安全上下文

---

## 最佳实践

### 1. 使用 BCrypt 密码编码器

```go
// ✅ 推荐：使用 BCrypt 编码密码
passwordEncoder := security.NewBCryptPasswordEncoder()
hashedPassword := passwordEncoder.Encode("user_password")

// ⚠️ 不推荐：使用 NoOpPasswordEncoder（仅用于开发测试）
passwordEncoder := security.NewNoOpPasswordEncoder()
```

### 2. 合理配置 URL 权限

```go
// ✅ 推荐：按最小权限原则配置 URL 权限
security.AuthorizeRequests(func(reg security.AuthorizeRequests) {
    reg.AntMatchers("/public/**", "/health").PermitAll()
    reg.AntMatchers("/admin/**").HasRole("ADMIN")
    reg.AntMatchers("/api/**").Authenticated()
    reg.AnyRequest().DenyAll()
})

// ⚠️ 不推荐：过度授权
security.AuthorizeRequests(func(reg security.AuthorizeRequests) {
    reg.AnyRequest().PermitAll()
})
```

### 3. 启用 CORS 和限流保护

```go
// ✅ 推荐：生产环境启用 CORS 和限流
security:
  enabled: true
  cors:
    enabled: true
    allowed-origins: "https://yourdomain.com"
    allowed-methods: "GET,POST"
  rate-limit:
    enabled: true
    rate: 100
    burst: 200
```

### 4. 使用 JWT 认证

```go
// ✅ 推荐：使用 JWT 进行无状态认证
jwt:
  enabled: true
  secret-key: my-secret-key-123456
  exclude-paths: /login,/register

// ⚠️ 不推荐：使用 Session 认证（有状态）
```

### 5. 自定义用户详情服务

```go
// ✅ 推荐：实现自定义 UserDetailsService
type CustomUserDetailsService struct {
    db *sql.DB
}

func (s *CustomUserDetailsService) LoadUserByUsername(username string) (security.UserDetails, error) {
    // 从数据库加载用户
    return user, nil
}

// ⚠️ 不推荐：仅使用内存用户详情服务（生产环境）
userDetailsService := security.NewInMemoryUserDetailsService()
```