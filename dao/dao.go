package dao

import (
	"project/model"

	"dgo.baisic.print/GIN/model"
	"gorm.io/gorm"
)

// 全局db，外部注入
var db *gorm.DB

func InitDao(database *gorm.DB) {
	db = database
}

// 根据用户名查询用户
func GetBookById(ID string) (*model.Book, error) {
	var b model.Book
	err := db.Where("ID = ?", ID).First(&b).Error
	return &b, err
}

// 创建用户
func createBook(book *model.Book) error {
	return db.Create(book).Error
}
