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

	// 缺少右括号时避免切片越界
	if !strings.HasSuffix(expr, ")") {
		return nil, fmt.Errorf("invalid method call: missing closing parenthesis")
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
		parsedArgs := splitArgsRespectingQuotes(argsStr)
		for _, arg := range parsedArgs {
			val, err := e.evaluate(strings.TrimSpace(arg), ctx)
			if err != nil {
				return nil, err
			}
			args = append(args, reflect.ValueOf(val))
		}
	}

	if method.Type().NumIn() != len(args) {
		return nil, fmt.Errorf("method %s expects %d arguments, got %d",
			methodName, method.Type().NumIn(), len(args))
	}

	// 将参数转换为方法期望的类型，避免类型不匹配导致 reflect.Call panic
	mt := method.Type()
	for i := range args {
		target := mt.In(i)
		arg := args[i]
		if !arg.IsValid() {
			if isNilable(target) {
				args[i] = reflect.Zero(target)
				continue
			}
			return nil, fmt.Errorf("method %s argument %d: cannot pass nil to %s", methodName, i, target)
		}
		if arg.Type().AssignableTo(target) {
			continue
		}
		if arg.Type().ConvertibleTo(target) {
			args[i] = arg.Convert(target)
			continue
		}
		return nil, fmt.Errorf("method %s argument %d: cannot convert %s to %s",
			methodName, i, arg.Type(), target)
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

// equals 比较两个值是否相等（替代 reflect.DeepEqual）。
// 使用类型安全的比较，避免 DeepEqual 在 nil/empty slice、不同数值类型上的不一致行为。
func equals(left, right any) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}

	switch l := left.(type) {
	case bool:
		r, ok := right.(bool)
		return ok && l == r
	case string:
		r, ok := right.(string)
		return ok && l == r
	case int:
		switch r := right.(type) {
		case int:
			return l == r
		case int8:
			return l == int(r)
		case int16:
			return l == int(r)
		case int32:
			return l == int(r)
		case int64:
			return int64(l) == r
		case uint:
			return l >= 0 && uint(l) == r
		case uint8:
			return l >= 0 && uint8(l) == r
		case uint16:
			return l >= 0 && uint16(l) == r
		case uint32:
			return l >= 0 && uint32(l) == r
		case uint64:
			return l >= 0 && uint64(l) == r
		case float32:
			return float64(l) == float64(r)
		case float64:
			return float64(l) == r
		}
	case int8:
		switch r := right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return equals(int64(l), toInt64(r))
		case float32, float64:
			return equals(float64(l), toFloat64Value(r))
		}
	case int16:
		switch r := right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return equals(int64(l), toInt64(r))
		case float32, float64:
			return equals(float64(l), toFloat64Value(r))
		}
	case int32:
		switch r := right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return equals(int64(l), toInt64(r))
		case float32, float64:
			return equals(float64(l), toFloat64Value(r))
		}
	case int64:
		switch r := right.(type) {
		case int:
			return l == int64(r)
		case int8:
			return l == int64(r)
		case int16:
			return l == int64(r)
		case int32:
			return l == int64(r)
		case int64:
			return l == r
		case uint:
			return l >= 0 && uint64(l) == uint64(r)
		case uint8:
			return l >= 0 && uint64(l) == uint64(r)
		case uint16:
			return l >= 0 && uint64(l) == uint64(r)
		case uint32:
			return l >= 0 && uint64(l) == uint64(r)
		case uint64:
			return l >= 0 && uint64(l) == r
		case float32:
			return float64(l) == float64(r)
		case float64:
			return float64(l) == r
		}
	case uint:
		switch r := right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return equals(uint64(l), toUint64(r))
		case float32, float64:
			return equals(float64(l), toFloat64Value(r))
		}
	case uint8:
		switch r := right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return equals(uint64(l), toUint64(r))
		case float32, float64:
			return equals(float64(l), toFloat64Value(r))
		}
	case uint16:
		switch r := right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return equals(uint64(l), toUint64(r))
		case float32, float64:
			return equals(float64(l), toFloat64Value(r))
		}
	case uint32:
		switch r := right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return equals(uint64(l), toUint64(r))
		case float32, float64:
			return equals(float64(l), toFloat64Value(r))
		}
	case uint64:
		switch r := right.(type) {
		case int:
			return int64(l) >= 0 && l == uint64(r)
		case int8:
			return l == uint64(r)
		case int16:
			return l == uint64(r)
		case int32:
			return l == uint64(r)
		case int64:
			return r >= 0 && l == uint64(r)
		case uint:
			return l == uint64(r)
		case uint8:
			return l == uint64(r)
		case uint16:
			return l == uint64(r)
		case uint32:
			return l == uint64(r)
		case uint64:
			return l == r
		case float32:
			return float64(l) == float64(r)
		case float64:
			return float64(l) == r
		}
	case float32:
		switch r := right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return equals(float64(l), toFloat64Value(r))
		}
	case float64:
		switch r := right.(type) {
		case int:
			return l == float64(r)
		case int8:
			return l == float64(r)
		case int16:
			return l == float64(r)
		case int32:
			return l == float64(r)
		case int64:
			return l == float64(r)
		case uint:
			return l == float64(r)
		case uint8:
			return l == float64(r)
		case uint16:
			return l == float64(r)
		case uint32:
			return l == float64(r)
		case uint64:
			return l == float64(r)
		case float32:
			return l == float64(r)
		case float64:
			return l == r
		}
	default:
		// 对于无法直接比较的类型，使用 reflect.DeepEqual 作为兜底
		return reflect.DeepEqual(left, right)
	}
	return false
}

// toInt64 attempts to convert v to int64.
func toInt64(v any) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int8:
		return int64(val)
	case int16:
		return int64(val)
	case int32:
		return int64(val)
	case int64:
		return val
	}
	return 0
}

// toUint64 attempts to convert v to uint64.
func toUint64(v any) uint64 {
	switch val := v.(type) {
	case uint:
		return uint64(val)
	case uint8:
		return uint64(val)
	case uint16:
		return uint64(val)
	case uint32:
		return uint64(val)
	case uint64:
		return val
	}
	return 0
}

// toFloat64Value attempts to convert v to float64.
func toFloat64Value(v any) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	}
	return 0
}
