package models

type VisitorLocation struct {
	BaseModel

	Country   string
	StateCode string
	City      string

	VisitorID string `gorm:"uniqueIndex"`
}
