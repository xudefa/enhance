package condition

import "fmt"

// moduleCondition 基于模块加载的条件实现
type moduleCondition struct {
	moduleName string // 模块名称
	missing    bool   // true 表示检查未加载
}

// OnModuleLoaded 创建基于模块加载的条件
//
// 当指定模块已加载到容器中时匹配。
// Go 中不支持 Java 的 Class.forName() 动态类加载,改用模块名称检查。
func OnModuleLoaded(moduleName string) Condition {
	return &moduleCondition{moduleName: moduleName, missing: false}
}

// OnMissingModule 创建基于模块未加载的条件
//
// 当指定模块未加载到容器中时匹配。
func OnMissingModule(moduleName string) Condition {
	return &moduleCondition{moduleName: moduleName, missing: true}
}

// Matches 实现 Condition 接口
func (c *moduleCondition) Matches(ctx ConditionContext) bool {
	container := ctx.Container()
	has := container.Has(c.moduleName)
	if c.missing {
		return !has
	}
	return has
}

// String 返回条件的字符串表示
func (c *moduleCondition) String() string {
	if c.missing {
		return fmt.Sprintf("OnMissingModule(%s)", c.moduleName)
	}
	return fmt.Sprintf("OnModuleLoaded(%s)", c.moduleName)
}
