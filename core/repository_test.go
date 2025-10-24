package core

import (
	"testing"

	"github.com/doug-martin/goqu/v9"
)

// TestEntity for testing purposes
type TestEntity struct {
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	Status int    `db:"status"`
}

func TestNewRepository(t *testing.T) {
	// This is a basic structure test
	// In a real scenario, you'd need a test database

	// Test that NewRepository doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewRepository panicked: %v", r)
		}
	}()

	// Create a nil db for structure testing only
	// In real tests, use a test database
	var db *DBLogger
	repo := NewRepository[TestEntity](db, "test_table", MySQL)

	if repo == nil {
		t.Error("Expected non-nil repository")
	}

	if repo.table != "test_table" {
		t.Errorf("Expected table name 'test_table', got '%s'", repo.table)
	}

	if repo.dbType != MySQL {
		t.Errorf("Expected dbType MySQL, got %v", repo.dbType)
	}
}

func TestBatchInsertOption(t *testing.T) {
	opt := DefaultBatchInsertOption

	if opt.BatchSize != 1000 {
		t.Errorf("Expected default batch size 1000, got %d", opt.BatchSize)
	}

	if opt.UseNamedExec != false {
		t.Errorf("Expected default UseNamedExec false, got %v", opt.UseNamedExec)
	}
}

func TestCalculateSafeBatchSize(t *testing.T) {
	tests := []struct {
		fieldCount int
		maxParams  int
		expected   int
	}{
		{10, 16384, 1310}, // (16384 * 0.8) / 10
		{20, 16384, 655},  // (16384 * 0.8) / 20
		{5, 10000, 1600},  // (10000 * 0.8) / 5
	}

	for _, tt := range tests {
		result := calculateSafeBatchSize(tt.fieldCount, tt.maxParams)
		if result != tt.expected {
			t.Errorf("calculateSafeBatchSize(%d, %d) = %d, expected %d",
				tt.fieldCount, tt.maxParams, result, tt.expected)
		}
	}
}

func TestDialectType(t *testing.T) {
	if MySQL != "mysql" {
		t.Errorf("Expected MySQL dialect to be 'mysql', got '%s'", MySQL)
	}

	if StarRocks != "starrocks" {
		t.Errorf("Expected StarRocks dialect to be 'starrocks', got '%s'", StarRocks)
	}
}

func TestPageResult(t *testing.T) {
	result := PageResult[TestEntity]{
		Items:    []*TestEntity{{ID: 1, Name: "test"}},
		Total:    100,
		Page:     1,
		PageSize: 10,
	}

	if len(result.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(result.Items))
	}

	if result.Total != 100 {
		t.Errorf("Expected total 100, got %d", result.Total)
	}

	if result.Page != 1 {
		t.Errorf("Expected page 1, got %d", result.Page)
	}

	if result.PageSize != 10 {
		t.Errorf("Expected page size 10, got %d", result.PageSize)
	}
}

// Example of how to use the repository (documentation)
func ExampleRepository_Query() {
	// This example shows the basic usage pattern
	// In a real scenario, initialize db properly

	var db *DBLogger // Initialize from ConnectMySQL
	repo := NewRepository[TestEntity](db, "test_entities", MySQL)

	// Query with conditions
	_ = repo.Query().
		Where(goqu.Ex{"status": 1}).
		OrderByRaw("created_at DESC").
		Limit(10)

	// Output: (example only, won't actually run)
}
