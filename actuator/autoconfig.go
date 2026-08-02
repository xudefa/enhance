package actuator

import (
	"reflect"

	"github.com/xudefa/enhance/actuator/admin"
	"github.com/xudefa/enhance/actuator/health"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
)

// ActuatorAutoConfiguration Actuator 自动配置
type ActuatorAutoConfiguration struct{}

// Configure 创建 Actuator 实例并注册为 Bean
func (a *ActuatorAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	act := New(ctx)

	agg := health.NewAggregator()

	if indicators, err := ctx.Container().Get(reflect.TypeOf((*health.Indicator)(nil)).Elem()); err == nil && len(indicators) > 0 {
		for _, ind := range indicators {
			if h, ok := ind.(health.Indicator); ok && h != nil {
				agg.AddIndicator(h)
			}
		}
	}
	act.SetHealthAggregator(agg)

	// 创建 Admin 服务器（可视化监控）
	registry := admin.NewApplicationRegistry()
	adminServer := admin.NewAdminServer(registry)
	_ = adminServer // 用于后续路由注册

	// 注册应用实例到 Admin
	if appNameVal, ok := ctx.Environment().GetProperty(AppName); ok {
		appName, ok := appNameVal.(string)
		if !ok {
			appName = DefaultAppName
		}
		appVersion := DefaultAppVersion
		if v, ok := ctx.Environment().GetProperty(AppVersion); ok && v != "" {
			if versionStr, ok := v.(string); ok {
				appVersion = versionStr
			}
		}
		instance := admin.NewApplicationInstance(appName, appVersion)
		registry.Register(instance)
	}
	return ctx.Container().RegisterInstance(act, reflect.TypeOf(act))
}

func init() {
	boot.RegisterAutoConfigWith(
		&ActuatorAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(ActuatorEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityMonitoringLayer)), // 监控层，最后执行，监控所有组件
	)
}
