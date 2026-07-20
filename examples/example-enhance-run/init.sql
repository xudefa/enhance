-- 创建数据库
CREATE DATABASE IF NOT EXISTS enhance_demo DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE enhance_demo;

-- 注意：users 表会通过 GORM 的 AutoMigrate 自动创建
-- 如果需要手动创建，可以使用以下 SQL：
--
-- CREATE TABLE IF NOT EXISTS users (
--     id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
--     name VARCHAR(100) NOT NULL,
--     email VARCHAR(100) NOT NULL UNIQUE,
--     age INT,
--     created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
--     updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
--     INDEX idx_email (email)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;