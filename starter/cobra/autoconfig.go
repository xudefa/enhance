// Package cobra 提供 Cobra CLI 框架自动配置。
package cobra

import (
	"fmt"
	"os"
	"reflect"

	"github.com/spf13/cobra"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

func init() {
	boot.RegisterAutoConfigWith(&CobraAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(CobraEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityInfrastructure)),
	)
}

// CobraAutoConfiguration Cobra CLI 框架自动配置类。
type CobraAutoConfiguration struct {
	logger  log.Logger
	rootCmd *cobra.Command
	config  *CobraConfig
}

// Configure 配置 Cobra CLI。
func (c *CobraAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load Cobra config: %w", err)
	}

	c.config = cfg

	c.rootCmd = &cobra.Command{
		Use:     cfg.Use,
		Short:   cfg.Short,
		Long:    cfg.Long,
		Version: cfg.Version,
	}

	c.rootCmd.SetOut(os.Stdout)
	c.rootCmd.SetErr(os.Stderr)

	if err := ctx.Container().RegisterInstance(c.rootCmd, reflect.TypeFor[*cobra.Command]()); err != nil {
		return fmt.Errorf("failed to register Cobra RootCmd: %w", err)
	}

	c.logger.Info(ctx.Context(), "Cobra CLI configured",
		log.KeyValue{Key: "name", Value: cfg.Use},
		log.KeyValue{Key: "version", Value: cfg.Version},
	)

	return nil
}

// GetRootCmd 获取根命令实例。
func (c *CobraAutoConfiguration) GetRootCmd() *cobra.Command {
	return c.rootCmd
}

// AddCommand 添加子命令。
func (c *CobraAutoConfiguration) AddCommand(cmd *cobra.Command) {
	c.rootCmd.AddCommand(cmd)
}

// Execute 执行 CLI。
func (c *CobraAutoConfiguration) Execute() error {
	return c.rootCmd.Execute()
}

// CobraConfig Cobra CLI 配置。
type CobraConfig struct {
	Enabled bool   `json:"enabled" mapstructure:"enabled"`
	Use     string `json:"use" mapstructure:"use"`
	Short   string `json:"short" mapstructure:"short"`
	Long    string `json:"long" mapstructure:"long"`
	Version string `json:"version" mapstructure:"version"`
}

// 配置常量。
const (
	CobraEnabled        = "cobra.enabled"
	DefaultCobraUse     = "app"
	DefaultCobraShort   = "A CLI application"
	DefaultCobraLong    = "A CLI application built with enhance framework"
	DefaultCobraVersion = "1.0.0"
	ConditionTrue       = "true"
)

// loadConfig 从 Environment 加载 Cobra 配置。
func (c *CobraAutoConfiguration) loadConfig(env *environment.Environment) (*CobraConfig, error) {
	cfg := &CobraConfig{
		Use:     DefaultCobraUse,
		Short:   DefaultCobraShort,
		Long:    DefaultCobraLong,
		Version: DefaultCobraVersion,
	}

	if err := env.BindPrefix("cobra", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind Cobra config: %w", err)
	}

	return cfg, nil
}
