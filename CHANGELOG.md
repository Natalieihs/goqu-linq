# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- GitHub Actions workflow for automated testing
- Bug report issue template
- CHANGELOG.md for tracking version history

### Changed
- Upgraded to Go 1.23
- Updated dependencies to latest versions
  - github.com/go-sql-driver/mysql v1.9.2 → v1.9.3

### Fixed
- Removed debug print statements from production code
- Cleaned up commented-out code blocks in queryable.go
- Applied go fmt to all source files

## [1.0.0] - 2024-01-XX

### Added
- Initial release of Goqu-LINQ
- Type-safe generic Repository pattern
- LINQ-style query builder (Queryable)
- In-memory LINQ operations (Enumerable)
- Unit of Work pattern with transaction support
- Batch insert and update operations with parameter limit handling
- Support for MySQL, PostgreSQL, StarRocks, and SQLite
- Query logging with Zap integration
- Slow query detection
- Aggregation and grouping support (Sum, Max, Min, Average, Count)
- Comprehensive examples and documentation
- MIT License

### Features

#### Repository Pattern
- Generic `Repository[T]` with full CRUD operations
- Transaction support via Unit of Work
- Batch operations with automatic parameter count management
- Support for custom queries and raw SQL

#### Query Builder
- Chainable LINQ-style API
- Where conditions with complex operators
- Ordering (single and multiple columns)
- Pagination (Skip/Take)
- First, FirstOrDefault, Single, Count, Any operations
- Join support
- Aggregations

#### In-Memory Operations
- Where, Select, OrderBy, OrderByDescending
- Skip, Take, First, Last
- GroupBy with custom selectors
- ToList, ToMap, ToSet conversions

#### Database Features
- Connection pooling
- Query logging
- Slow query warnings
- Multiple database dialect support
- Transaction management
- Batch operation optimization

### Dependencies
- github.com/doug-martin/goqu/v9 v9.19.0
- github.com/jmoiron/sqlx v1.4.0
- github.com/go-sql-driver/mysql v1.9.3
- go.uber.org/zap v1.27.0

---

## Version History Notes

### Versioning Strategy

This project uses [Semantic Versioning](https://semver.org/):
- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

### Upgrade Guide

When upgrading between versions, please check the specific version sections above for:
- **Breaking Changes**: Changes that may require code modifications
- **Deprecated Features**: Features that will be removed in future versions
- **Migration Steps**: Specific actions needed when upgrading

### Categories

We use these categories to organize changes:
- **Added**: New features
- **Changed**: Changes to existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Security vulnerability fixes

---

[Unreleased]: https://github.com/Natalieihs/goqu-linq/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/Natalieihs/goqu-linq/releases/tag/v1.0.0
