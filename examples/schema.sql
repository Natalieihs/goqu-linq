-- Database setup for Goqu-LINQ examples

-- Create database
CREATE DATABASE IF NOT EXISTS testdb CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE testdb;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL,
    age INT NOT NULL,
    status TINYINT NOT NULL DEFAULT 1 COMMENT '0=inactive, 1=active, 2=suspended',
    created_at BIGINT NOT NULL,
    updated_at BIGINT DEFAULT NULL,
    INDEX idx_username (username),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User accounts';

-- Orders table (for relationship examples)
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    order_no VARCHAR(50) NOT NULL UNIQUE,
    amount DECIMAL(10,2) NOT NULL,
    status TINYINT NOT NULL DEFAULT 0 COMMENT '0=pending, 1=completed, 2=cancelled',
    created_at BIGINT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_order_no (order_no),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User orders';

-- Insert sample data
INSERT INTO users (username, email, age, status, created_at) VALUES
('john_doe', 'john@example.com', 25, 1, UNIX_TIMESTAMP()),
('jane_smith', 'jane@example.com', 30, 1, UNIX_TIMESTAMP()),
('bob_wilson', 'bob@example.com', 22, 1, UNIX_TIMESTAMP()),
('alice_brown', 'alice@example.com', 28, 1, UNIX_TIMESTAMP()),
('charlie_davis', 'charlie@example.com', 35, 0, UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE email=VALUES(email);

INSERT INTO orders (user_id, order_no, amount, status, created_at)
SELECT
    u.id,
    CONCAT('ORD', LPAD(ROW_NUMBER() OVER (ORDER BY u.id), 6, '0')),
    ROUND(RAND() * 1000 + 50, 2),
    FLOOR(RAND() * 3),
    UNIX_TIMESTAMP()
FROM users u
WHERE u.status = 1
LIMIT 10
ON DUPLICATE KEY UPDATE amount=VALUES(amount);

-- Show tables
SHOW TABLES;

-- Show sample data
SELECT * FROM users LIMIT 5;
SELECT * FROM orders LIMIT 5;
