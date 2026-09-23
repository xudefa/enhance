// Package spel 提供 Spring Expression Language (SpEL) 表达式支持，用于 enhance 框架。
package spel

import (
	"fmt"
	"reflect"
)

// GetRootObject 返回表达式求值的根对象。
func (c *standardEvaluationContextImpl) GetRootObject() any {
	return c.rootObject
}

// SetRootObject 设置表达式求值的根对象。
func (c *standardEvaluationContextImpl) SetRootObject(root any) {
	c.rootObject = root
}

// GetVariable 按键读取求值变量并返回其是否存在。
func (c *standardEvaluationContextImpl) GetVariable(name string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.variables[name]
	return v, ok
}

// SetVariable 设置求值变量。
func (c *standardEvaluationContextImpl) SetVariable(name string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.variables[name] = value
}

// GetPropertyAccessor 返回属性访问器。
func (c *standardEvaluationContextImpl) GetPropertyAccessor() PropertyAccessor {
	return c.propertyAccessor
}

// GetProperty 通过反射或标签读取目标结构体的属性值。
func (a *reflectPropertyAccessorImpl) GetProperty(target any, name string) (any, error) {
	if target == nil {
		return nil, fmt.Errorf("cannot get property %s from nil", name)
	}

	rv := reflect.ValueOf(target)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("cannot get property from non-struct type")
	}

	field := rv.FieldByName(name)
	if !field.IsValid() {
		structType := rv.Type()
		for i := range structType.NumField() {
			fd := structType.Field(i)
			if fd.Tag.Get("json") == name || fd.Tag.Get("spel") == name {
				field = rv.Field(i)
				break
			}
		}
	}

	if !field.IsValid() {
		return nil, fmt.Errorf("property %s not found", name)
	}

	if !field.CanInterface() {
		return nil, fmt.Errorf("property %s is not exported", name)
	}

	return field.Interface(), nil
}

// SetProperty 通过反射设置目标结构体的属性值，自动处理类型转换。
func (a *reflectPropertyAccessorImpl) SetProperty(target any, name string, value any) error {
	if target == nil {
		return fmt.Errorf("cannot set property %s on nil", name)
	}

	rv := reflect.ValueOf(target)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("cannot set property on non-struct type")
	}

	field := rv.FieldByName(name)
	if !field.IsValid() {
		return fmt.Errorf("property %s not found", name)
	}

	if !field.CanSet() {
		return fmt.Errorf("property %s is not settable", name)
	}

	valueRV := reflect.ValueOf(value)
	if !valueRV.IsValid() {
		// nil 值只能赋给可空类型
		if isNilable(field.Type()) {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return fmt.Errorf("property %s: cannot set nil to %s", name, field.Type())
	}

	if !valueRV.Type().AssignableTo(field.Type()) {
		if !valueRV.Type().ConvertibleTo(field.Type()) {
			return fmt.Errorf("property %s: cannot set value of type %s to %s", name, valueRV.Type(), field.Type())
		}
		valueRV = valueRV.Convert(field.Type())
	}

	field.Set(valueRV)
	return nil
}

// isNilable 判断类型是否为可空类型（指针、切片、map、通道、函数、接口）。
func isNilable(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
		return true
	}
	return false
}
