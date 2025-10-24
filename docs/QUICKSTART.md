# Quick Start Guide

This guide will help you get started with Goqu-LINQ in 5 minutes.

## Prerequisites

- Go 1.21 or higher
- MySQL 5.7+ or compatible database
- Basic understanding of Go and SQL

## Installation

```bash
go get github.com/Natalieihs/goqu-linq
```

## Step 1: Create Database and Table

```sql
CREATE DATABASE testdb CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE testdb;

CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL,
    age INT NOT NULL,
    status TINYINT NOT NULL DEFAULT 1,
    created_at BIGINT NOT NULL,
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## Step 2: Define Your Entity

```go
package main

import (
    "github.com/Natalieihs/goqu-linq/core"
)

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

## Step 3: Create Repository

```go
type UserRepository struct {
    *core.Repository[User]
}

func NewUserRepository(db *core.DBLogger) *UserRepository {
    return &UserRepository{
        Repository: core.NewRepository[User](db, "users", core.MySQL),
    }
}
```

## Step 4: Connect and Use

```go
package main

import (
    "log"
    "time"

    "github.com/doug-martin/goqu/v9"
    _ "github.com/go-sql-driver/mysql"
    "go.uber.org/zap"
)

func main() {
    // 1. Initialize logger
    logger, err := zap.NewDevelopment()
    if err != nil {
        log.Fatal(err)
    }
    defer logger.Sync()

    // 2. Connect to database
    dsn := "root:password@tcp(localhost:3306)/testdb?parseTime=true&charset=utf8mb4"
    db, err := core.ConnectMySQL(dsn, logger, "app")
    if err != nil {
        log.Fatal("Failed to connect:", err)
    }
    defer db.Close()

    // 3. Create repository
    userRepo := NewUserRepository(db)

    // 4. Create a user
    newUser := &User{
        Username:  "john_doe",
        Email:     "john@example.com",
        Age:       25,
        Status:    1,
        CreatedAt: time.Now().Unix(),
    }

    if err := userRepo.Create(newUser); err != nil {
        log.Printf("Failed to create user: %v", err)
    } else {
        log.Println("✓ User created successfully")
    }

    // 5. Query users
    users, err := userRepo.Query().
        Where(goqu.Ex{"status": 1}).
        OrderByRaw("created_at DESC").
        Limit(10).
        ToList()

    if err != nil {
        log.Printf("Failed to query: %v", err)
    } else {
        log.Printf("✓ Found %d users", len(users))
        for _, user := range users {
            log.Printf("  - %s (%s)", user.Username, user.Email)
        }
    }

    // 6. Update user
    err = userRepo.UpdateFieldsByCondition(
        goqu.Ex{"username": "john_doe"},
        map[string]interface{}{
            "age": 26,
        },
    )
    if err != nil {
        log.Printf("Failed to update: %v", err)
    } else {
        log.Println("✓ User updated")
    }

    // 7. Count users
    count, err := userRepo.Query().
        Where(goqu.Ex{"status": 1}).
        Count()

    if err != nil {
        log.Printf("Failed to count: %v", err)
    } else {
        log.Printf("✓ Total active users: %d", count)
    }

    log.Println("🎉 Quick start completed!")
}
```

## Step 5: Run

```bash
go mod init myapp
go get github.com/Natalieihs/goqu-linq
go get github.com/doug-martin/goqu/v9
go get github.com/go-sql-driver/mysql
go get go.uber.org/zap
go run main.go
```

## Common Operations

### Query with Multiple Conditions

```go
users, err := userRepo.Query().
    Where(goqu.Ex{
        "age":    goqu.Op{"gte": 18},
        "status": goqu.Op{"in": []int{1, 2}},
    }).
    ToList()
```

### Pagination

```go
page := 1
pageSize := 20
offset := (page - 1) * pageSize

users, err := userRepo.Query().
    Where(goqu.Ex{"status": 1}).
    OrderByRaw("created_at DESC").
    Skip(offset).
    Take(pageSize).
    ToList()
```

### Aggregate Functions

```go
// Sum
totalAge, _ := userRepo.Query().Sum("age")

// Max
maxAge, _ := userRepo.Query().Max("age")

// Min
minAge, _ := userRepo.Query().Min("age")
```

### Batch Operations

```go
users := []*User{
    {Username: "user1", Email: "user1@example.com", Age: 20, Status: 1},
    {Username: "user2", Email: "user2@example.com", Age: 25, Status: 1},
}

err := userRepo.BatchInsert(users, nil)
```

### Transactions

```go
uow := core.NewUnitOfWork(db)
err := uow.RunInTransaction(func(uow core.IUnitOfWork) error {
    repo := userRepo.WithUnitOfWork(uow)

    // All operations here are in the same transaction
    if err := repo.Create(newUser); err != nil {
        return err // Automatic rollback
    }

    if err := repo.UpdateFieldsByCondition(
        goqu.Ex{"id": newUser.ID},
        map[string]interface{}{"status": 2},
    ); err != nil {
        return err // Automatic rollback
    }

    return nil // Commit
})
```

## Next Steps

- Read the [Full Documentation](../README.md)
- Check out [More Examples](../examples/)
- Learn about [Best Practices](BEST_PRACTICES.md)

## Troubleshooting

### Connection Issues

Make sure:
1. MySQL is running: `systemctl status mysql`
2. User has proper permissions: `GRANT ALL ON testdb.* TO 'user'@'localhost'`
3. Firewall allows connections
4. DSN format is correct

### Common Errors

**"Error 1045: Access denied"**
- Check username and password in DSN

**"Error 1049: Unknown database"**
- Create the database first

**"dial tcp: connection refused"**
- Check if MySQL is running
- Verify host and port

## Support

- Check [Issues](https://github.com/Natalieihs/goqu-linq/issues)
- Read [API Documentation](API.md)
- See [Examples](../examples/)
