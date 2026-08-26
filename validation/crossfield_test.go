package validation

import (
	"reflect"
	"testing"
	"time"
)

// TestFieldMatch 测试字段匹配验证
func TestFieldMatch(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Password        string `validate:"required,min=6"`
		ConfirmPassword string `validate:"fieldmatch=Password"`
	}

	user := User{Password: "password123", ConfirmPassword: "password123"}
	err := validator.Validate(user)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	user.ConfirmPassword = "password456"
	err = validator.Validate(user)
	if err == nil {
		t.Error("预期因密码不匹配而验证失败")
	}
}

// TestFieldNE 测试字段不等于验证
func TestFieldNE(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		OldPassword string `validate:"required"`
		NewPassword string `validate:"required,fieldne=OldPassword"`
	}

	user := User{OldPassword: "oldpass", NewPassword: "newpass"}
	err := validator.Validate(user)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	user.NewPassword = "oldpass"
	err = validator.Validate(user)
	if err == nil {
		t.Error("预期因新密码不能与旧密码相同而验证失败")
	}
}

// TestFieldGT 测试字段大于验证
func TestFieldGT(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type Range struct {
		Min int `validate:"required"`
		Max int `validate:"required,fieldgt=Min"`
	}

	r := Range{Min: 10, Max: 20}
	err := validator.Validate(r)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	r.Max = 10
	err = validator.Validate(r)
	if err == nil {
		t.Error("预期因 Max 必须大于 Min 而验证失败")
	}
}

// TestFieldLT 测试字段小于验证
func TestFieldLT(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type Range struct {
		Min int `validate:"required"`
		Max int `validate:"required,fieldlt=Min"`
	}

	r := Range{Min: 20, Max: 10}
	err := validator.Validate(r)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	r.Max = 20
	err = validator.Validate(r)
	if err == nil {
		t.Error("预期因 Max 必须小于 Min 而验证失败")
	}
}

// TestFieldGTE 测试字段大于等于验证
func TestFieldGTE(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type Range struct {
		Min int `validate:"required"`
		Max int `validate:"required,fieldgte=Min"`
	}

	r := Range{Min: 10, Max: 10}
	err := validator.Validate(r)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	r.Max = 5
	err = validator.Validate(r)
	if err == nil {
		t.Error("预期因 Max 必须大于等于 Min 而验证失败")
	}
}

// TestFieldLTE 测试字段小于等于验证
func TestFieldLTE(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type Range struct {
		Min int `validate:"required"`
		Max int `validate:"required,fieldlte=Min"`
	}

	r := Range{Min: 10, Max: 10}
	err := validator.Validate(r)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	r.Max = 15
	err = validator.Validate(r)
	if err == nil {
		t.Error("预期因 Max 必须小于等于 Min 而验证失败")
	}
}

// TestWhenCondition 测试条件依赖验证
func TestWhenCondition(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Type        string `validate:"required,oneof=personal business"`
		CompanyName string `validate:"when=Type==business:required;min=2"`
	}

	user := User{Type: "personal", CompanyName: ""}
	err := validator.Validate(user)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	user.Type = "business"
	user.CompanyName = ""
	err = validator.Validate(user)
	if err == nil {
		t.Error("预期因 Type 为 business 时 CompanyName 必填而验证失败")
	}

	user.CompanyName = "ABC Company"
	err = validator.Validate(user)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}
}

// TestWhenConditionWithAge 测试年龄条件验证
func TestWhenConditionWithAge(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Age        int    `validate:"required,min=1"`
		ParentName string `validate:"when=Age<18:required"`
	}

	user := User{Age: 25, ParentName: ""}
	err := validator.Validate(user)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	user.Age = 15
	err = validator.Validate(user)
	if err == nil {
		t.Error("预期因 Age 小于 18 时 ParentName 必填而验证失败")
	}

	user.ParentName = "张三"
	err = validator.Validate(user)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}
}

// TestCombinedCrossField 测试组合跨字段验证
func TestCombinedCrossField(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Password        string `validate:"required,min=6"`
		ConfirmPassword string `validate:"required,fieldmatch=Password"`
		Age             int    `validate:"required,min=1"`
		ParentName      string `validate:"when=Age<18:required"`
	}

	user := User{
		Password:        "password123",
		ConfirmPassword: "password123",
		Age:             25,
		ParentName:      "",
	}
	err := validator.Validate(user)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	user.ConfirmPassword = "password456"
	err = validator.Validate(user)
	if err == nil {
		t.Error("预期因密码不匹配而验证失败")
	}
}

func TestCompareFieldValue_Int(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type Config struct {
		Min int
		Max int
	}

	cfg := Config{Min: 10, Max: 20}
	if !validator.compareFieldValue(cfg, "Max", "15", ">") {
		t.Error("expected Max > 15 to be true")
	}
	if !validator.compareFieldValue(cfg, "Max", "20", ">=") {
		t.Error("expected Max >= 20 to be true")
	}
	if !validator.compareFieldValue(cfg, "Min", "15", "<") {
		t.Error("expected Min < 15 to be true")
	}
	if !validator.compareFieldValue(cfg, "Min", "10", "<=") {
		t.Error("expected Min <= 10 to be true")
	}
}

func TestCompareFieldValue_Float(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type Config struct {
		Price float64
	}

	cfg := Config{Price: 19.99}
	if !validator.compareFieldValue(cfg, "Price", "15.0", ">") {
		t.Error("expected Price > 15.0 to be true")
	}
	if !validator.compareFieldValue(cfg, "Price", "20.0", "<") {
		t.Error("expected Price < 20.0 to be true")
	}
}

func TestCompareFieldValue_String(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type Config struct {
		Name string
	}

	cfg := Config{Name: "hello"}
	if !validator.compareFieldValue(cfg, "Name", "hi", ">") {
		t.Error("expected 'hello' length > 'hi' length to be true")
	}
	if !validator.compareFieldValue(cfg, "Name", "hello world", "<") {
		t.Error("expected 'hello' length < 'hello world' length to be true")
	}
}

func TestCompareFieldValue_Uint(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type Config struct {
		Count uint
	}

	cfg := Config{Count: 100}
	if !validator.compareFieldValue(cfg, "Count", "50", ">") {
		t.Error("expected Count > 50 to be true")
	}
	if !validator.compareFieldValue(cfg, "Count", "200", "<") {
		t.Error("expected Count < 200 to be true")
	}
}

func TestCompareFieldValue_InvalidType(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type Config struct {
		Data time.Time
	}

	cfg := Config{Data: time.Now()}
	if validator.compareFieldValue(cfg, "Data", "2024-01-01", ">") {
		t.Error("expected time.Time comparison to return false")
	}
}

func TestCompareFieldValue_InvalidValue(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type Config struct {
		Count int
	}

	cfg := Config{Count: 10}
	if validator.compareFieldValue(cfg, "Count", "not_a_number", ">") {
		t.Error("expected invalid number parsing to return false")
	}
}

func TestCompareFieldValue_NilValue(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type Config struct {
		Name *string
	}

	cfg := Config{Name: nil}
	if validator.compareFieldValue(cfg, "Name", "test", ">") {
		t.Error("expected nil value comparison to return false")
	}
}

func TestEvaluateCondition_Equal(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Type string
	}

	user := User{Type: "admin"}
	if !validator.evaluateCondition("Type==admin", user) {
		t.Error("expected Type==admin to be true")
	}
	if validator.evaluateCondition("Type==user", user) {
		t.Error("expected Type==user to be false")
	}
}

func TestEvaluateCondition_NotEqual(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Status string
	}

	user := User{Status: "active"}
	if validator.evaluateCondition("Status!=active", user) {
		t.Error("expected Status!=active to be false")
	}
	if !validator.evaluateCondition("Status!=inactive", user) {
		t.Error("expected Status!=inactive to be true")
	}
}

func TestEvaluateCondition_LessThan(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Age int
	}

	user := User{Age: 15}
	if !validator.evaluateCondition("Age<18", user) {
		t.Error("expected Age<18 to be true")
	}
	if validator.evaluateCondition("Age<10", user) {
		t.Error("expected Age<10 to be false")
	}
}

func TestEvaluateCondition_GreaterThan(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Score int
	}

	user := User{Score: 85}
	if !validator.evaluateCondition("Score>80", user) {
		t.Error("expected Score>80 to be true")
	}
	if validator.evaluateCondition("Score>90", user) {
		t.Error("expected Score>90 to be false")
	}
}

func TestEvaluateCondition_LessThanOrEqual(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Age int
	}

	user := User{Age: 18}
	if !validator.evaluateCondition("Age<=18", user) {
		t.Error("expected Age<=18 to be true")
	}
	if validator.evaluateCondition("Age<=15", user) {
		t.Error("expected Age<=15 to be false")
	}
}

func TestEvaluateCondition_GreaterThanOrEqual(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Score int
	}

	user := User{Score: 80}
	if !validator.evaluateCondition("Score>=80", user) {
		t.Error("expected Score>=80 to be true")
	}
	if validator.evaluateCondition("Score>=90", user) {
		t.Error("expected Score>=90 to be false")
	}
}

func TestEvaluateCondition_InvalidCondition(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Name string
	}

	user := User{Name: "test"}
	if validator.evaluateCondition("InvalidCondition", user) {
		t.Error("expected invalid condition to return false")
	}
}

func TestGetFieldValue_ValidField(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Name string
	}

	user := User{Name: "John"}
	val := validator.getFieldValue(user, "Name")
	if val != "John" {
		t.Errorf("expected 'John', got %v", val)
	}
}

func TestGetFieldValue_InvalidField(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Name string
	}

	user := User{Name: "John"}
	val := validator.getFieldValue(user, "NonExistent")
	if val != nil {
		t.Errorf("expected nil for non-existent field, got %v", val)
	}
}

func TestGetFieldValue_Pointer(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Name string
	}

	user := &User{Name: "Jane"}
	val := validator.getFieldValue(user, "Name")
	if val != "Jane" {
		t.Errorf("expected 'Jane', got %v", val)
	}
}

func TestFieldsEqual_DifferentTypes(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	v1 := reflect.ValueOf("test")
	v2 := reflect.ValueOf(123)

	if validator.fieldsEqual(v1, v2) {
		t.Error("expected different types to not be equal")
	}
}

func TestFieldGreaterThan_UnsupportedType(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type Config struct {
		Data []int
	}

	cfg := Config{Data: []int{1, 2, 3}}
	v1 := reflect.ValueOf(cfg.Data)
	v2 := reflect.ValueOf([]int{4, 5})

	if validator.fieldGreaterThan(v1, v2) {
		t.Error("expected slice comparison to return false")
	}
}

func TestFieldLessThanOrEqual_String(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	v1 := reflect.ValueOf("hi")
	v2 := reflect.ValueOf("hello")

	if !validator.fieldLessThanOrEqual(v1, v2) {
		t.Error("expected 'hi' length <= 'hello' length")
	}
}

func TestValidateCrossField_InvalidRule(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Password        string
		ConfirmPassword string
	}

	user := User{Password: "123", ConfirmPassword: "456"}
	err := validator.validateCrossField(reflect.ValueOf(user.ConfirmPassword), "invalidrule", "ConfirmPassword", user)
	if err == nil {
		t.Error("expected error for invalid rule format")
	}
}

func TestValidateCrossField_NonExistentField(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Password string
	}

	user := User{Password: "123"}
	err := validator.validateCrossField(reflect.ValueOf(""), "fieldmatch=NonExistent", "Password", user)
	if err == nil {
		t.Error("expected error for non-existent field")
	}
}

func TestValidateWhenCondition_InvalidRule(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Name string
	}

	user := User{Name: "test"}
	errors := validator.validateWhenCondition(reflect.ValueOf(user.Name), "invalid", "Name", user)
	if len(errors) > 0 {
		t.Errorf("expected no errors for invalid rule, got %v", errors)
	}
}

func TestValidateWhenCondition_InvalidConditionFormat(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Name string
	}

	user := User{Name: "test"}
	errors := validator.validateWhenCondition(reflect.ValueOf(user.Name), "when=condition", "Name", user)
	if len(errors) > 0 {
		t.Errorf("expected no errors for invalid condition format, got %v", errors)
	}
}

func TestInterfaceOrNil_CanInterface(t *testing.T) {
	t.Parallel()

	v := reflect.ValueOf("test")
	result := interfaceOrNil(v)
	if result != "test" {
		t.Errorf("expected 'test', got %v", result)
	}
}

func TestInterfaceOrNil_CannotInterface(t *testing.T) {
	t.Parallel()

	type internal struct {
		unexported string
	}

	obj := internal{unexported: "secret"}
	v := reflect.ValueOf(obj).FieldByName("unexported")
	result := interfaceOrNil(v)
	if result != nil {
		t.Errorf("expected nil for unexported field, got %v", result)
	}
}
