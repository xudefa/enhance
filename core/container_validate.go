package core

import (
	"fmt"
	"reflect"
	"strings"
)

// Validate 验证所有已注册Bean的依赖是否可解析，检测循环依赖。
func (c *defaultContainer) Validate() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.initialized {
		return fmt.Errorf("container already initialized, cannot validate")
	}

	// 1. 检查所有Bean的依赖是否已注册
	if err := c.validateDependencies(); err != nil {
		return err
	}

	// 2. 检测循环依赖
	if err := c.detectCircularDependencies(); err != nil {
		return err
	}

	return nil
}

// validateDependencies 检查所有Bean的依赖是否已注册
func (c *defaultContainer) validateDependencies() error {
	types := c.Types()
	typeSet := make(map[reflect.Type]bool)
	for _, typ := range types {
		typeSet[typ] = true
	}

	for _, typ := range types {
		if err := c.validateTypeDependencies(typ, typeSet); err != nil {
			return err
		}
	}

	return nil
}

// validateTypeDependencies 检查指定类型的依赖
func (c *defaultContainer) validateTypeDependencies(typ reflect.Type, typeSet map[reflect.Type]bool) error {
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

		if c.parent == nil {
			return fmt.Errorf("dependency not found: field '%s' of type '%s' requires '%s'",
				field.Name, typ.String(), fieldType.String())
		}

		ext, ok := c.parent.(ContainerExt)
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

// detectCircularDependencies 检测循环依赖
func (c *defaultContainer) detectCircularDependencies() error {
	types := c.Types()
	visited := make(map[reflect.Type]bool)
	recStack := make(map[reflect.Type]bool)

	for _, typ := range types {
		if !visited[typ] {
			if err := c.detectCircularDFS(typ, visited, recStack, []string{}); err != nil {
				return err
			}
		}
	}

	return nil
}

// detectCircularDFS 深度优先搜索检测循环依赖
func (c *defaultContainer) detectCircularDFS(typ reflect.Type, visited, recStack map[reflect.Type]bool, path []string) error {
	visited[typ] = true
	recStack[typ] = true
	path = append(path, typ.String())

	actualType := typ
	if typ.Kind() == reflect.Ptr {
		actualType = typ.Elem()
	}

	if actualType.Kind() == reflect.Struct {
		for i := 0; i < actualType.NumField(); i++ {
			field := actualType.Field(i)
			if !field.IsExported() {
				continue
			}

			if _, ok := field.Tag.Lookup("inject"); ok {
				fieldType := field.Type

				if recStack[fieldType] {
					path = append(path, fieldType.String())
					return fmt.Errorf("circular dependency detected: %s", strings.Join(path, " -> "))
				}

				if !visited[fieldType] {
					if err := c.detectCircularDFS(fieldType, visited, recStack, path); err != nil {
						return err
					}
				}
			}
		}
	}

	recStack[typ] = false
	return nil
}
