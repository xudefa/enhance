// Package viper 提供 Viper 配置管理增强自动配置。
//
// Viper 是 Go 语言的配置管理库，支持多种配置格式。
//
// 功能特性：
//   - 自动配置 Viper 实例
//   - 支持 YAML/JSON/TOML 等格式
//   - 配置文件热更新
//   - 环境变量覆盖
//
// 配置示例：
//
//	{
//	  "viper": {
//	    "enabled": true,
//	    "config-name": "application",
//	    "config-type": "yaml",
//	    "config-path": ".",
//	    "watch-changes": false
//	  }
//	}
//
// 使用示例：
//
//	v := core.MustGetBean[*viper.Viper](app.Container())
//	value := v.GetString("app.name")
package viper
