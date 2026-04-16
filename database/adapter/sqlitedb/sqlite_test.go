package sqlitedb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/f2xme/gox/database"
	"github.com/f2xme/gox/database/adapter/gormbase"
	"github.com/f2xme/gox/database/adapter/sqlitedb"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type TestUser struct {
	ID   uint   `gorm:"primaryKey"`
	Name string
}

// TestNew 测试创建 SQLite 数据库连接
func TestNew(t *testing.T) {
	db, err := sqlitedb.New(":memory:")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if db == nil {
		t.Fatal("expected non-nil database")
	}
	defer db.Close()

	if err := db.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

// TestUnwrap 测试获取底层 GORM 引擎
func TestUnwrap(t *testing.T) {
	db, err := sqlitedb.New(":memory:")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	engine := db.Unwrap()
	if engine == nil {
		t.Fatal("Unwrap() returned nil")
	}

	gormDB, ok := engine.(*gorm.DB)
	if !ok {
		t.Fatalf("Unwrap() type = %T, want *gorm.DB", engine)
	}

	if gormDB == nil {
		t.Error("Unwrap() returned nil *gorm.DB")
	}
}

// TestTransaction_Success 测试事务成功提交
func TestTransaction_Success(t *testing.T) {
	db, err := sqlitedb.New(":memory:")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	ctx := context.Background()
	err = db.Transaction(ctx, func(tx database.DB) error {
		return tx.Create(ctx, &TestUser{Name: "Alice"})
	})

	if err != nil {
		t.Errorf("Transaction() error = %v, want nil", err)
	}

	// Verify the record was committed
	var count int64
	if err := db.Model(&TestUser{}).Count(ctx, &count); err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 record after commit, got %d", count)
	}
}

// TestTransaction_Rollback 测试事务回滚
func TestTransaction_Rollback(t *testing.T) {
	db, err := sqlitedb.New(":memory:")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	ctx := context.Background()
	expectedErr := errors.New("rollback test")
	err = db.Transaction(ctx, func(tx database.DB) error {
		if createErr := tx.Create(ctx, &TestUser{Name: "Bob"}); createErr != nil {
			return createErr
		}
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("Transaction() error = %v, want %v", err, expectedErr)
	}

	// Verify the record was rolled back
	var count int64
	if err := db.Model(&TestUser{}).Count(ctx, &count); err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 records after rollback, got %d", count)
	}
}

// TestAutoMigrate 测试自动迁移表结构
func TestAutoMigrate(t *testing.T) {
	db, err := sqlitedb.New(":memory:")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	err = db.AutoMigrate(&TestUser{})
	if err != nil {
		t.Errorf("AutoMigrate() error = %v", err)
	}

	gormDB := db.Unwrap().(*gorm.DB)
	hasTable := gormDB.Migrator().HasTable(&TestUser{})
	if !hasTable {
		t.Error("AutoMigrate() did not create table")
	}
}

// TestClose 测试关闭数据库连接
func TestClose(t *testing.T) {
	db, err := sqlitedb.New(":memory:")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Verify connection is closed by attempting to use it
	gormDB := db.Unwrap().(*gorm.DB)
	sqlDB, _ := gormDB.DB()
	if err := sqlDB.Ping(); err == nil {
		t.Error("expected error pinging closed connection")
	}
}

// TestWithMaxOpenConns 测试设置最大打开连接数
func TestWithMaxOpenConns(t *testing.T) {
	db, err := sqlitedb.New(":memory:", gormbase.WithMaxOpenConns(50))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	gormDB := db.Unwrap().(*gorm.DB)
	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}

	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 50 {
		t.Errorf("MaxOpenConnections = %d, want 50", stats.MaxOpenConnections)
	}
}

// TestWithLogger 测试自定义日志记录器
func TestWithLogger(t *testing.T) {
	customLogger := logger.Default.LogMode(logger.Silent)
	db, err := sqlitedb.New(":memory:", gormbase.WithLogger(customLogger))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Errorf("AutoMigrate with custom logger failed: %v", err)
	}
}

// TestWithDryRun 测试 DryRun 模式（不执行实际 SQL）
func TestWithDryRun(t *testing.T) {
	db, err := sqlitedb.New(":memory:", gormbase.WithDryRun())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	gormDB := db.Unwrap().(*gorm.DB)
	result := gormDB.Create(&TestUser{Name: "Charlie"})

	if result.Error != nil {
		t.Errorf("Create() in dry run returned error: %v", result.Error)
	}

	var count int64
	gormDB.Model(&TestUser{}).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 records in dry run mode, got %d", count)
	}
}

// TestWithSingularTable 测试使用单数表名
func TestWithSingularTable(t *testing.T) {
	db, err := sqlitedb.New(":memory:", gormbase.WithSingularTable())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	gormDB := db.Unwrap().(*gorm.DB)
	hasTable := gormDB.Migrator().HasTable("test_user")
	if !hasTable {
		t.Error("expected singular table name 'test_user'")
	}

	hasPlural := gormDB.Migrator().HasTable("test_users")
	if hasPlural {
		t.Error("unexpected plural table name 'test_users' exists")
	}
}

// TestWithTablePrefix 测试表名前缀
func TestWithTablePrefix(t *testing.T) {
	db, err := sqlitedb.New(":memory:", gormbase.WithTablePrefix("app_"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	gormDB := db.Unwrap().(*gorm.DB)
	hasTable := gormDB.Migrator().HasTable("app_test_users")
	if !hasTable {
		t.Error("expected table with prefix 'app_test_users'")
	}

	hasUnprefixed := gormDB.Migrator().HasTable("test_users")
	if hasUnprefixed {
		t.Error("unexpected unprefixed table 'test_users' exists")
	}
}

// TestMustNew_Success 测试 MustNew 成功场景
func TestMustNew_Success(t *testing.T) {
	db := sqlitedb.MustNew(":memory:")
	if db == nil {
		t.Fatal("MustNew() returned nil")
	}
	defer db.Close()
}

// TestMustNew_Panic 测试 MustNew 在无效路径时 panic
func TestMustNew_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNew should panic on invalid file path")
		}
	}()
	sqlitedb.MustNew("/invalid/path/that/does/not/exist/db.sqlite")
}
