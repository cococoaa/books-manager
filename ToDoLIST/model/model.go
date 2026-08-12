package model

import "gorm.io/gorm"

type Todo struct {
	gorm.Model
	Title  string `gorm:"not null;comment:'待办事情内容'" json:"title"`
	Status bool   `gorm:"not null;default:false;comment:'是否完成'" json:"status"`
}
