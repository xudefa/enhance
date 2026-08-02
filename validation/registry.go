// Package validation 提供参数校验功能，用于 enhance 框架。
package validation

import (
	"reflect"
)

// NewValidatorRegistry 创建新的验证器注册表。
func NewValidatorRegistry() *ValidatorRegistry {
	return &ValidatorRegistry{}
}

// Register 注册结构体验证器。
func (r *ValidatorRegistry) Register(name string, validator CustomValidator) {
	r.validators.Store(name, validator)
}

// RegisterFunc 注册函数式验证器。
func (r *ValidatorRegistry) RegisterFunc(name string, validator func(reflect.Value, string) (bool, string)) {
	r.funcValidators.Store(name, validator)
}

// Get 获取结构体验证器。
func (r *ValidatorRegistry) Get(name string) (CustomValidator, bool) {
	v, ok := r.validators.Load(name)
	if !ok {
		return nil, false
	}
	return v.(CustomValidator), true
}

// GetFunc 获取函数式验证器。
func (r *ValidatorRegistry) GetFunc(name string) (func(reflect.Value, string) (bool, string), bool) {
	v, ok := r.funcValidators.Load(name)
	if !ok {
		return nil, false
	}
	return v.(func(reflect.Value, string) (bool, string)), true
}

// Unregister 注销验证器。
func (r *ValidatorRegistry) Unregister(name string) {
	r.validators.Delete(name)
	r.funcValidators.Delete(name)
}
