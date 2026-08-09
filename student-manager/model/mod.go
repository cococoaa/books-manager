package model

import (
	"time"

	"gorm.io/gorm"
)

type Student struct {
	gorm.Model
	Name  string    `gorm:"not null;comment:'学生姓名'"`
	Age   int       `gorm:"not null;comment:'学生年龄'"`
	Score gezi      `gorm:"serializer:json;comment:'学生分数'" json:"score"`
	Time  time.Time `gorm:"not null;comment:'录入时间'"`
}
type gezi struct {
	Yuwen   int `json:"yuwen"`
	Math    int `json:"math"`
	English int `json:"english"`
	Physics int `json:"physics"`
}
