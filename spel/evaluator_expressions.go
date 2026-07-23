package spel

import (
	"fmt"
	"reflect"
	"strings"
)

func (e *complexExpressionImpl) evaluateMethodCall(expr string, ctx EvaluationContext) (any, error) {
	dotIdx := strings.Index(expr, "(")
	if dotIdx < 0 {
		return nil, fmt.Errorf("invalid method call")
	}

	methodName := strings.TrimSpace(expr[:dotIdx])
	argsStr := expr[dotIdx+1 : len(expr)-1]

	root := ctx.GetRootObject()
	if root == nil {
		return nil, fmt.Errorf("root object is nil")
	}

	v := reflect.ValueOf(root)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	// 先尝试在值上查找方法，如果找不到，再尝试在指针上查找
	method := v.MethodByName(methodName)
	if !method.IsValid() && v.CanAddr() {
		method = v.Addr().MethodByName(methodName)
	}

	if !method.IsValid() {
		return nil, fmt.Errorf("method %s not found", methodName)
	}

	var args []reflect.Value
	if argsStr != "" {
		for _, arg := range strings.Split(argsStr, ",") {
			val, err := e.evaluate(strings.TrimSpace(arg), ctx)
			if err != nil {
				return nil, err
			}
			args = append(args, reflect.ValueOf(val))
		}
	}

	results := method.Call(args)
	if len(results) == 0 {
		return nil, nil
	}

	return results[0].Interface(), nil
}

func (e *complexExpressionImpl) evaluatePropertyChain(expr string, ctx EvaluationContext) (any, error) {
	parts := strings.Split(expr, ".")
	current := ctx.GetRootObject()

	for _, part := range parts {
		if current == nil {
			return nil, fmt.Errorf("cannot access property %s on nil", part)
		}

		val, err := ctx.GetPropertyAccessor().GetProperty(current, part)
		if err != nil {
			return nil, err
		}
		current = val
	}

	return current, nil
}

func (e *complexExpressionImpl) evaluateLiteral(expr string) (any, error) {
	expr = strings.TrimSpace(expr)

	if strings.HasPrefix(expr, "'") && strings.HasSuffix(expr, "'") {
		return expr[1 : len(expr)-1], nil
	}

	if expr == "true" {
		return true, nil
	}
	if expr == "false" {
		return false, nil
	}

	if expr == "null" {
		return nil, nil
	}

	if num, err := parseInt(expr, 10, 64); err == nil {
		return num, nil
	}
	if num, err := parseFloat(expr, 64); err == nil {
		return num, nil
	}

	return nil, fmt.Errorf("unknown literal: %s", expr)
}

func (e *complexExpressionImpl) evaluate(expr string, ctx EvaluationContext) (any, error) {
	expr = strings.TrimSpace(expr)

	if expr == "true" || expr == "false" || expr == "null" ||
		strings.HasPrefix(expr, "'") || isNumber(expr) {
		return e.evaluateLiteral(expr)
	}

	if val, ok := ctx.GetVariable(expr); ok {
		return val, nil
	}

	// 递归处理逻辑运算符
	if strings.Contains(expr, "&&") || strings.Contains(expr, "||") {
		return e.evaluateLogical(expr, ctx)
	}

	// 递归处理比较运算符
	for _, op := range []string{"==", "!=", ">=", "<=", ">", "<"} {
		if idx := strings.Index(expr, op); idx > 0 {
			return e.evaluateComparison(expr, op, ctx)
		}
	}

	// 递归处理算术运算符
	for _, op := range []string{"+", "-", "*", "/"} {
		if idx := strings.LastIndex(expr, op); idx > 0 {
			return e.evaluateArithmetic(expr, op, idx, ctx)
		}
	}

	if strings.Contains(expr, ".") {
		return e.evaluatePropertyChain(expr, ctx)
	}

	root := ctx.GetRootObject()
	if root != nil {
		return ctx.GetPropertyAccessor().GetProperty(root, expr)
	}

	return nil, fmt.Errorf("unable to evaluate: %s", expr)
}

func compareValues(left, right any, op string) (bool, error) {
	lv := reflect.ValueOf(left)
	rv := reflect.ValueOf(right)

	if lv.Kind() != rv.Kind() {
		if lv.Kind() >= reflect.Int && lv.Kind() <= reflect.Uint64 {
			left = lv.Convert(reflect.TypeFor[int64]()).Int()
		}
		if rv.Kind() >= reflect.Int && rv.Kind() <= reflect.Uint64 {
			right = rv.Convert(reflect.TypeFor[int64]()).Int()
		}
	}

	switch op {
	case "==":
		return reflect.DeepEqual(left, right), nil
	case "!=":
		return !reflect.DeepEqual(left, right), nil
	case ">":
		return left.(int64) > right.(int64), nil
	case "<":
		return left.(int64) < right.(int64), nil
	case ">=":
		return left.(int64) >= right.(int64), nil
	case "<=":
		return left.(int64) <= right.(int64), nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", op)
	}
}

func arithmetic(left, right any, op string) (any, error) {
	lv := reflect.ValueOf(left)
	rv := reflect.ValueOf(right)

	var ln, rn int64
	if lv.Kind() >= reflect.Int && lv.Kind() <= reflect.Uint64 {
		ln = lv.Convert(reflect.TypeFor[int64]()).Int()
	}
	if rv.Kind() >= reflect.Int && rv.Kind() <= reflect.Uint64 {
		rn = rv.Convert(reflect.TypeFor[int64]()).Int()
	}

	switch op {
	case "+":
		return ln + rn, nil
	case "-":
		return ln - rn, nil
	case "*":
		return ln * rn, nil
	case "/":
		if rn == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return ln / rn, nil
	default:
		return nil, fmt.Errorf("unsupported operator: %s", op)
	}
}

func isTruthy(v any) bool {
	if v == nil {
		return false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() != 0
	case reflect.String:
		return rv.String() != ""
	default:
		return true
	}
}

func isNumber(s string) bool {
	if s == "" {
		return false
	}
	_, err := parseInt(s, 10, 64)
	if err == nil {
		return true
	}
	_, err = parseFloat(s, 64)
	return err == nil
}

// parseInt 简单整数解析。
func parseInt(s string, _, _ int) (int64, error) {
	var n int64
	negative := false
	i := 0

	if len(s) > 0 && s[0] == '-' {
		negative = true
		i = 1
	}

	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid integer: %s", s)
		}
		n = n*10 + int64(c-'0')
	}

	if negative {
		return -n, nil
	}
	return n, nil
}

// parseFloat 简单浮点数解析。
func parseFloat(s string, _ int) (float64, error) {
	var n, frac float64
	var negative bool
	var inFrac bool
	var fracDiv float64 = 1

	i := 0
	if len(s) > 0 && s[0] == '-' {
		negative = true
		i = 1
	}

	for ; i < len(s); i++ {
		c := s[i]
		if c == '.' {
			inFrac = true
			continue
		}
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid float: %s", s)
		}
		digit := float64(c - '0')
		if inFrac {
			fracDiv *= 10
			frac += digit / fracDiv
			continue
		}
		n = n*10 + digit
	}

	result := n + frac
	if negative {
		result = -result
	}
	return result, nil
}
