package models

type MigrationFlag struct {
	BaseModel
	Name string `gorm:"unique"`
}
