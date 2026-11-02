package models

type Region struct {
	BaseModel

	Name string `gorm:"unique"`
	// unix time an admin retired the region, 0 while it is in use. A retired region stays in the
	// table so its name can still be told apart, it is just not offered any more.
	RetiredAt int64 `gorm:"index"`
	// the region a retired region was merged into: events and scrapers that still use its name
	// land there. Nil if the region was retired by deleting its events.
	MergedIntoID *int64
}

func (r Region) Retired() bool {
	return r.RetiredAt > 0
}
