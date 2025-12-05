package database

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/dspo/go-homework/internal/model"
)

func NewGorm() *gorm.DB {
	// DSN 格式: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	dsn := "root:changeme@tcp(mysql:3306)/go_dev?charset=utf8mb4&parseTime=True&loc=Local"
	orm, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}

	if err = orm.Migrator().AutoMigrate(
		new(model.Audit),
		new(model.Project),
		new(model.Role),
		new(model.Team),
		new(model.User),
		new(model.UserProject),
		new(model.UserRole),
		new(model.UserTeam),
	); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	return orm
}
