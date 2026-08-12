package dao

import (
	"dgo.baisic.print/GIN/ToDoLIST/model"
	"gorm.io/gorm"
)

// 全局db，外部注入
var db *gorm.DB

func InitDao(database *gorm.DB) {
	db = database
}

// 根据ID查询图书
func GetTodoById(ID int) (*model.Todo, error) {
	var b model.Todo
	err := db.Where("id = ?", ID).First(&b).Error
	return &b, err
}

// 查询全部图书
func GetAllTodo() ([]model.Todo, error) {
	var ToDoLIST []model.Todo
	err := db.Find(&ToDoLIST).Error
	return ToDoLIST, err
}

// 新增图书
func CreateTodo(todo *model.Todo) error {
	return db.Create(todo).Error
}

// 更新todo
func UpdateTodo(todo *model.Todo) error {
	return db.Save(todo).Error
}

// 删除图书
func DeleteTodo(ID int) error {
	return db.Where("id = ?", ID).Delete(&model.Todo{}).Error
}
