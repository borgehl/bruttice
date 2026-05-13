package models

type Tag struct {
	BaseModel
	Name string `gorm:"unique"`
}
