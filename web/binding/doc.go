// Package binding 提供 HTTP 参数绑定功能。
//
// 该模块提供将 HTTP 请求数据（表单、查询参数、JSON 请求体）绑定到 Go 结构体的功能。
// 支持自定义结构体标签和字段验证。
//
// # 功能特性
//
//   - 表单数据绑定：将 POST 表单数据绑定到结构体
//   - 查询参数绑定：将 URL 查询参数绑定到结构体
//   - JSON 绑定：将 JSON 请求体解码到结构体
//   - 自定义标签：支持自定义结构体标签名（默认 "form"）
//   - 字段验证：支持 required 标记的必填字段验证
//
// # 使用方式
//
// 定义结构体：
//
//	type UserForm struct {
//	    Name  string `form:"name,required"`
//	    Email string `form:"email"`
//	    Age   int    `form:"age"`
//	}
//
// 绑定表单数据：
//
//	binder := binding.NewBinder()
//	var form UserForm
//	err := binder.Bind(req, &form)
//
// 绑定查询参数：
//
//	err := binder.BindQuery(req, &form)
//
// 绑定 JSON：
//
//	err := binder.BindJSON(req, &form)
//
// 使用自定义标签：
//
//	binder := binding.NewBinder(binding.WithTagName("json"))
package binding
