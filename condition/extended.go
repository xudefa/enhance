package condition

import (
	"fmt"
	"os"
	"strings"

	"github.com/xudefa/enhance/spel"
)

// OnExpression 创建基于 SpEL 表达式的条件
//
// 表达式支持以下语法:
//   - 属性引用: ${server.port} 或 server.port
//   - 比较运算: >, <, >=, <=, ==, !=
//   - 逻辑运算: &&, ||, !
//   - 三元运算: condition ? trueVal : falseVal
//   - 字面量: 字符串、数字、布尔值
//
// 示例:
//
//	condition.OnExpression("${server.port} > 8080")
//	condition.OnExpression("${app.env} == 'prod' && ${debug.enabled} == false")
func OnExpression(expression string) Condition {
	return &expressionCondition{
		expression: expression,
	}
}

// expressionCondition 基于 SpEL 表达式的条件实现
type expressionCondition struct {
	expression string // SpEL 表达式
}

func (e *expressionCondition) Matches(ctx ConditionContext) bool {
	expr := e.resolvePlaceholders(e.expression, ctx)

	parser := spel.NewSpelParser()
	parsed, err := parser.ParseExpression(expr)
	if err != nil {
		return false
	}

	evalCtx := newConditionEvaluationContext(ctx)
	result, err := parsed.GetValue(evalCtx)
	if err != nil {
		return false
	}

	if boolResult, ok := result.(bool); ok {
		return boolResult
	}

	if result == nil {
		return false
	}

	return true
}

func (e *expressionCondition) String() string {
	return fmt.Sprintf("OnExpression(%s)", e.expression)
}

// resolvePlaceholders 解析 ${...} 占位符
func (e *expressionCondition) resolvePlaceholders(expr string, ctx ConditionContext) string {
	for {
		start := strings.Index(expr, "${")
		if start == -1 {
			break
		}
		end := strings.Index(expr[start:], "}")
		if end == -1 {
			break
		}
		end += start

		placeholder := expr[start+2 : end]
		val, ok := ctx.GetProperty(placeholder)
		if !ok {
			return expr
		}

		// 将值转换为字符串并添加引号（如果是字符串）
		strVal := valAsString(val)
		// 如果是字符串类型，添加引号以便 SpEL 正确解析
		if _, isString := val.(string); isString {
			strVal = "'" + strVal + "'"
		}
		expr = expr[:start] + strVal + expr[end+1:]
	}
	return expr
}

// conditionEvaluationContext 条件表达式求值上下文
type conditionEvaluationContext struct {
	ctx ConditionContext
}

func newConditionEvaluationContext(ctx ConditionContext) *conditionEvaluationContext {
	return &conditionEvaluationContext{ctx: ctx}
}

func (c *conditionEvaluationContext) GetRootObject() any {
	return nil
}

func (c *conditionEvaluationContext) SetRootObject(root any) {
}

func (c *conditionEvaluationContext) GetVariable(name string) (any, bool) {
	return c.ctx.GetProperty(name)
}

func (c *conditionEvaluationContext) SetVariable(name string, value any) {
}

func (c *conditionEvaluationContext) GetPropertyAccessor() spel.PropertyAccessor {
	return spel.NewReflectPropertyAccessor()
}

// OnResourceExists 创建基于资源存在性的条件
//
// 支持以下资源前缀:
//   - classpath: - 类路径资源（Go 中检查文件是否存在）
//   - file: - 文件系统资源
//   - 无前缀 - 默认为 classpath
//
// 示例:
//
//	condition.OnResourceExists("classpath:config.yml")
//	condition.OnResourceExists("file:/etc/app/config.json")
//	condition.OnResourceExists("config.yml") // 等同于 classpath:config.yml
func OnResourceExists(location string) Condition {
	return &resourceCondition{
		location: location,
		missing:  false,
	}
}

// OnResourceMissing 创建基于资源不存在性的条件
//
// 当指定资源不存在时匹配。
func OnResourceMissing(location string) Condition {
	return &resourceCondition{
		location: location,
		missing:  true,
	}
}

// resourceCondition 基于资源存在性的条件实现
type resourceCondition struct {
	location string // 资源位置
	missing  bool   // true 表示检查不存在
}

func (r *resourceCondition) Matches(ctx ConditionContext) bool {
	path := r.resolvePath(r.location)

	// 如果是相对路径，尝试使用当前工作目录
	if !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, ".") {
		cwd, err := os.Getwd()
		if err == nil {
			path = cwd + "/" + path
		}
	}

	_, err := os.Stat(path)
	exists := err == nil

	if r.missing {
		return !exists
	}
	return exists
}

func (r *resourceCondition) String() string {
	if r.missing {
		return fmt.Sprintf("OnResourceMissing(%s)", r.location)
	}
	return fmt.Sprintf("OnResourceExists(%s)", r.location)
}

// resolvePath 解析资源路径
func (r *resourceCondition) resolvePath(location string) string {
	if strings.HasPrefix(location, "file:") {
		return location[5:]
	}

	if strings.HasPrefix(location, "classpath:") {
		return location[10:]
	}

	return location
}

// OnEnvVarExists 创建基于环境变量存在性的条件
//
// 当指定的环境变量存在时匹配。
//
// 示例:
//
//	condition.OnEnvVarExists("DATABASE_URL")
//	condition.OnEnvVarExists("REDIS_HOST")
func OnEnvVarExists(envVar string) Condition {
	return &envVarCondition{
		envVar:  envVar,
		missing: false,
	}
}

// OnEnvVarMissing 创建基于环境变量不存在性的条件
//
// 当指定的环境变量不存在时匹配。
func OnEnvVarMissing(envVar string) Condition {
	return &envVarCondition{
		envVar:  envVar,
		missing: true,
	}
}

// envVarCondition 基于环境变量存在性的条件实现
type envVarCondition struct {
	envVar  string // 环境变量名
	missing bool   // true 表示检查不存在
}

func (e *envVarCondition) Matches(ctx ConditionContext) bool {
	_, exists := os.LookupEnv(e.envVar)

	if e.missing {
		return !exists
	}
	return exists
}

func (e *envVarCondition) String() string {
	if e.missing {
		return fmt.Sprintf("OnEnvVarMissing(%s)", e.envVar)
	}
	return fmt.Sprintf("OnEnvVarExists(%s)", e.envVar)
}

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
	var orConditions []Condition
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
