// Package spel 提供 Spring Expression Language (SpEL) 表达式支持，用于 enhance 框架。
package spel

import (
	"fmt"
	"reflect"
)

func (c *standardEvaluationContextImpl) GetRootObject() any {
	return c.rootObject
}

func (c *standardEvaluationContextImpl) SetRootObject(root any) {
	c.rootObject = root
}

func (c *standardEvaluationContextImpl) GetVariable(name string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.variables[name]
	return v, ok
}

func (c *standardEvaluationContextImpl) SetVariable(name string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.variables[name] = value
}

func (c *standardEvaluationContextImpl) GetPropertyAccessor() PropertyAccessor {
	return c.propertyAccessor
}

func (a *reflectPropertyAccessorImpl) GetProperty(target any, name string) (any, error) {
	if target == nil {
		return nil, fmt.Errorf("cannot get property %s from nil", name)
	}

	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("cannot get property from non-struct type")
	}

	field := v.FieldByName(name)
	if !field.IsValid() {
		t := v.Type()
		for i := range t.NumField() {
			fd := t.Field(i)
			if fd.Tag.Get("json") == name || fd.Tag.Get("spel") == name {
				field = v.Field(i)
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

func (a *reflectPropertyAccessorImpl) SetProperty(target any, name string, value any) error {
	if target == nil {
		return fmt.Errorf("cannot set property %s on nil", name)
	}

	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("cannot set property on non-struct type")
	}

	field := v.FieldByName(name)
	if !field.IsValid() {
		return fmt.Errorf("property %s not found", name)
	}

	if !field.CanSet() {
		return fmt.Errorf("property %s is not settable", name)
	}

	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		// nil 值只能赋给可空类型
		if isNilable(field.Type()) {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return fmt.Errorf("property %s: cannot set nil to %s", name, field.Type())
	}

	if !rv.Type().AssignableTo(field.Type()) {
		if !rv.Type().ConvertibleTo(field.Type()) {
			return fmt.Errorf("property %s: cannot set value of type %s to %s", name, rv.Type(), field.Type())
		}
		rv = rv.Convert(field.Type())
	}

	field.Set(rv)
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
