package dao

import (
	"dgo.baisic.print/baisic/studentguanlixitong/model"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDao(database *gorm.DB) {
	db = database
}

// 新增学生
func CreateStudent(student *model.Student) error {
	return db.Create(student).Error
}

// 查询全部学生
func GETALLStudent() ([]model.Student, error) {
	var studentlist []model.Student
	err := db.Find(&studentlist).Error
	return studentlist, err
}

// 查询单个学生信息
func GETStudentbyID(ID int) (*model.Student, error) {
	var s model.Student
	err := db.Where("id=?", ID).First(&s).Error
	return &s, err
}

// 更新学生信息
func UpdateStudent(student *model.Student) error {
	return db.Save(student).Error
}

// 删除学生
func DeleteStudentByID(ID int) error {
	return db.Where("id=?", ID).Delete(&model.Student{}).Error
}
