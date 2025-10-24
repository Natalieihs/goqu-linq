# Goqu-LINQ Project Information

## Project Overview

**Goqu-LINQ** is an open-source, type-safe ORM framework for Go that provides LINQ-style query capabilities. It was extracted from a production financial services platform handling millions of transactions daily.

## Origin

This project was extracted from the [bsimerchant](../bsimerchant) codebase's `contrib/dao` module, which has been battle-tested in production for:
- Multi-country gaming/gambling platform
- 90+ payment integrations
- Multi-tenant architecture
- High-volume transaction processing

## Key Statistics

### Source Codebase
- **Total Files**: 483 Go files in dao module
- **Repositories**: 190 table-specific repositories
- **Code Volume**: ~20,000 lines of core ORM code
- **Production Usage**: Yes, in production since 2023

### Extracted Project
- **Core Files**: 7 main files
- **Lines of Code**: ~3,000 core framework
- **Test Coverage**: Basic structure tests included
- **Dependencies**: 4 main dependencies

## Architecture

### Core Components

1. **interfaces.go** (187 lines)
   - `IRepository[T]` - Main repository interface
   - `IQueryable[T]` - LINQ-style query interface
   - `IGroupingQuery[T]` - Grouping and aggregation
   - `IEnumerable[T]` - In-memory operations

2. **repository.go** (863 lines)
   - Generic `Repository[T]` implementation
   - CRUD operations
   - Batch insert/update with optimization
   - Unit of Work pattern
   - Transaction management

3. **queryable.go** (~800 lines)
   - Query building and execution
   - Chainable API
   - Type conversions
   - SQL generation

4. **enumerable.go** (~300 lines)
   - In-memory LINQ operations
   - Lambda expressions
   - Sorting, filtering, grouping

5. **group.go** (~200 lines)
   - SQL GROUP BY support
   - Aggregate builders
   - Having clauses

6. **db.go** (120 lines)
   - Database connection wrapper
   - Query logging
   - Slow query detection

## Technical Stack

### Core Dependencies
```
github.com/doug-martin/goqu/v9 v9.19.0    - SQL builder
github.com/jmoiron/sqlx v1.4.0            - SQL extensions
github.com/go-sql-driver/mysql v1.9.2     - MySQL driver
go.uber.org/zap v1.27.0                   - Structured logging
```

### Supported Databases
- MySQL 5.7+
- MariaDB 10.3+
- PostgreSQL 12+ (via goqu)
- SQLite (via goqu)
- StarRocks (analytics)

## Design Patterns

### 1. Repository Pattern
```
Application Layer
       ↓
UserRepository (specific)
       ↓
Repository[T] (generic)
       ↓
Database
```

### 2. Unit of Work
```go
uow := NewUnitOfWork(db)
uow.RunInTransaction(func(uow IUnitOfWork) error {
    // All operations in same transaction
    // Auto-commit on success
    // Auto-rollback on error or panic
})
```

### 3. Query Builder (Fluent API)
```go
Query().
    Where(conditions).
    OrderBy(columns).
    Skip(offset).
    Take(limit).
    ToList()
```

## Performance Optimizations

### 1. Batch Operations
- Automatic splitting to prevent parameter overflow
- MySQL's `max_prepared_stmt_count` limit handling
- Calculates safe batch size: `(maxParams * 0.8) / fieldCount`

### 2. Query Optimization
- Explicit column selection
- Prepared statements
- Connection pooling
- Slow query detection (>5s threshold)

### 3. Memory Efficiency
- Streaming for large datasets
- Pagination support
- Efficient batch processing

## Comparison with Other ORMs

| Feature | Goqu-LINQ | GORM | ENT | sqlx |
|---------|-----------|------|-----|------|
| Type Safety | ✅ Generics | ❌ Reflection | ✅ Code-gen | ⚠️ Manual |
| LINQ Style | ✅ Full | ❌ | ⚠️ Partial | ❌ |
| Batch Ops | ✅ Optimized | ⚠️ Basic | ✅ Yes | ❌ Manual |
| Learning Curve | Medium | Low | High | Low |
| Performance | High | Medium | High | Highest |
| Magic | Low | High | Medium | None |

## Usage Patterns

### Simple Query
```go
users, err := userRepo.Query().
    Where(goqu.Ex{"status": 1}).
    ToList()
```

### Complex Query
```go
users, err := userRepo.Query().
    Where(goqu.Ex{
        "age": goqu.Op{"between": goqu.Range(18, 65)},
        "status": goqu.Op{"in": []int{1, 2}},
    }).
    OrderByRaw("created_at DESC").
    Skip(20).
    Take(10).
    ToList()
```

### Transaction
```go
uow := NewUnitOfWork(db)
err := uow.RunInTransaction(func(uow IUnitOfWork) error {
    repo := userRepo.WithUnitOfWork(uow)
    // operations...
    return nil
})
```

### Batch Insert
```go
users := []*User{...}
err := userRepo.BatchInsert(users, &BatchInsertOption{
    BatchSize: 1000,
})
```

## Roadmap

### Phase 1: Core (Completed)
- [x] Extract from source project
- [x] Clean up dependencies
- [x] Create documentation
- [x] Basic tests
- [x] Examples

### Phase 2: Enhancement (Future)
- [ ] More comprehensive tests
- [ ] Benchmark suite
- [ ] Additional database support
- [ ] Migration tools
- [ ] Code generation tools

### Phase 3: Community (Future)
- [ ] Gather feedback
- [ ] Performance optimizations
- [ ] Plugin system
- [ ] Extended documentation

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md)

Areas we need help:
- Test coverage
- Documentation
- Examples
- Bug reports
- Performance testing

## License

MIT License - See [LICENSE](LICENSE)

## Credits

### Original Development
- Extracted from production financial platform
- Developed and tested under high-load conditions
- Battle-tested with millions of daily transactions

### Core Team
- Original implementation: Production team at bsimerchant
- Open source extraction: 2024

### Dependencies
- [goqu](https://github.com/doug-martin/goqu) - Excellent SQL builder
- [sqlx](https://github.com/jmoiron/sqlx) - SQL extensions
- [zap](https://github.com/uber-go/zap) - Fast logging
- Inspired by [.NET LINQ](https://docs.microsoft.com/en-us/dotnet/csharp/linq/)

## Contact & Support

- **Issues**: [GitHub Issues](https://github.com/Natalieihs/goqu-linq/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Natalieihs/goqu-linq/discussions)
- **Email**: (Add your email)

## Acknowledgments

Special thanks to:
- The bsimerchant team for the original implementation
- The Go community for excellent tooling
- All contributors and early adopters

---

**Status**: Ready for production use (extracted from production code)

**Version**: 1.0.0 (Initial release)

**Last Updated**: 2024-10-24
