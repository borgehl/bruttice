package models

type UntrackedHit struct {
	Date  string `gorm:"primaryKey"`
	Path  string `gorm:"primaryKey"`
	Count int64
}
