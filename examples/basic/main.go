package main

import (
	"fmt"
	"log"

	"github.com/Natalieihs/goqu-linq/core"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

// User represents a user entity
type User struct {
	ID        int64  `db:"id" json:"id"`
	Username  string `db:"username" json:"username"`
	Email     string `db:"email" json:"email"`
	Age       int    `db:"age" json:"age"`
	Status    int    `db:"status" json:"status"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
}

// TableName returns the table name for User
func (u *User) TableName() string {
	return "users"
}

// UserRepository wraps the generic repository
type UserRepository struct {
	*core.Repository[User]
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *core.DBLogger) *UserRepository {
	return &UserRepository{
		Repository: core.NewRepository[User](db, "users", core.MySQL),
	}
}

// Custom methods for UserRepository
func (r *UserRepository) GetByUsername(username string) (*User, error) {
	return r.Query().
		Where(goqu.Ex{"username": username}).
		FirstOrDefault()
}

func (r *UserRepository) GetActiveUsers(page, pageSize int) ([]*User, error) {
	offset := (page - 1) * pageSize
	return r.Query().
		Where(goqu.Ex{"status": 1}).
		OrderByRaw("created_at DESC").
		Skip(offset).
		Take(pageSize).
		ToList()
}

func main() {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQL
	// DSN format: user:password@tcp(host:port)/database?parseTime=true
	dsn := "root:password@tcp(localhost:3306)/testdb?parseTime=true&charset=utf8mb4"
	db, err := core.ConnectMySQL(dsn, logger, "demo")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create repository
	userRepo := NewUserRepository(db)

	// Example 1: Create a new user
	newUser := &User{
		Username:  "john_doe",
		Email:     "john@example.com",
		Age:       25,
		Status:    1,
		CreatedAt: 1609459200, // Unix timestamp
	}
	if err := userRepo.Create(newUser); err != nil {
		log.Printf("Failed to create user: %v", err)
	} else {
		fmt.Println("✓ Created new user")
	}

	// Example 2: Query single user
	user, err := userRepo.GetByUsername("john_doe")
	if err != nil {
		log.Printf("Failed to get user: %v", err)
	} else {
		fmt.Printf("✓ Found user: %+v\n", user)
	}

	// Example 3: Query with conditions
	users, err := userRepo.Query().
		Where(goqu.Ex{
			"age":    goqu.Op{"gte": 18},
			"status": 1,
		}).
		OrderByRaw("age DESC").
		Limit(10).
		ToList()
	if err != nil {
		log.Printf("Failed to query users: %v", err)
	} else {
		fmt.Printf("✓ Found %d active adult users\n", len(users))
	}

	// Example 4: Count users
	count, err := userRepo.Query().
		Where(goqu.Ex{"status": 1}).
		Count()
	if err != nil {
		log.Printf("Failed to count users: %v", err)
	} else {
		fmt.Printf("✓ Total active users: %d\n", count)
	}

	// Example 5: Update user
	if err := userRepo.UpdateFieldsByCondition(
		goqu.Ex{"username": "john_doe"},
		map[string]interface{}{
			"age":   26,
			"email": "newemail@example.com",
		},
	); err != nil {
		log.Printf("Failed to update user: %v", err)
	} else {
		fmt.Println("✓ Updated user")
	}

	// Example 6: Batch insert
	newUsers := []*User{
		{Username: "alice", Email: "alice@example.com", Age: 22, Status: 1},
		{Username: "bob", Email: "bob@example.com", Age: 28, Status: 1},
		{Username: "charlie", Email: "charlie@example.com", Age: 35, Status: 1},
	}
	if err := userRepo.BatchInsert(newUsers, nil); err != nil {
		log.Printf("Failed to batch insert: %v", err)
	} else {
		fmt.Printf("✓ Batch inserted %d users\n", len(newUsers))
	}

	// Example 7: Aggregate query
	totalAge, err := userRepo.Query().
		Where(goqu.Ex{"status": 1}).
		Sum("age")
	if err != nil {
		log.Printf("Failed to sum ages: %v", err)
	} else {
		fmt.Printf("✓ Total age of active users: %.0f\n", totalAge)
	}

	// Example 8: Transaction (Unit of Work)
	uow := core.NewUnitOfWork(db)
	err = uow.RunInTransaction(func(uow core.IUnitOfWork) error {
		repoWithTx := userRepo.WithUnitOfWork(uow)

		// Create user in transaction
		txUser := &User{
			Username: "tx_user",
			Email:    "tx@example.com",
			Age:      30,
			Status:   1,
		}
		if err := repoWithTx.Create(txUser); err != nil {
			return err
		}

		// Update in same transaction
		return repoWithTx.UpdateFieldsByCondition(
			goqu.Ex{"username": "tx_user"},
			map[string]interface{}{"age": 31},
		)
	})
	if err != nil {
		log.Printf("Transaction failed: %v", err)
	} else {
		fmt.Println("✓ Transaction completed successfully")
	}

	// Example 9: Delete users
	if err := userRepo.BatchDelete(goqu.Ex{"status": 0}); err != nil {
		log.Printf("Failed to delete users: %v", err)
	} else {
		fmt.Println("✓ Deleted inactive users")
	}

	fmt.Println("\n🎉 All examples completed!")
}
