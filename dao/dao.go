package dao

import (
	"dgo.baisic.print/GIN/model"

	"gorm.io/gorm"
)

// 全局db，外部注入
var db *gorm.DB

func InitDao(database *gorm.DB) {
	db = database
}

// 根据ID查询图书
func GetBookById(ID string) (*model.Book, error) {
	var b model.Book
	err := db.Where("id = ?", ID).First(&b).Error
	return &b, err
}

// 查询全部图书
func GetAllBooks() ([]model.Book, error) {
	var books []model.Book
	err := db.Find(&books).Error
	return books, err
}

// 新增图书
func CreateBook(book *model.Book) error {
	return db.Create(book).Error
}

// 更新图书
func UpdateBook(book *model.Book) error {
	return db.Save(book).Error
}

// 删除图书
func DeleteBook(ID string) error {
	return db.Where("id = ?", ID).Delete(&model.Book{}).Error
}
