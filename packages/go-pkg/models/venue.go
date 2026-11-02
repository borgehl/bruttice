package models

import "strings"

type Venue struct {
	BaseModel

	Name          string `gorm:"unique"`
	Slug          string
	Link          string
	IsScraped     bool
	DefaultRegion string
}

func (v *Venue) InternalUpdate() {
	v.Slug = strings.ToLower(strings.TrimPrefix(v.Name, "The "))
}
