package models

type Region struct {
	BaseModel

	Name string `gorm:"unique"`
}
