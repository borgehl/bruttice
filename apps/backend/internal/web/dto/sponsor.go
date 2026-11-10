package dto

// SponsorResponse is a sponsor card as visitors get it. The sponsor's own link is not in it: the
// card links to ClickURL, which counts the click and then redirects.
type SponsorResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Text string `json:"text"`
	// empty if the sponsor has no picture
	ImageURL string `json:"imageUrl"`
	ClickURL string `json:"clickUrl"`
}

// AdminSponsorResponse is a sponsor as the admin list shows it.
type AdminSponsorResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Text     string `json:"text"`
	Link     string `json:"link"`
	StartsOn string `json:"startsOn"`
	EndsOn   string `json:"endsOn"`
	Enabled  bool   `json:"enabled"`
	// empty if the sponsor has no picture
	ImageURL string `json:"imageUrl"`
	// whether the card is shown today
	Active        bool  `json:"active"`
	Impressions   int64 `json:"impressions"`
	Clicks        int64 `json:"clicks"`
	Impressions30 int64 `json:"impressions30"`
	Clicks30      int64 `json:"clicks30"`
}

// SponsorRequest creates or changes a sponsor.
type SponsorRequest struct {
	Name     string `json:"name"`
	Text     string `json:"text"`
	Link     string `json:"link"`
	StartsOn string `json:"startsOn"`
	EndsOn   string `json:"endsOn"`
	Enabled  bool   `json:"enabled"`
	// a new picture as base64, left out to keep the one the sponsor has
	Image       *string `json:"image"`
	RemoveImage bool    `json:"removeImage"`
}
