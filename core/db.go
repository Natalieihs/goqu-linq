package core

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// DBLogger wraps sqlx.DB with logging capabilities
type DBLogger struct {
	*sqlx.DB
	logger *zap.Logger
	prefix string
}

// Tx wraps sqlx.Tx for transaction operations
type Tx = sqlx.Tx

// NewDBLogger creates a new DBLogger instance
func NewDBLogger(db *sqlx.DB, logger *zap.Logger, prefix string) *DBLogger {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}
	return &DBLogger{
		DB:     db,
		logger: logger,
		prefix: prefix,
	}
}

// ConnectMySQL connects to MySQL database
func ConnectMySQL(dsn string, logger *zap.Logger, prefix string) (*DBLogger, error) {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	return NewDBLogger(db, logger, prefix), nil
}

// GetPrefix returns the database prefix
func (db *DBLogger) GetPrefix() string {
	return db.prefix
}

// Begin starts a transaction
func (db *DBLogger) Begin() (*Tx, error) {
	return db.DB.Beginx()
}

// ExecContext executes a query with context
func (db *DBLogger) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := db.DB.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	db.logQuery(ctx, "Exec", query, args, err, duration)
	return result, err
}

// QueryContext queries with context
func (db *DBLogger) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := db.DB.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	db.logQuery(ctx, "Query", query, args, err, duration)
	return rows, err
}

// logQuery logs database operations
func (db *DBLogger) logQuery(ctx context.Context, operation, query string, args []interface{}, err error, duration time.Duration) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("query", query),
		zap.Any("args", args),
		zap.Duration("duration", duration),
		zap.String("prefix", db.prefix),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		db.logger.Error("Database operation failed", fields...)
	} else if duration > 5*time.Second {
		db.logger.Warn("Slow query detected", fields...)
	} else {
		db.logger.Debug("Database operation", fields...)
	}
}
