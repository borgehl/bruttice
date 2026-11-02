package models

type Age struct {
	BaseModel

	Name  string `gorm:"unique"`
	Value uint8
}
