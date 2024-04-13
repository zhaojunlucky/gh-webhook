package model

import "gorm.io/gorm"

// GitHub server configuration
type GitHub struct {
	gorm.Model
	Web  string
	API  string `gorm:"uniqueIndex"`
	Name string `gorm:"uniqueIndex"`
}
