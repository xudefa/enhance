package testing

import (
	"reflect"
	"testing"
)

// Assert 断言条件为真。
func Assert(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Fatal(msg)
	}
}

// AssertEqual 断言两个值相等。
func AssertEqual(t *testing.T, expected, actual any, msg ...string) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		message := "assertion failed"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Fatalf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNoError 断言无错误。
func AssertNoError(t *testing.T, err error, msg ...string) {
	t.Helper()
	if err != nil {
		message := "unexpected error"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Fatalf("%s: %v", message, err)
	}
}

// AssertError 断言有错误。
func AssertError(t *testing.T, err error, msg ...string) {
	t.Helper()
	if err == nil {
		message := "expected error but got none"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Fatal(message)
	}
}

// AssertNil 断言值为 nil。
func AssertNil(t *testing.T, value any, msg ...string) {
	t.Helper()
	if value != nil {
		message := "expected nil"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Fatalf("%s: got %v", message, value)
	}
}

// AssertNotNil 断言值不为 nil。
func AssertNotNil(t *testing.T, value any, msg ...string) {
	t.Helper()
	if value == nil {
		message := "expected not nil"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Fatal(message)
	}
}

// AssertTrue 断言条件为真。
func AssertTrue(t *testing.T, condition bool, msg ...string) {
	t.Helper()
	if !condition {
		message := "expected true"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Fatal(message)
	}
}

// AssertFalse 断言条件为假。
func AssertFalse(t *testing.T, condition bool, msg ...string) {
	t.Helper()
	if condition {
		message := "expected false"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Fatal(message)
	}
}

// SkipIf 满足条件时跳过测试。
func SkipIf(t *testing.T, condition bool, reason string) {
	t.Helper()
	if condition {
		t.Skip(reason)
	}
}
