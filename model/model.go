package model

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Model struct {
	CreatedAt int
	UpdatedAt int
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func NewDBEngine() (*gorm.DB, error) {
	s := "%s:%s@tcp(%s)/%s?charset=%s&parseTime=%t&loc=Local"
	dsn := fmt.Sprintf(s, "root", "api@123456789", "127.0.0.1:3306", "api_service", "utf8mb4", true)
	opts := gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}}
	db, err := gorm.Open(mysql.Open(dsn), &opts)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Illust{}, &Tag{})
	return db, nil
}
