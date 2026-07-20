package main

import (
	"time"
)

// User 用户模型
type User struct {
	ID        uint      `xorm:"'id' pk autoincr" json:"id"`
	Name      string    `xorm:"varchar(100) notnull" json:"name"`
	Email     string    `xorm:"varchar(100) unique notnull" json:"email"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
