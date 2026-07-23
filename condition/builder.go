package condition

// ConditionBuilder 条件构建器，提供流式 DSL
//
// 示例:
//
//	condition.New().
//	    OnProperty("feature.enabled", "true").
//	    And().
//	    OnBean("dataSource").
//	    Or().
//	    OnProfile("dev").
//	    Build()
type ConditionBuilder struct {
	conditions []Condition
	operators  []string // "and", "or"
}

// New 创建条件构建器
func New() *ConditionBuilder {
	return &ConditionBuilder{
		conditions: make([]Condition, 0),
		operators:  make([]string, 0),
	}
}

// OnProperty 添加属性条件
func (b *ConditionBuilder) OnProperty(key string, expectedValue ...string) *ConditionBuilder {
	b.conditions = append(b.conditions, OnProperty(key, expectedValue...))
	return b
}

// OnMissingProperty 添加属性缺失条件
func (b *ConditionBuilder) OnMissingProperty(key string) *ConditionBuilder {
	b.conditions = append(b.conditions, OnMissingProperty(key))
	return b
}

// OnBean 添加 Bean 存在条件
func (b *ConditionBuilder) OnBean(beanID string) *ConditionBuilder {
	b.conditions = append(b.conditions, OnBean(beanID))
	return b
}

// OnMissingBean 添加 Bean 缺失条件
func (b *ConditionBuilder) OnMissingBean(beanID string) *ConditionBuilder {
	b.conditions = append(b.conditions, OnMissingBean(beanID))
	return b
}

// OnProfile 添加 Profile 条件
func (b *ConditionBuilder) OnProfile(profile string) *ConditionBuilder {
	b.conditions = append(b.conditions, OnProfile(profile))
	return b
}

// OnModuleLoaded 添加模块加载条件
func (b *ConditionBuilder) OnModuleLoaded(moduleName string) *ConditionBuilder {
	b.conditions = append(b.conditions, OnModuleLoaded(moduleName))
	return b
}

// OnMissingModule 添加模块缺失条件
func (b *ConditionBuilder) OnMissingModule(moduleName string) *ConditionBuilder {
	b.conditions = append(b.conditions, OnMissingModule(moduleName))
	return b
}

// OnExpression 添加表达式条件
func (b *ConditionBuilder) OnExpression(expression string) *ConditionBuilder {
	b.conditions = append(b.conditions, OnExpression(expression))
	return b
}

// OnResourceExists 添加资源存在条件
func (b *ConditionBuilder) OnResourceExists(location string) *ConditionBuilder {
	b.conditions = append(b.conditions, OnResourceExists(location))
	return b
}

// OnResourceMissing 添加资源缺失条件
func (b *ConditionBuilder) OnResourceMissing(location string) *ConditionBuilder {
	b.conditions = append(b.conditions, OnResourceMissing(location))
	return b
}

// OnEnvVarExists 添加环境变量存在条件
func (b *ConditionBuilder) OnEnvVarExists(envVar string) *ConditionBuilder {
	b.conditions = append(b.conditions, OnEnvVarExists(envVar))
	return b
}

// OnEnvVarMissing 添加环境变量缺失条件
func (b *ConditionBuilder) OnEnvVarMissing(envVar string) *ConditionBuilder {
	b.conditions = append(b.conditions, OnEnvVarMissing(envVar))
	return b
}

// And 添加逻辑与操作
func (b *ConditionBuilder) And() *ConditionBuilder {
	b.operators = append(b.operators, "and")
	return b
}

// Or 添加逻辑或操作
func (b *ConditionBuilder) Or() *ConditionBuilder {
	b.operators = append(b.operators, "or")
	return b
}

// Not 添加逻辑非操作（对下一个条件取反）
func (b *ConditionBuilder) Not() *ConditionBuilder {
	b.operators = append(b.operators, "not")
	return b
}

// All 添加一个 All 复合条件
func (b *ConditionBuilder) All(conditions ...Condition) *ConditionBuilder {
	b.conditions = append(b.conditions, All(conditions...))
	return b
}

// Any 添加一个 Any 复合条件
func (b *ConditionBuilder) Any(conditions ...Condition) *ConditionBuilder {
	b.conditions = append(b.conditions, Any(conditions...))
	return b
}

// Build 构建最终条件
func (b *ConditionBuilder) Build() Condition {
	if len(b.conditions) == 0 {
		return &alwaysTrueCondition{}
	}

	// 如果有操作符，即使只有一个条件也需要处理
	if len(b.operators) > 0 {
		return b.buildComposite()
	}

	if len(b.conditions) == 1 {
		return b.conditions[0]
	}

	return All(b.conditions...)
}

// buildComposite 构建复合条件
func (b *ConditionBuilder) buildComposite() Condition {
	if len(b.conditions) == 0 {
		return &alwaysTrueCondition{}
	}

	if len(b.operators) == 0 {
		return All(b.conditions...)
	}

	// 处理 Not 操作符
	processedConditions := make([]Condition, 0, len(b.conditions))
	processedOperators := make([]string, 0, len(b.operators))
	notPending := false
	condIndex := 0

	for i := 0; i < len(b.operators); i++ {
		op := b.operators[i]

		if op == "not" {
			notPending = true
			continue
		}

		if condIndex < len(b.conditions) {
			cond := b.conditions[condIndex]
			if notPending {
				processedConditions = append(processedConditions, Not(cond))
				notPending = false
				processedOperators = append(processedOperators, op)
				condIndex++
				continue
			}
			processedConditions = append(processedConditions, cond)
			processedOperators = append(processedOperators, op)
			condIndex++
		}
	}

	// 处理最后一个条件
	if condIndex < len(b.conditions) {
		cond := b.conditions[condIndex]
		if notPending {
			processedConditions = append(processedConditions, Not(cond))
			return b.buildWithOperators(processedConditions, processedOperators)
		}
		processedConditions = append(processedConditions, cond)
	}

	// 按操作符分组
	return b.buildWithOperators(processedConditions, processedOperators)
}

// buildWithOperators 根据操作符构建条件
func (b *ConditionBuilder) buildWithOperators(conditions []Condition, operators []string) Condition {
	if len(conditions) == 0 {
		return &alwaysTrueCondition{}
	}

	// 如果没有操作符，默认使用 All
	if len(operators) == 0 {
		return All(conditions...)
	}

	// 检查是否全部是 AND 或全部是 OR
	allAnd := true
	allOr := true
	for _, op := range operators {
		if op != "and" {
			allAnd = false
		}
		if op != "or" {
			allOr = false
		}
	}

	if allAnd {
		return All(conditions...)
	}

	if allOr {
		return Any(conditions...)
	}

	// 混合操作符，需要分组处理
	return b.buildMixed(conditions, operators)
}

// buildMixed 构建混合操作符条件
func (b *ConditionBuilder) buildMixed(conditions []Condition, operators []string) Condition {
	if len(conditions) == 0 {
		return &alwaysTrueCondition{}
	}

	// 按 OR 分割成多个 AND 组
	var groups [][]Condition
	currentGroup := []Condition{conditions[0]}

	for i, op := range operators {
		if op == "or" {
			groups = append(groups, currentGroup)
			if i+1 < len(conditions) {
				currentGroup = []Condition{conditions[i+1]}
			} else {
				currentGroup = []Condition{}
			}
			continue
		}
		if i+1 < len(conditions) {
			currentGroup = append(currentGroup, conditions[i+1])
		}
	}

	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}

	// 每组构建为 All，然后用 Any 连接
	orConditions := make([]Condition, 0, len(groups))
	for _, group := range groups {
		if len(group) == 1 {
			orConditions = append(orConditions, group[0])
			continue
		}
		orConditions = append(orConditions, All(group...))
	}

	if len(orConditions) == 1 {
		return orConditions[0]
	}

	return Any(orConditions...)
}

// alwaysTrueCondition 永远为 true 的条件
type alwaysTrueCondition struct{}

func (a *alwaysTrueCondition) Matches(ctx ConditionContext) bool {
	return true
}

func (a *alwaysTrueCondition) String() string {
	return "AlwaysTrue"
}

// All 改进版：支持链式调用
func AllWith(conditions ...Condition) ConditionBuilderGroup {
	return ConditionBuilderGroup{
		groups:     conditions,
		groupTypes: []string{},
	}
}

// ConditionBuilderGroup 条件组构建器
type ConditionBuilderGroup struct {
	groups     []Condition // 条件组列表
	groupTypes []string    // 组类型："and" 或 "or"
}

// Or 添加或条件组
func (g ConditionBuilderGroup) Or(conditions ...Condition) ConditionBuilderGroup {
	if len(g.groups) == 0 {
		// 第一个组：直接添加条件
		return ConditionBuilderGroup{
			groups:     conditions,
			groupTypes: []string{},
		}
	}

	// 将新条件作为 OR 分支添加
	allGroups := append(g.groups, conditions...)
	return ConditionBuilderGroup{
		groups:     allGroups,
		groupTypes: append(g.groupTypes, "or"),
	}
}

// And 添加与条件组
func (g ConditionBuilderGroup) And(conditions ...Condition) ConditionBuilderGroup {
	if len(g.groups) == 0 {
		// 第一个组：直接添加条件
		return ConditionBuilderGroup{
			groups:     conditions,
			groupTypes: []string{},
		}
	}

	// 将新条件作为 AND 分支添加
	allGroups := append(g.groups, conditions...)
	return ConditionBuilderGroup{
		groups:     allGroups,
		groupTypes: append(g.groupTypes, "and"),
	}
}

// Build 构建最终条件
func (g ConditionBuilderGroup) Build() Condition {
	if len(g.groups) == 0 {
		return &alwaysTrueCondition{}
	}

	if len(g.groups) == 1 {
		return g.groups[0]
	}

	// 检查组类型
	allOr := true
	allAnd := true
	for _, t := range g.groupTypes {
		if t != "or" {
			allOr = false
		}
		if t != "and" {
			allAnd = false
		}
	}

	if allAnd {
		return All(g.groups...)
	}

	if allOr {
		return Any(g.groups...)
	}

	// 混合类型，默认使用 Any 连接各组
	return Any(g.groups...)
}
