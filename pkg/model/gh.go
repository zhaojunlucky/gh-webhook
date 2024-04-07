package model

import "gorm.io/gorm"

// GitHub server configuration
type GitHub struct {
	gorm.Model
	Web  string
	API  string
	Name string
}
