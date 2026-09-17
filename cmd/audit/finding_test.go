package main

import "testing"

func TestDetect_AllFindingsComplete(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, "package p\nfunc F() int { return 0 }\n", false)
	findings := detect(f)
	for _, fd := range findings {
		if fd.Category == "" {
			t.Errorf("发现缺失类别: %+v", fd)
		}
		if fd.File == "" {
			t.Errorf("发现缺失文件路径: %+v", fd)
		}
	}
}

func TestDetect_TestFileSkipsUnallowedCategories(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
import "testing"
func TestX(t *testing.T) {
	err := doSomething()
	_ = err
	for i := 0; i < 10; i++ {
		_ = i
	}
}`, true)
	findings := detect(f)
	for _, fd := range findings {
		if fd.Category == CategoryErrorContext ||
			fd.Category == CategoryDocumentation ||
			fd.Category == CategoryMagicNumbers {
			t.Errorf("测试文件不应报告 %s: %+v", fd.Category, fd)
		}
	}
}
