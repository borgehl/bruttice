package models

import "gorm.io/gorm"

type BaseModel struct {
	ID        int64          `gorm:"unique;primaryKey;autoIncrement"`
	CreatedAt int64          `gorm:"autoCreateTime;<-:create"`
	UpdatedAt int64          `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
