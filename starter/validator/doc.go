// Package validator 提供数据验证自动配置。
//
// Validator 是 Go 语言最流行的数据验证库。
//
// 功能特性：
//   - 自动配置验证器
//   - 支持结构体标签验证
//   - 自定义验证器支持
//   - 内置常用验证规则
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
//	    Age   int    `validate:"required,min=1,max=130"`
//	}
//
//	v := core.MustGetBean[*validator.Validate](app.Container())
//	err := v.Struct(user)
package validator
