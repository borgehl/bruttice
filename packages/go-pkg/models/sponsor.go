package models

// SponsorDateLayout is how StartsOn and EndsOn are written.
const SponsorDateLayout = "2006-01-02"

// Sponsor is a paid card shown once at the end of the first page of the calendar. Sponsors are
// sold directly and managed by an admin.
type Sponsor struct {
	BaseModel
	Name string
	// one short line under the name
	Text string
	// where a click goes. Visitors reach it through the backend, which counts the click.
	Link string
	// the picture, empty if there is none. Held in the database, so a list must not load it
	// unless it needs it.
	Image     []byte
	ImageType string
	// the first and last day the card is shown, as 2006-01-02 in the site's time zone. Empty
	// means no limit on that side.
	StartsOn string
	EndsOn   string
	Enabled  bool
}

// ActiveOn says if the card is shown on day (2006-01-02). The dates are in a layout that sorts as text.
func (s Sponsor) ActiveOn(day string) bool {
	if !s.Enabled {
		return false
	}
	if s.StartsOn != "" && day < s.StartsOn {
		return false
	}
	if s.EndsOn != "" && day > s.EndsOn {
		return false
	}
	return true
}

// SponsorDailyCount is how often a sponsor's card was seen and clicked on one day. It holds no
// cookie, address or visitor id, so it can be counted without consent.
type SponsorDailyCount struct {
	Date        string `gorm:"primaryKey"`
	SponsorID   int64  `gorm:"primaryKey"`
	Impressions int64
	Clicks      int64
}
