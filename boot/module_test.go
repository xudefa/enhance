package boot

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/core"
)

// 测试用数据库类型
type testDatabase struct {
	URL string
}

// 测试用仓库类型
type testRepository struct {
	DB *testDatabase
}

func TestModule_Install(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()

	// 创建模块
	dbModule := NewModule("database",
		Provide(func() *testDatabase {
			return &testDatabase{URL: "localhost:5432"}
		}),
		Provide(func(db *testDatabase) *testRepository {
			return &testRepository{DB: db}
		}),
	)

	// 安装模块
	err := dbModule.Install(container)
	if err != nil {
		t.Fatalf("failed to install module: %v", err)
	}

	// 列出所有 Bean
	beans := container.ListBeans()
	t.Logf("registered beans: %v", beans)

	// 获取 Bean
	repo, err := container.Get(reflect.TypeFor[*testRepository]())
	if err != nil {
		t.Fatalf("failed to get repository: %v", err)
	}

	repository := repo[0].(*testRepository)
	if repository.DB.URL != "localhost:5432" {
		t.Errorf("expected URL 'localhost:5432', got '%s'", repository.DB.URL)
	}
}

func TestModule_Invoke(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()

	type Database struct {
		URL      string
		Migrated bool
	}

	invoked := false

	// 创建模块
	dbModule := NewModule("database",
		Provide(func() *Database {
			return &Database{URL: "localhost:5432"}
		}),
		Invoke(func(db *Database) error {
			db.Migrated = true
			invoked = true
			return nil
		}),
	)

	// 安装模块
	err := dbModule.Install(container)
	if err != nil {
		t.Fatalf("failed to install module: %v", err)
	}

	if !invoked {
		t.Error("expected invoke function to be called")
	}

	// 验证数据库已迁移
	db, err := container.Get(reflect.TypeFor[*Database]())
	if err != nil {
		t.Fatalf("failed to get database: %v", err)
	}

	database := db[0].(*Database)
	if !database.Migrated {
		t.Error("expected database to be migrated")
	}
}

func TestModule_WithApplication(t *testing.T) {
	t.Parallel()
	// 定义模块
	dbModule := NewModule("database",
		Provide(func() string {
			return "test-db"
		}),
	)

	// 创建应用并安装模块
	app, err := NewApplication(
		WithAppName("test-app"),
		WithModulesOption(dbModule),
	)
	if err != nil {
		t.Fatalf("failed to create application: %v", err)
	}

	// 启动应用（会安装模块）
	err = app.Start()
	if err != nil {
		t.Fatalf("failed to start application: %v", err)
	}

	// 验证模块已安装
	dbs, err := app.Container().Get(reflect.TypeFor[string]())
	if err != nil {
		t.Fatalf("failed to get database bean: %v", err)
	}

	if dbs[0].(string) != "test-db" {
		t.Errorf("expected 'test-db', got '%s'", dbs[0])
	}

	// 停止应用
	err = app.Stop()
	if err != nil {
		t.Fatalf("failed to stop application: %v", err)
	}
}
