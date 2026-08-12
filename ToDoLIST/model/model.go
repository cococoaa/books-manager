package model

import "gorm.io/gorm"

type Todo struct {
	gorm.Model
	Content string `gorm:"not null;comment:'待办事情内容'"`
}
