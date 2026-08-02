// Package validator 提供数据验证自动配置。
//
// Validator 是 Go 语言最流行的数据验证库，支持结构体标签验证。
// 本模块提供自动配置支持，内置常用验证规则和自定义验证器注册。
//
// 功能特性：
//   - 自动配置验证器
//   - 支持结构体标签验证
//   - 自定义验证器支持
//   - 内置常用验证规则（手机号、身份证号）
//   - 支持上下文取消
//
// 配置示例：
//
//	{
//	  "validator": {
//	    "enabled": true,
//	    "enable-custom-validators": true
//	  }
//	}
//
// 使用示例：
//
//	type User struct {
//	    Name  string `validate:"required,min=3,max=50"`
//	    Email string `validate:"required,email"`
//	    Phone string `validate:"required,phone"`
//	}
//
//	v := core.MustGetBean[*validator.Validate](app.Container())
//	err := v.Struct(user)
package validator

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

// init 注册 Validator 自动配置类。
// 当配置 validator.enabled=true 时自动触发配置。
func init() {
	boot.RegisterAutoConfigWith(&ValidatorAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(ValidatorEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityInfrastructure)),
	)
}

// ValidatorAutoConfiguration 数据验证自动配置类。
// 负责初始化 validator.Validate 实例并注册到 IoC 容器。
type ValidatorAutoConfiguration struct {
	logger   log.Logger          // 日志记录器
	validate *validator.Validate // 验证器实例
	config   *ValidatorConfig    // 验证器配置
}

// Configure 配置数据验证器。
// 创建 validator.Validate 实例，注册自定义验证器，并注册到 IoC 容器。
func (c *ValidatorAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	// 获取日志记录器，如果不存在则使用默认日志器
	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	// 加载配置
	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load Validator config: %w", err)
	}

	c.config = cfg
	c.validate = validator.New()

	// 注册自定义验证器（手机号、身份证号等）
	if cfg.EnableCustomValidators {
		c.registerCustomValidators(ctx.Context(), c.logger)
	}

	// 注册验证器实例到 IoC 容器
	if err := ctx.Container().RegisterInstance(c.validate, reflect.TypeFor[*validator.Validate]()); err != nil {
		return fmt.Errorf("failed to register Validator instance: %w", err)
	}

	c.logger.Info(ctx.Context(), "Validator configured")

	return nil
}

// GetValidator 获取验证器实例。
// 返回底层的 *validator.Validate 实例，可用于高级验证操作。
func (c *ValidatorAutoConfiguration) GetValidator() *validator.Validate {
	return c.validate
}

// Validate 验证数据结构。
// 使用结构体标签进行验证，返回验证错误。
//
// 使用示例：
//
//	type User struct {
//	    Name  string `validate:"required,min=3"`
//	    Email string `validate:"required,email"`
//	}
//	err := validator.Validate(user)
func (c *ValidatorAutoConfiguration) Validate(s interface{}) error {
	return c.validate.Struct(s)
}

// ValidateVar 验证单个变量。
// 对单个字段进行验证，适用于简单场景。
//
// 使用示例：
//
//	err := validator.ValidateVar("test@example.com", "email")
func (c *ValidatorAutoConfiguration) ValidateVar(field interface{}, tag string) error {
	return c.validate.Var(field, tag)
}

// ValidateStructCtx 带上下文的验证数据结构。
// 支持上下文取消，适用于长时间运行的验证场景。
//
// 使用示例：
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	err := validator.ValidateStructCtx(ctx, user)
func (c *ValidatorAutoConfiguration) ValidateStructCtx(ctx context.Context, s interface{}) error {
	return c.validate.StructCtx(ctx, s)
}

// ValidateVarCtx 带上下文的验证单个变量。
// 支持上下文取消，适用于长时间运行的验证场景。
//
// 使用示例：
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	err := validator.ValidateVarCtx(ctx, "test@example.com", "email")
func (c *ValidatorAutoConfiguration) ValidateVarCtx(ctx context.Context, field interface{}, tag string) error {
	return c.validate.VarCtx(ctx, field, tag)
}

// RegisterValidation 注册自定义验证器。
// 允许用户注册自己的验证函数，扩展验证规则。
//
// 使用示例：
//
//	validator.RegisterValidation("custom", func(fl validator.FieldLevel) bool {
//	    // 自定义验证逻辑
//	    return true
//	})
func (c *ValidatorAutoConfiguration) RegisterValidation(name string, fn validator.Func) error {
	return c.validate.RegisterValidation(name, fn)
}

// registerCustomValidators 注册内置自定义验证器。
// 注册手机号和身份证号等常用验证规则。
func (c *ValidatorAutoConfiguration) registerCustomValidators(ctx context.Context, logger log.Logger) {
	// 注册手机号验证器（中国大陆手机号）
	// 验证规则：11 位数字，前缀为 13x-19x
	if err := c.validate.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		if len(phone) != 11 {
			return false
		}
		// 验证所有位均为数字
		for i := 0; i < len(phone); i++ {
			if phone[i] < '0' || phone[i] > '9' {
				return false
			}
		}
		// 验证手机号前缀（13x, 14x, 15x, 16x, 17x, 18x, 19x）
		prefix := phone[:2]
		validPrefixes := []string{"13", "14", "15", "16", "17", "18", "19"}
		for _, p := range validPrefixes {
			if prefix == p {
				return true
			}
		}
		return false
	}); err != nil {
		logger.Warn(ctx, "failed to register phone validator", log.KeyValue{Key: "error", Value: err.Error()})
	}

	// 注册身份证号验证器（18 位）
	// 验证规则：前 17 位为数字，最后一位为数字或 X
	if err := c.validate.RegisterValidation("idcard", func(fl validator.FieldLevel) bool {
		idcard := fl.Field().String()
		if len(idcard) != 18 {
			return false
		}
		// 验证前 17 位为数字
		for i := 0; i < 17; i++ {
			if idcard[i] < '0' || idcard[i] > '9' {
				return false
			}
		}
		// 最后一位可以是数字或 X（校验码）
		lastChar := idcard[17]
		return (lastChar >= '0' && lastChar <= '9') || lastChar == 'X' || lastChar == 'x'
	}); err != nil {
		logger.Warn(ctx, "failed to register idcard validator", log.KeyValue{Key: "error", Value: err.Error()})
	}
}

// ValidatorConfig 数据验证配置。
// 包含验证器的所有可配置参数。
type ValidatorConfig struct {
	Enabled                bool `json:"enabled" mapstructure:"enabled"`                                   // 是否启用验证器
	EnableCustomValidators bool `json:"enable-custom-validators" mapstructure:"enable-custom-validators"` // 是否启用自定义验证器
}

// 配置常量。
const (
	ValidatorEnabled              = "validator.enabled" // 启用条件配置键
	DefaultEnableCustomValidators = true                // 默认启用自定义验证器
	ConditionTrue                 = "true"              // 条件真值
)

// loadConfig 从 Environment 加载 Validator 配置。
// 使用默认值初始化配置，然后从配置中心绑定用户自定义值。
func (c *ValidatorAutoConfiguration) loadConfig(env *environment.Environment) (*ValidatorConfig, error) {
	cfg := &ValidatorConfig{
		EnableCustomValidators: DefaultEnableCustomValidators,
	}

	if err := env.BindPrefix("validator", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind Validator config: %w", err)
	}

	return cfg, nil
}
