// Package dto holds the request and response bodies of the HTTP API.
//
// Handlers decode requests into these types and encode responses from them, and never
// expose a database model. Public endpoints are in public.go, the admin dashboard in
// admin.go, and the internal ingest API in ingest.go and common.go.
package dto

type EventResponse struct {
	Guid      string   `json:"guid"`
	Title     string   `json:"title"`
	Link      string   `json:"link"`
	Venue     string   `json:"venue"`
	VenueLink string   `json:"venueLink"`
	Age       string   `json:"age"`
	Region    string   `json:"region"`
	StartAt   string   `json:"startAt"`
	DoorsAt   string   `json:"doorsAt,omitempty"`
	CoverCost string   `json:"coverCost"`
	FlyerURL  string   `json:"flyerUrl,omitempty"`
	Tags      []string `json:"tags"`
}

type EventDayResponse struct {
	Day    string          `json:"day"`
	Events []EventResponse `json:"events"`
}

type VenueResponse struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	Link string `json:"link"`
	// Region is the name of the venue's region, empty if it has none.
	Region string `json:"region"`
}

// SiteResponse is what differs between deployments of the site: its name and contact
// address, the About text, and which optional pages it has. It comes from config.yaml.
type SiteResponse struct {
	Name         string       `json:"name"`
	Region       string       `json:"region"`
	ContactEmail string       `json:"contactEmail"`
	Description  string       `json:"description"`
	Features     SiteFeatures `json:"features"`
	// PopBookerURL is the calendar the Pop! Booker page embeds, empty if there is none.
	PopBookerURL string `json:"popBookerUrl,omitempty"`
}

type SiteFeatures struct {
	PopBooker bool `json:"popBooker"`
}

type TagResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type AgeResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Value uint8  `json:"value"`
}

type RegionResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type FilterOptionsResponse struct {
	Tags    []TagResponse    `json:"tags"`
	Venues  []VenueResponse  `json:"venues"`
	Ages    []AgeResponse    `json:"ages"`
	Regions []RegionResponse `json:"regions"`
}

type CreateFilterRequest struct {
	Tags          []string `json:"tags"`
	TagsInclude   bool     `json:"tagsInclude"`
	Ages          []uint8  `json:"ages"`
	AgesLessThan  bool     `json:"agesLessThan"`
	Venues        []string `json:"venues"`
	VenuesInclude bool     `json:"venuesInclude"`
}

type CreateFilterResponse struct {
	FilterKey string `json:"filterKey"`
}

type GetFilterResponse struct {
	Tags          []string `json:"tags,omitempty"`
	TagsInclude   bool     `json:"tagsInclude,omitempty"`
	Age           uint8    `json:"age,omitempty"`
	AgesLessThan  bool     `json:"agesLessThan,omitempty"`
	Venues        []string `json:"venues,omitempty"`
	VenuesInclude bool     `json:"venuesInclude,omitempty"`
}

type AddEventRequest struct {
	Title     string   `json:"title"`
	Link      string   `json:"link"`
	Venue     string   `json:"venue"`
	Region    string   `json:"region"`
	Age       string   `json:"age"`
	Date      string   `json:"date"`
	StartTime string   `json:"startTime"`
	DoorsTime string   `json:"doorsTime"`
	CoverCost string   `json:"coverCost"`
	Tags      []string `json:"tags"`
	Email     string   `json:"email"`
}

type AddEventResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type VerifyEventResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type VisitorConsentRequest struct {
	Path  string `json:"path"`
	Query string `json:"query"`
}
