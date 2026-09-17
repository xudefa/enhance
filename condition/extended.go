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

// Matches 解析占位符并求值 SpEL 表达式，非布尔结果按存在性判断。
func (e *expressionCondition) Matches(ctx ConditionContext) bool {
	expr := e.resolvePlaceholders(e.expression, ctx)

	parser := spel.NewSpelParser()
	parsed, err := parser.ParseExpression(expr)
	if err != nil {
		return false
	}

	evalCtx := newConditionEvaluationContext(ctx)
	parsedValue, err := parsed.GetValue(evalCtx)
	if err != nil {
		return false
	}

	if boolResult, ok := parsedValue.(bool); ok {
		return boolResult
	}

	if parsedValue == nil {
		return false
	}

	return true
}

// String 返回表达式条件的可读描述。
func (e *expressionCondition) String() string {
	return fmt.Sprintf("OnExpression(%s)", e.expression)
}

// resolvePlaceholders 解析 ${...} 占位符
func (e *expressionCondition) resolvePlaceholders(expr string, ctx ConditionContext) string {
	const maxIterations = 10
	for range maxIterations {
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
		propValue, ok := ctx.GetProperty(placeholder)
		if !ok {
			return expr
		}

		// 将值转换为字符串并添加引号（如果是字符串）
		strVal := valAsString(propValue)
		// 如果是字符串类型，添加引号以便 SpEL 正确解析
		if _, isString := propValue.(string); isString {
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

// GetRootObject 返回表达式求值的根对象。
func (c *conditionEvaluationContext) GetRootObject() any {
	return nil
}

// SetRootObject 设置表达式求值的根对象。
func (c *conditionEvaluationContext) SetRootObject(root any) {
}

// GetVariable 按键读取求值变量，读取失败返回 false。
func (c *conditionEvaluationContext) GetVariable(name string) (any, bool) {
	return c.ctx.GetProperty(name)
}

// SetVariable 设置求值变量。
func (c *conditionEvaluationContext) SetVariable(name string, value any) {
}

// GetPropertyAccessor 返回属性访问器，用于解析表达式中的属性引用。
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

// Matches 检查指定资源是否存在（或不存在，取决于配置），相对路径基于当前工作目录解析。
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

// String 返回资源条件的可读描述。
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

// Matches 检查指定环境变量是否存在（或不存在，取决于配置）。
func (e *envVarCondition) Matches(ctx ConditionContext) bool {
	_, exists := os.LookupEnv(e.envVar)

	if e.missing {
		return !exists
	}
	return exists
}

// String 返回环境变量条件的可读描述。
func (e *envVarCondition) String() string {
	if e.missing {
		return fmt.Sprintf("OnEnvVarMissing(%s)", e.envVar)
	}
	return fmt.Sprintf("OnEnvVarExists(%s)", e.envVar)
}
