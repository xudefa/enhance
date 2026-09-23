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

	if !strings.HasSuffix(expr, ")") {
		return nil, fmt.Errorf("invalid method call: missing closing parenthesis")
	}

	methodName := strings.TrimSpace(expr[:dotIdx])
	argsStr := expr[dotIdx+1 : len(expr)-1]

	method, err := e.resolveRootMethod(ctx, methodName)
	if err != nil {
		return nil, fmt.Errorf("resolve root method: %w", err)
	}

	args, err := e.evaluateMethodCallArgs(argsStr, ctx, method, methodName)
	if err != nil {
		return nil, fmt.Errorf("evaluate method call args: %w", err)
	}

	results := method.Call(args)
	if len(results) == 0 {
		return nil, nil
	}

	return results[0].Interface(), nil
}

// resolveRootMethod 从根对象上解析指定名称的方法，找不到时返回错误。
func (e *complexExpressionImpl) resolveRootMethod(ctx EvaluationContext, methodName string) (reflect.Value, error) {
	root := ctx.GetRootObject()
	if root == nil {
		return reflect.Value{}, fmt.Errorf("root object is nil")
	}

	rootValue := reflect.ValueOf(root)
	if rootValue.Kind() == reflect.Pointer {
		rootValue = rootValue.Elem()
	}

	method := rootValue.MethodByName(methodName)
	if !method.IsValid() && rootValue.CanAddr() {
		method = rootValue.Addr().MethodByName(methodName)
	}

	if !method.IsValid() {
		return reflect.Value{}, fmt.Errorf("method %s not found", methodName)
	}

	return method, nil
}

// evaluateMethodCallArgs 解析并转换方法调用参数。
func (e *complexExpressionImpl) evaluateMethodCallArgs(argsStr string, ctx EvaluationContext, method reflect.Value, methodName string) ([]reflect.Value, error) {
	var args []reflect.Value
	if argsStr != "" {
		var err error
		args, err = e.evaluateMethodArgs(argsStr, ctx)
		if err != nil {
			return nil, fmt.Errorf("evaluate method args: %w", err)
		}
	}

	if method.Type().NumIn() != len(args) {
		return nil, fmt.Errorf("method %s expects %d arguments, got %d",
			methodName, method.Type().NumIn(), len(args))
	}

	if err := convertArgsToMethodParams(method.Type(), args, methodName); err != nil {
		return nil, fmt.Errorf("convert args to method params: %w", err)
	}

	return args, nil
}

// evaluateMethodArgs 求值方法调用参数列表。
func (e *complexExpressionImpl) evaluateMethodArgs(argsStr string, ctx EvaluationContext) ([]reflect.Value, error) {
	parsedArgs := splitArgsRespectingQuotes(argsStr)
	args := make([]reflect.Value, 0, len(parsedArgs))
	for _, arg := range parsedArgs {
		argValue, err := e.evaluate(strings.TrimSpace(arg), ctx)
		if err != nil {
			return nil, fmt.Errorf("求值方法参数 %s 失败: %w", arg, err)
		}
		args = append(args, reflect.ValueOf(argValue))
	}
	return args, nil
}

// convertArgsToMethodParams 将参数值转换为方法期望的参数类型。
func convertArgsToMethodParams(methodType reflect.Type, args []reflect.Value, methodName string) error {
	mt := methodType
	for i := range args {
		target := mt.In(i)
		arg := args[i]
		if !arg.IsValid() {
			if isNilable(target) {
				args[i] = reflect.Zero(target)
				continue
			}
			return fmt.Errorf("method %s argument %d: cannot pass nil to %s", methodName, i, target)
		}
		if arg.Type().AssignableTo(target) {
			continue
		}
		if arg.Type().ConvertibleTo(target) {
			args[i] = arg.Convert(target)
			continue
		}
		return fmt.Errorf("method %s argument %d: cannot convert %s to %s",
			methodName, i, arg.Type(), target)
	}
	return nil
}

func (e *complexExpressionImpl) evaluatePropertyChain(expr string, ctx EvaluationContext) (any, error) {
	parts := strings.Split(expr, ".")
	current := ctx.GetRootObject()

	for _, part := range parts {
		if current == nil {
			return nil, fmt.Errorf("cannot access property %s on nil", part)
		}

		propValue, err := ctx.GetPropertyAccessor().GetProperty(current, part)
		if err != nil {
			return nil, fmt.Errorf("获取属性链属性 %s 失败: %w", part, err)
		}
		current = propValue
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

	if variableValue, ok := ctx.GetVariable(expr); ok {
		return variableValue, nil
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
	switch op {
	case "==":
		return equals(left, right), nil
	case "!=":
		return !equals(left, right), nil
	case ">", "<", ">=", "<=":
		lf, lok := toFloat64(left)
		rf, rok := toFloat64(right)
		if !lok || !rok {
			return false, fmt.Errorf("cannot compare %T and %T", left, right)
		}
		switch op {
		case ">":
			return lf > rf, nil
		case "<":
			return lf < rf, nil
		case ">=":
			return lf >= rf, nil
		case "<=":
			return lf <= rf, nil
		}
	}
	return false, fmt.Errorf("unsupported operator: %s", op)
}

// toFloat64 converts any numeric type to float64. Returns false for non-numeric types.
func toFloat64(v any) (float64, bool) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return rv.Float(), true
	}
	return 0, false
}

func arithmetic(left, right any, op string) (any, error) {
	lf, lok := toFloat64(left)
	rf, rok := toFloat64(right)
	if !lok || !rok {
		return nil, fmt.Errorf("unsupported operand types: %T and %T", left, right)
	}

	switch op {
	case "+":
		return lf + rf, nil
	case "-":
		return lf - rf, nil
	case "*":
		return lf * rf, nil
	case "/":
		if rf == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return lf / rf, nil
	default:
		return nil, fmt.Errorf("unsupported operator: %s", op)
	}
}

func splitArgsRespectingQuotes(argsStr string) []string {
	var args []string
	var current strings.Builder
	inQuote := false

	for _, ch := range argsStr {
		switch {
		case ch == '\'':
			inQuote = !inQuote
			current.WriteRune(ch)
		case ch == ',' && !inQuote:
			args = append(args, current.String())
			current.Reset()
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
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
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}

	var parsed int64
	negative := false
	pos := 0

	if len(s) > 0 && s[0] == '-' {
		negative = true
		pos = 1
	}

	for ; pos < len(s); pos++ {
		c := s[pos]
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid integer: %s", s)
		}
		parsed = parsed*10 + int64(c-'0')
	}

	if negative {
		return -parsed, nil
	}
	return parsed, nil
}

// parseFloat 简单浮点数解析。
func parseFloat(s string, _ int) (float64, error) {
	var intPart, fracPart float64
	var negative bool
	var inFrac bool
	var fracDiv float64 = 1

	pos := 0
	if len(s) > 0 && s[0] == '-' {
		negative = true
		pos = 1
	}

	for ; pos < len(s); pos++ {
		c := s[pos]
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
			fracPart += digit / fracDiv
			continue
		}
		intPart = intPart*10 + digit
	}

	parsedValue := intPart + fracPart
	if negative {
		parsedValue = -parsedValue
	}
	return parsedValue, nil
}
