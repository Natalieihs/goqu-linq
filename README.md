# Goqu-LINQ

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

**Goqu-LINQ** is a powerful, type-safe ORM framework for Go that brings LINQ-style query capabilities to your database operations. Built on top of [goqu](https://github.com/doug-martin/goqu), it provides a fluent, chainable API with full generic support.

## ✨ Features

- 🔷 **Type-Safe Generics**: Full Go 1.18+ generic support for compile-time type safety
- 🔗 **LINQ-Style API**: Chainable, fluent query interface inspired by .NET LINQ
- 📦 **Repository Pattern**: Clean separation of data access logic
- 🔄 **Unit of Work**: Built-in transaction management with automatic rollback
- ⚡ **Batch Operations**: Optimized bulk insert/update with automatic batching
- 🔍 **Rich Query Capabilities**: Complex filters, joins, aggregations, and grouping
- 📊 **In-Memory Operations**: LINQ-style enumerable for in-memory collections
- 🎯 **No Magic**: Clear, explicit code generation-free approach
- 📝 **SQL Logging**: Built-in query logging with zap integration
- 🌐 **Multi-Database**: Support for MySQL, PostgreSQL, StarRocks, and more

## 📦 Installation

```bash
go get github.com/Natalieihs/goqu-linq
```

## 🚀 Quick Start

### 1. Define Your Entity

```go
type User struct {
    ID        int64  `db:"id" json:"id"`
    Username  string `db:"username" json:"username"`
    Email     string `db:"email" json:"email"`
    Age       int    `db:"age" json:"age"`
    Status    int    `db:"status" json:"status"`
    CreatedAt int64  `db:"created_at" json:"created_at"`
}

func (u *User) TableName() string {
    return "users"
}
```

### 2. Create Repository

```go
import (
    "github.com/Natalieihs/goqu-linq/core"
    "go.uber.org/zap"
)

type UserRepository struct {
    *core.Repository[User]
}

func NewUserRepository(db *core.DBLogger) *UserRepository {
    return &UserRepository{
        Repository: core.NewRepository[User](db, "users", core.MySQL),
    }
}
```

### 3. Connect and Use

```go
// Initialize logger
logger, _ := zap.NewDevelopment()

// Connect to database
dsn := "root:password@tcp(localhost:3306)/testdb?parseTime=true"
db, err := core.ConnectMySQL(dsn, logger, "app")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Create repository
userRepo := NewUserRepository(db)

// Query users
users, err := userRepo.Query().
    Where(goqu.Ex{"age": goqu.Op{"gte": 18}}).
    OrderByRaw("created_at DESC").
    Limit(10).
    ToList()
```

## 📖 Usage Examples

### Basic CRUD Operations

#### Create
```go
newUser := &User{
    Username:  "john_doe",
    Email:     "john@example.com",
    Age:       25,
    Status:    1,
}
err := userRepo.Create(newUser)
```

#### Read
```go
// Single record
user, err := userRepo.Query().
    Where(goqu.Ex{"username": "john_doe"}).
    FirstOrDefault()

// Multiple records
users, err := userRepo.Query().
    Where(goqu.Ex{"status": 1}).
    OrderByRaw("age DESC").
    ToList()

// Count
count, err := userRepo.Query().
    Where(goqu.Ex{"status": 1}).
    Count()
```

#### Update
```go
// Update specific fields
err := userRepo.UpdateFieldsByCondition(
    goqu.Ex{"username": "john_doe"},
    map[string]interface{}{
        "age":   26,
        "email": "newemail@example.com",
    },
)

// Update entire entity
user.Age = 27
err := userRepo.Update(user)
```

#### Delete
```go
err := userRepo.BatchDelete(goqu.Ex{"status": 0})
```

### Advanced Queries

#### Complex Filters
```go
users, err := userRepo.Query().
    Where(goqu.Ex{
        "age":    goqu.Op{"between": goqu.Range(18, 65)},
        "status": goqu.Op{"in": []int{1, 2}},
        "email":  goqu.Op{"like": "%@example.com"},
    }).
    OrderByRaw("created_at DESC").
    Skip(20).
    Take(10).
    ToList()
```

#### Aggregations
```go
// Sum
totalAge, err := userRepo.Query().
    Where(goqu.Ex{"status": 1}).
    Sum("age")

// Max
maxAge, err := userRepo.Query().Max("age")

// Min
minAge, err := userRepo.Query().Min("age")
```

#### Raw SQL
```go
users, err := userRepo.Query().
    WhereRaw("age > ? AND created_at > ?", 18, time.Now().Unix()).
    ToList()
```

### Batch Operations

#### Batch Insert
```go
users := []*User{
    {Username: "alice", Email: "alice@example.com", Age: 22},
    {Username: "bob", Email: "bob@example.com", Age: 28},
    {Username: "charlie", Email: "charlie@example.com", Age: 35},
}

// Automatic batching to prevent parameter overflow
err := userRepo.BatchInsert(users, &core.BatchInsertOption{
    BatchSize: 1000,
})
```

#### Batch Update
```go
updates := []*User{
    {ID: 1, Status: 2, Age: 26},
    {ID: 2, Status: 2, Age: 29},
    {ID: 3, Status: 2, Age: 36},
}

err := userRepo.BatchUpdate(updates, &core.BatchUpdateOption{
    KeyField:     "id",
    UpdateFields: []string{"status", "age"},
})
```

### Transaction (Unit of Work)

```go
uow := core.NewUnitOfWork(db)
err := uow.RunInTransaction(func(uow core.IUnitOfWork) error {
    repoWithTx := userRepo.WithUnitOfWork(uow)

    // Create user
    if err := repoWithTx.Create(newUser); err != nil {
        return err // Automatic rollback
    }

    // Update balance
    if err := repoWithTx.UpdateFieldsByCondition(
        goqu.Ex{"id": newUser.ID},
        map[string]interface{}{
            "balance": goqu.L("balance + ?", 100),
        },
    ); err != nil {
        return err // Automatic rollback
    }

    return nil // Commit
})
```

### In-Memory Operations (Enumerable)

```go
// Load data from database
users, _ := userRepo.Query().ToList()

// Convert to enumerable for in-memory operations
enum := core.NewEnumerable(users)

// LINQ-style operations
adults := enum.
    Where(func(u User) bool { return u.Age >= 18 }).
    OrderBy(func(a, b User) bool { return a.Age < b.Age }).
    Take(10).
    ToList()

// Aggregations
avgAge := enum.Average(func(u User) float64 { return float64(u.Age) })
totalAge := enum.Sum(func(u User) float64 { return float64(u.Age) })

// Grouping
grouped := enum.GroupBy(func(u User) interface{} { return u.Status })
```

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     Application Layer                    │
│                   (Controllers/Services)                 │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│                   Repository Layer                       │
│            (UserRepository, OrderRepository)             │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│                Core Repository[T]                        │
│          (Generic CRUD + Query Operations)               │
└──────────────────────┬──────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        ▼              ▼               ▼
  ┌──────────┐  ┌──────────┐  ┌──────────────┐
  │IQueryable│  │Enumerable│  │Unit of Work  │
  └────┬─────┘  └──────────┘  └──────────────┘
       │
       ▼
  ┌──────────┐
  │  goqu    │
  └────┬─────┘
       │
       ▼
  ┌──────────┐
  │   sqlx   │
  └────┬─────┘
       │
       ▼
  ┌──────────┐
  │  MySQL   │
  └──────────┘
```

## 🔧 Core Interfaces

### IRepository[T]
```go
type IRepository[T any] interface {
    IReadRepository[T]
    IWriteRepository[T]
}

type IReadRepository[T any] interface {
    Query() IQueryable[T]
}

type IWriteRepository[T any] interface {
    Create(entity *T) error
    Update(entity *T) error
    UpdateByCondition(condition goqu.Ex, entity *T) error
    UpdateFieldsByCondition(condition goqu.Ex, fields map[string]interface{}) error
    BatchCreate(entities []*T) error
    BatchDelete(condition goqu.Ex) error
    BatchInsert(entities []*T, opt *BatchInsertOption) error
    BatchUpdate(entities []*T, opt *BatchUpdateOption) error
}
```

### IQueryable[T]
```go
type IQueryable[T any] interface {
    // Filtering
    Where(condition goqu.Ex) IQueryable[T]
    WhereRaw(condition string, args ...interface{}) IQueryable[T]

    // Ordering
    OrderBy(cols ...string) IQueryable[T]
    OrderByRaw(order string) IQueryable[T]

    // Pagination
    Skip(offset int) IQueryable[T]
    Take(limit int) IQueryable[T]
    Limit(limit int) IQueryable[T]

    // Selection
    Select(cols ...interface{}) IQueryable[T]
    SelectRaw(cols ...string) IQueryable[T]

    // Grouping
    GroupBy(keySelector func(T) interface{}) IGroupingQuery[T]
    GroupByColumns(cols ...string) IQueryable[T]

    // Execution
    FirstOrDefault() (*T, error)
    ToList() ([]*T, error)
    Count() (int64, error)

    // Aggregations
    Sum(field string) (float64, error)
    Max(field string) (interface{}, error)
    Min(field string) (interface{}, error)

    // Pagination
    ToPagedList(page, size int, condition goqu.Ex) (*PageResult[T], error)

    // Utilities
    ToSQL() (sql string, params []interface{}, err error)
}
```

## 🎯 Design Principles

1. **Type Safety**: Leveraging Go generics for compile-time type checking
2. **Explicit over Implicit**: No hidden magic, clear intentions
3. **Separation of Concerns**: Clean architecture with repository pattern
4. **Testability**: Easy to mock and test
5. **Performance**: Optimized batch operations and query generation
6. **Flexibility**: Extensible for custom business logic

## 🔍 Comparison with Other ORMs

| Feature | Goqu-LINQ | GORM | ENT |
|---------|-----------|------|-----|
| **Generics Support** | ✅ Full | ❌ No | ✅ Code-gen |
| **Type Safety** | ✅ Compile-time | ⚠️ Runtime | ✅ Compile-time |
| **LINQ Style** | ✅ Full | ❌ No | ⚠️ Partial |
| **Batch Optimization** | ✅ Auto | ⚠️ Basic | ✅ Yes |
| **Unit of Work** | ✅ Built-in | ⚠️ Manual | ✅ Built-in |
| **Learning Curve** | ⚠️ Medium | ✅ Low | ⚠️ High |
| **Query Flexibility** | ✅ High | ✅ High | ⚠️ Medium |
| **Raw SQL** | ✅ Easy | ✅ Easy | ⚠️ Limited |

## 📊 Performance Tips

1. **Use Batch Operations**: For bulk inserts/updates, use `BatchInsert`/`BatchUpdate`
2. **Limit Query Fields**: Use `Select()` to fetch only needed columns
3. **Use Transactions**: Group related operations in `UnitOfWork`
4. **Connection Pooling**: Configure `SetMaxOpenConns` and `SetMaxIdleConns`
5. **Index Your Tables**: Ensure proper database indexing for query conditions

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built on top of [goqu](https://github.com/doug-martin/goqu) - Excellent SQL builder
- Inspired by [LINQ](https://docs.microsoft.com/en-us/dotnet/csharp/programming-guide/concepts/linq/) - .NET's LINQ
- Uses [sqlx](https://github.com/jmoiron/sqlx) - Extensions to database/sql
- Logging via [zap](https://github.com/uber-go/zap) - Fast, structured logging

## 📚 Documentation

For more detailed documentation, see:
- [API Reference](docs/API.md)
- [Examples](examples/)
- [Best Practices](docs/BEST_PRACTICES.md)

## 💬 Support

- Create an issue for bug reports or feature requests
- Star ⭐ the repository if you find it useful!
- Follow for updates

---

**Made with ❤️ for the Go community**
