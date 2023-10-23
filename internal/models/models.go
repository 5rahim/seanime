package models

import (
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

type Token struct {
	BaseModel
	Value string `json:"value"`
}

type LocalFiles struct {
	BaseModel
	Value []byte `gorm:"column:value" json:"value"`
}
