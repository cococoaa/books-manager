package model

import "gorm.io/gorm"

type Book struct {
	gorm.Model
	ID     string `gorm:"not null;comment:'图书ID'"`
	Title  string `gorm:"not null;comment:'图书名称'`
	Author string `gorm:"not null;comment:'作者'"`
	Year   int    `gorm:"not null;comment:'图书出版日期'"`
}
