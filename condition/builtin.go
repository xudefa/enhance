// Package condition 提供条件化注册功能，用于 enhance 框架。
//
// 条件实现已按"一实现一文件"原则拆分到独立文件中：
//   - property_condition.go: propertyCondition
//   - property_or_default_condition.go: propertyOrDefaultCondition
//   - missing_property_condition.go: missingPropertyCondition
//   - bean_condition.go: beanCondition
//   - profile_condition.go: profileCondition
//   - module_condition.go: moduleCondition
//   - property_prefix_condition.go: propertyPrefixCondition
//   - custom_condition.go: customCondition
package condition
