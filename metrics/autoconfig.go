package metrics

import (
	"reflect"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
)

// MetricsAutoConfiguration Metrics 自动配置
type MetricsAutoConfiguration struct{}

// Configure 注册 SimpleRegistry 为单例 Bean
func (m *MetricsAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	reg := NewSimpleRegistry()
	if err := ctx.Container().RegisterInstance(reg, reflect.TypeOf(reg)); err != nil {
		return err
	}
	return nil
}

func init() {
	boot.RegisterAutoConfigWith(
		&MetricsAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(MetricsEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityMonitoringLayer)), // 监控层，与 Actuator 同级
	)
}
