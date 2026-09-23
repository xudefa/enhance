// Package health 提供健康检查功能，用于 enhance 框架。
//
// 实现类已按"一实现一文件"原则拆分到独立文件中：
//   - simple_indicator.go: SimpleIndicator
//   - indicator_registry.go: IndicatorRegistry + 全局注册表
//   - health_status_builder.go: HealthBuilder
//   - runtime_health_indicator.go: RuntimeHealthIndicator
//   - system_health_indicator.go: SystemHealthIndicator
//   - health_check_service.go: HealthCheckService + 默认服务
package health
