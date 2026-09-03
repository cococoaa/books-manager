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
	Time  time.Time `gorm:"comment:'录入时间'" json:"time"`
}
type gezi struct {
	Yuwen   int `json:"yuwen"`
	Math    int `json:"math"`
	English int `json:"english"`
	Physics int `json:"physics"`
}

// 用户（用于登录鉴权）
type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(64);not null;uniqueIndex;comment:'用户名'" json:"username"`
	Password string `gorm:"type:varchar(255);not null;comment:'密码'" json:"-"`
}
