package casbin

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/security"
)

func init() {
	boot.RegisterAutoConfigWith(&CasbinAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(security.SecurityEnabled, security.ConditionTrue),
			condition.OnProperty(security.CasbinEnabled, security.ConditionTrue),
			condition.OnProperty(security.CasbinPolicyType, security.DefaultCasbinPolicyType),
		),
		boot.WithOrder(int(boot.OrderPriorityAuthorization)), // 授权层，在认证之后、安全核心之前执行
	)
}

// CasbinAutoConfiguration Casbin 自动配置类。
//
// 当配置文件中启用 Casbin 时自动生效（security.enabled=true 且 security.casbin.enabled=true）。
// 负责注册 CasbinEnforcer 和 CasbinVoter 到 IoC 容器中。
//
// 执行顺序：Order = -1200（授权层），在 CasbinGorm（-1300）之后执行。
//
// 依赖关系：
//  1. 在认证层（-1500）之后执行（依赖认证信息）
//  2. 在安全核心层（-100）之前执行（为安全过滤器链提供投票器）
//  3. 如果 casbin-gorm（-1300）已注册 CasbinEnforcer，则直接使用，否则创建默认实例
type CasbinAutoConfiguration struct {
	logger log.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

// CasbinConfig Casbin 授权配置。
type CasbinConfig struct {
	Enabled          bool   `json:"enabled" mapstructure:"enabled"`
	ModelType        string `json:"model-type" mapstructure:"model-type"`
	ModelPath        string `json:"model-path" mapstructure:"model-path"`
	ModelText        string `json:"model-text" mapstructure:"model-text"`
	PolicyType       string `json:"policy-type" mapstructure:"policy-type"`
	PolicyPath       string `json:"policy-path" mapstructure:"policy-path"`
	PolicyText       string `json:"policy-text" mapstructure:"policy-text"`
	AutoLoad         bool   `json:"auto-load" mapstructure:"auto-load"`
	AutoLoadInterval int    `json:"auto-load-interval" mapstructure:"auto-load-interval"`
}

// Configure 配置 Casbin 授权。
//
// 该方法在自动配置阶段调用，负责：
//  1. 从 Environment 中读取 Casbin 配置（模型路径、策略路径等）
//  2. 检查容器中是否已有 CasbinEnforcer 实例，如果有则直接使用
//  3. 如果没有，则创建并注册 DefaultCasbinEnforcer（用于权限检查）
//  4. 根据配置加载模型和策略（支持文件、字符串等方式）
//  5. 创建并注册 CasbinVoter（实现 AccessDecisionVoter 接口）
//  6. 如果启用自动加载，启动定时刷新策略
//
// CasbinVoter 会被 SecurityAutoConfiguration 自动检测到并添加到 AccessDecisionManager 中，
// 无需手动注册。
func (c *CasbinAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	container := ctx.Container()
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](container, ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	c.logger.Info(ctx.Context(), "configuring Casbin authorization...")

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load Casbin config: %w", err)
	}

	if err := c.validateConfig(cfg); err != nil {
		return fmt.Errorf("Casbin config validation failed: %w", err)
	}

	var enforcer security.CasbinEnforcer
	if beans, err := container.Get(reflect.TypeFor[security.CasbinEnforcer]()); err == nil && len(beans) > 0 {
		enforcer, _ = beans[0].(security.CasbinEnforcer)
		c.logger.Info(ctx.Context(), "using existing CasbinEnforcer instance from container")
	} else {
		// 根据配置选择模型和策略路径
		modelPath := cfg.ModelPath
		if modelPath == "" {
			modelPath = security.DefaultCasbinModelPath
		}
		policyPath := cfg.PolicyPath
		if policyPath == "" {
			policyPath = security.DefaultCasbinPolicyPath
		}

		created, err := NewCasbinEnforcer(ctx.Context(), c.logger, modelPath, policyPath)
		if err != nil {
			return fmt.Errorf("failed to create CasbinEnforcer: %w", err)
		}
		enforcer = created
		c.logger.Info(ctx.Context(), "creating default CasbinEnforcer instance",
			log.KeyValue{Key: "model-path", Value: modelPath},
			log.KeyValue{Key: "policy-path", Value: policyPath},
		)
	}

	if err := container.RegisterInstance(enforcer, reflect.TypeFor[security.CasbinEnforcer]()); err != nil {
		return fmt.Errorf("failed to register CasbinEnforcer: %w", err)
	}
	// 同时注册具体类型，方便直接获取 DefaultCasbinEnforcer 使用其扩展方法
	if defaultEnforcer, ok := enforcer.(*DefaultCasbinEnforcer); ok {
		if err := container.RegisterInstance(defaultEnforcer, reflect.TypeFor[*DefaultCasbinEnforcer]()); err != nil {
			return fmt.Errorf("failed to register DefaultCasbinEnforcer: %w", err)
		}
	}
	c.logger.Info(ctx.Context(), "CasbinEnforcer registered")

	voter, err := security.NewCasbinVoter(enforcer)
	if err != nil {
		return fmt.Errorf("failed to create CasbinVoter: %w", err)
	}
	if err := container.RegisterInstance(voter, reflect.TypeFor[security.CasbinVoter]()); err != nil {
		return fmt.Errorf("failed to register CasbinVoter: %w", err)
	}
	c.logger.Info(ctx.Context(), "CasbinVoter registered")

	admBeans, err := container.Get(reflect.TypeFor[security.AccessDecisionManager]())
	if err == nil && len(admBeans) > 0 {
		c.logger.Info(ctx.Context(), "AccessDecisionManager detected, CasbinVoter ready")
	}

	if cfg.AutoLoad {
		c.ctx = ctx.Context()
		c.startAutoReload(enforcer, cfg)
	}

	c.logger.Info(ctx.Context(), "Casbin configuration complete",
		log.KeyValue{Key: "model-type", Value: cfg.ModelType},
		log.KeyValue{Key: "policy-type", Value: cfg.PolicyType},
	)

	return nil
}

// validateConfig 验证 Casbin 配置。
func (c *CasbinAutoConfiguration) validateConfig(cfg *CasbinConfig) error {
	switch cfg.ModelType {
	case "file":
		if cfg.ModelPath == "" {
			return fmt.Errorf("model-path must not be empty when model-type is file")
		}
	case "string":
		if cfg.ModelText == "" {
			return fmt.Errorf("model-text must not be empty when model-type is string")
		}
	default:
		return fmt.Errorf("unsupported model-type: %s, supported values: file, string", cfg.ModelType)
	}

	switch cfg.PolicyType {
	case "file":
		if cfg.PolicyPath == "" {
			return fmt.Errorf("policy-path must not be empty when policy-type is file")
		}
	case "string":
		if cfg.PolicyText == "" {
			return fmt.Errorf("policy-text must not be empty when policy-type is string")
		}
	default:
		return fmt.Errorf("unsupported policy-type: %s, supported values: file, string", cfg.PolicyType)
	}

	return nil
}

// startAutoReload 启动自动重新加载策略。
func (c *CasbinAutoConfiguration) startAutoReload(enforcer security.CasbinEnforcer, cfg *CasbinConfig) {
	interval := cfg.AutoLoadInterval
	if interval <= 0 {
		interval = security.DefaultCasbinAutoLoadInterval
	}

	c.logger.Info(c.ctx, "starting Casbin policy auto-reload",
		log.KeyValue{Key: "interval", Value: fmt.Sprintf("%d min", interval)},
	)

	go func() {
		defer recoverLog("casbin policy auto-reload", c.ctx, c.logger)
		ticker := time.NewTicker(time.Duration(interval) * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := enforcer.LoadPolicy(c.ctx); err != nil {
					c.logger.Warn(c.ctx, "failed to auto-reload Casbin policy",
						log.KeyValue{Key: "error", Value: err.Error()},
					)
					continue
				}
				c.logger.Info(c.ctx, "Casbin policy auto-reloaded")
			case <-c.ctx.Done():
				return
			}
		}
	}()
}

// Close 停止自动刷新定时器，释放 goroutine 资源。
func (c *CasbinAutoConfiguration) Close() {
	if c.cancel != nil {
		c.cancel()
	}
}

// recoverLog recovers from panic and logs the error.
func recoverLog(component string, ctx context.Context, logger log.Logger) {
	if r := recover(); r != nil {
		logger.Error(ctx, fmt.Sprintf("%s panic recovered", component),
			log.KeyValue{Key: "panic", Value: fmt.Sprintf("%v", r)},
		)
	}
}

// loadConfig 从 Environment 加载 Casbin 配置。
func (c *CasbinAutoConfiguration) loadConfig(env *environment.Environment) (*CasbinConfig, error) {
	cfg := &CasbinConfig{
		ModelType:        security.DefaultCasbinModelType,
		ModelPath:        security.DefaultCasbinModelPath,
		PolicyType:       security.DefaultCasbinPolicyType,
		PolicyPath:       security.DefaultCasbinPolicyPath,
		AutoLoad:         security.DefaultCasbinAutoLoad,
		AutoLoadInterval: security.DefaultCasbinAutoLoadInterval,
	}

	if err := env.BindPrefix("security.casbin", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind Casbin config: %w", err)
	}

	return cfg, nil
}
