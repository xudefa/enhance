package spel

import (
	"fmt"
	"strings"
)

// ParseExpression 解析并分类表达式，返回适合的表达式实现。
func (p *spelParserImpl) ParseExpression(expression string) (Expression, error) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil, fmt.Errorf("empty expression")
	}

	// 保留关键字
	if isLiteral(expression) {
		return &complexExpressionImpl{raw: expression}, nil
	}

	// 简单属性名：只包含字母、数字、下划线，且不包含运算符
	if isSimpleProperty(expression) {
		return &propertyExpressionImpl{property: expression}, nil
	}

	return &complexExpressionImpl{raw: expression}, nil
}

// isLiteral 检查是否为字面量（true/false/null/数字）。
func isLiteral(expr string) bool {
	switch strings.ToLower(expr) {
	case "true", "false", "null":
		return true
	}
	// 检查是否为数字
	if len(expr) == 0 {
		return false
	}
	for _, c := range expr {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// isSimpleProperty 检查是否为简单属性名。
func isSimpleProperty(expr string) bool {
	if len(expr) == 0 {
		return false
	}
	// 纯数字不是属性名
	allDigits := true
	for _, c := range expr {
		if c >= '0' && c <= '9' {
			continue
		}
		allDigits = false
		isLower := c >= 'a' && c <= 'z'
		isUpper := c >= 'A' && c <= 'Z'
		isUnderscore := c == '_'
		if !isLower && !isUpper && !isUnderscore {
			return false
		}
	}
	return !allDigits
}

// getRootObject 获取根对象，nil 时返回错误，复用公共逻辑
func getRootObject(ctx EvaluationContext) (any, error) {
	root := ctx.GetRootObject()
	if root == nil {
		return nil, fmt.Errorf("root object is nil")
	}
	return root, nil
}

// GetValue 求值属性表达式，优先读取变量，其次读取根对象属性。
func (e *propertyExpressionImpl) GetValue(ctx EvaluationContext) (any, error) {
	if variableValue, ok := ctx.GetVariable(e.property); ok {
		return variableValue, nil
	}

	root, err := getRootObject(ctx)
	if err != nil {
		return nil, err
	}

	return ctx.GetPropertyAccessor().GetProperty(root, e.property)
}

// SetValue 将值写入根对象的指定属性。
func (e *propertyExpressionImpl) SetValue(ctx EvaluationContext, value any) error {
	root, err := getRootObject(ctx)
	if err != nil {
		return err
	}

	return ctx.GetPropertyAccessor().SetProperty(root, e.property, value)
}

// String 返回属性表达式的属性名。
func (e *propertyExpressionImpl) String() string {
	return e.property
}

// GetValue 求值复杂表达式，按三元、逻辑、比较、算术等语法分派。
func (e *complexExpressionImpl) GetValue(ctx EvaluationContext) (any, error) {
	expr := e.raw

	if idx := strings.Index(expr, "?"); idx > 0 {
		return e.evaluateTernary(expr, ctx)
	}

	if strings.Contains(expr, "&&") || strings.Contains(expr, "||") {
		return e.evaluateLogical(expr, ctx)
	}

	for _, op := range []string{"==", "!=", ">=", "<=", ">", "<"} {
		if idx := strings.Index(expr, op); idx > 0 {
			return e.evaluateComparison(expr, op, ctx)
		}
	}

	for _, op := range []string{"+", "-", "*", "/"} {
		if idx := strings.LastIndex(expr, op); idx > 0 {
			return e.evaluateArithmetic(expr, op, idx, ctx)
		}
	}

	if strings.Contains(expr, "(") {
		return e.evaluateMethodCall(expr, ctx)
	}

	if strings.Contains(expr, ".") {
		return e.evaluatePropertyChain(expr, ctx)
	}

	return e.evaluateLiteral(expr)
}

// SetValue 复杂表达式不支持赋值，直接返回错误。
func (e *complexExpressionImpl) SetValue(ctx EvaluationContext, value any) error {
	return fmt.Errorf("cannot set value on complex expression")
}

// String 返回复杂表达式的原文。
func (e *complexExpressionImpl) String() string {
	return e.raw
}

func (e *complexExpressionImpl) evaluateTernary(expr string, ctx EvaluationContext) (any, error) {
	qIdx := strings.Index(expr, "?")
	cIdx := strings.LastIndex(expr, ":")

	if qIdx < 0 || cIdx < 0 || cIdx <= qIdx {
		return nil, fmt.Errorf("invalid ternary expression")
	}

	condition := strings.TrimSpace(expr[:qIdx])
	trueExpr := strings.TrimSpace(expr[qIdx+1 : cIdx])
	falseExpr := strings.TrimSpace(expr[cIdx+1:])

	condVal, err := e.evaluate(condition, ctx)
	if err != nil {
		return nil, fmt.Errorf("求值三元条件 %s 失败: %w", condition, err)
	}

	if isTruthy(condVal) {
		return e.evaluate(trueExpr, ctx)
	}
	return e.evaluate(falseExpr, ctx)
}

func (e *complexExpressionImpl) evaluateLogical(expr string, ctx EvaluationContext) (any, error) {
	var op, left, right string

	if idx := strings.Index(expr, "&&"); idx > 0 {
		op = "&&"
		left = strings.TrimSpace(expr[:idx])
		right = strings.TrimSpace(expr[idx+2:])
	}
	if idx := strings.Index(expr, "||"); idx > 0 {
		op = "||"
		left = strings.TrimSpace(expr[:idx])
		right = strings.TrimSpace(expr[idx+2:])
	}
	if op == "" {
		return nil, fmt.Errorf("invalid logical expression")
	}

	leftVal, err := e.evaluate(left, ctx)
	if err != nil {
		return nil, fmt.Errorf("求值逻辑左操作数 %s 失败: %w", left, err)
	}

	rightVal, err := e.evaluate(right, ctx)
	if err != nil {
		return nil, fmt.Errorf("求值逻辑右操作数 %s 失败: %w", right, err)
	}

	leftTruthy := isTruthy(leftVal)
	rightTruthy := isTruthy(rightVal)

	if op == "&&" {
		return leftTruthy && rightTruthy, nil
	}
	return leftTruthy || rightTruthy, nil
}

func (e *complexExpressionImpl) evaluateComparison(expr, op string, ctx EvaluationContext) (any, error) {
	idx := strings.Index(expr, op)
	left := strings.TrimSpace(expr[:idx])
	right := strings.TrimSpace(expr[idx+len(op):])

	leftVal, err := e.evaluate(left, ctx)
	if err != nil {
		return nil, fmt.Errorf("求值比较左操作数 %s 失败: %w", left, err)
	}

	rightVal, err := e.evaluate(right, ctx)
	if err != nil {
		return nil, fmt.Errorf("求值比较右操作数 %s 失败: %w", right, err)
	}

	return compareValues(leftVal, rightVal, op)
}

func (e *complexExpressionImpl) evaluateArithmetic(expr, op string, idx int, ctx EvaluationContext) (any, error) {
	left := strings.TrimSpace(expr[:idx])
	right := strings.TrimSpace(expr[idx+1:])

	leftVal, err := e.evaluate(left, ctx)
	if err != nil {
		return nil, fmt.Errorf("求值算术左操作数 %s 失败: %w", left, err)
	}

	rightVal, err := e.evaluate(right, ctx)
	if err != nil {
		return nil, fmt.Errorf("求值算术右操作数 %s 失败: %w", right, err)
	}

	return arithmetic(leftVal, rightVal, op)
}

// GlobalSpelParser 全局 SpEL 解析器实例。
var GlobalSpelParser = NewSpelParser()

// ParseExpression 解析表达式（便捷函数）。
func ParseExpression(expression string) (Expression, error) {
	return GlobalSpelParser.ParseExpression(expression)
}

// Evaluate 计算表达式（便捷函数）。
func Evaluate(expression string, root any) (any, error) {
	expr, err := ParseExpression(expression)
	if err != nil {
		return nil, fmt.Errorf("解析表达式 %q 失败: %w", expression, err)
	}
	ctx := NewStandardEvaluationContext(root)
	return expr.GetValue(ctx)
}
