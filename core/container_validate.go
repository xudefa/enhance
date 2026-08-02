package core

import (
	"fmt"
	"reflect"
	"strings"
)

// Validate 验证所有已注册Bean的依赖是否可解析，检测循环依赖。
func (c *defaultContainer) Validate() error {
	// 短暂持锁读取状态快照，校验过程在锁外执行，
	// 避免在持有 RLock 时再次获取锁（validateDependencies/Types）导致死锁
	c.mu.RLock()
	if c.initialized {
		c.mu.RUnlock()
		return fmt.Errorf("container already initialized, cannot validate")
	}
	parent := c.parent
	beans := c.reg.ListBeans()
	c.mu.RUnlock()

	types := make([]reflect.Type, 0, len(beans))
	for _, def := range beans {
		types = append(types, def.Type)
	}

	// 1. 检查所有Bean的依赖是否已注册
	if err := c.validateDependencies(types, parent); err != nil {
		return err
	}

	// 2. 检测循环依赖
	if err := c.detectCircularDependencies(types); err != nil {
		return err
	}

	return nil
}

// validateDependencies 检查所有Bean的依赖是否已注册
func (c *defaultContainer) validateDependencies(types []reflect.Type, parent Container) error {
	typeSet := make(map[reflect.Type]bool)
	for _, typ := range types {
		typeSet[toPtrType(typ)] = true
		typeSet[typ] = true
	}

	for _, typ := range types {
		if err := c.validateTypeDependencies(typ, typeSet, parent); err != nil {
			return err
		}
	}

	return nil
}

// validateTypeDependencies 检查指定类型的依赖
func (c *defaultContainer) validateTypeDependencies(typ reflect.Type, typeSet map[reflect.Type]bool, parent Container) error {
	actualType := typ
	if typ.Kind() == reflect.Ptr {
		actualType = typ.Elem()
	}

	if actualType.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < actualType.NumField(); i++ {
		field := actualType.Field(i)
		if !field.IsExported() {
			continue
		}

		// 检查是否有 inject 标签（包括空标签）
		if _, ok := field.Tag.Lookup("inject"); !ok {
			continue
		}

		fieldType := field.Type
		if typeSet[fieldType] {
			continue
		}

		if parent == nil {
			return fmt.Errorf("dependency not found: field '%s' of type '%s' requires '%s'",
				field.Name, typ.String(), fieldType.String())
		}

		ext, ok := parent.(ContainerExt)
		if !ok {
			return fmt.Errorf("dependency not found: field '%s' of type '%s' requires '%s'",
				field.Name, typ.String(), fieldType.String())
		}

		parentTypes := ext.Types()
		found := false
		for _, pt := range parentTypes {
			if pt == fieldType {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("dependency not found: field '%s' of type '%s' requires '%s'",
				field.Name, typ.String(), fieldType.String())
		}
	}

	return nil
}

// detectCircularDependencies 检测循环依赖。
//
// 注意：此检测仅覆盖 struct 字段注入（inject tag）产生的依赖，
// 构造器注入（Factory 函数闭包捕获）的循环依赖无法在此静态检测。
func (c *defaultContainer) detectCircularDependencies(types []reflect.Type) error {
	visited := make(map[reflect.Type]bool)
	recStack := make(map[reflect.Type]bool)

	for _, typ := range types {
		key := toPtrType(typ)
		if !visited[key] {
			if err := c.detectCircularDFS(key, visited, recStack, []string{}); err != nil {
				return err
			}
		}
	}

	return nil
}

// toPtrType 将类型统一转换为指针类型，避免指针和非指针类型被当作不同 key。
func toPtrType(typ reflect.Type) reflect.Type {
	if typ.Kind() != reflect.Ptr {
		return reflect.PointerTo(typ)
	}
	return typ
}

// detectCircularDFS 深度优先搜索检测循环依赖
func (c *defaultContainer) detectCircularDFS(key reflect.Type, visited, recStack map[reflect.Type]bool, path []string) error {
	visited[key] = true
	recStack[key] = true
	path = append(path, key.String())

	actualType := key
	if key.Kind() == reflect.Ptr {
		actualType = key.Elem()
	}

	if actualType.Kind() == reflect.Struct {
		for i := 0; i < actualType.NumField(); i++ {
			field := actualType.Field(i)
			if !field.IsExported() {
				continue
			}

			if _, ok := field.Tag.Lookup("inject"); ok {
				fieldKey := toPtrType(field.Type)

				if recStack[fieldKey] {
					path = append(path, fieldKey.String())
					return fmt.Errorf("circular dependency detected: %s", strings.Join(path, " -> "))
				}

				if !visited[fieldKey] {
					if err := c.detectCircularDFS(fieldKey, visited, recStack, path); err != nil {
						return err
					}
				}
			}
		}
	}

	recStack[key] = false
	return nil
}
