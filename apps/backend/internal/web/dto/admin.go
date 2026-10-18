package dto

// Admin login request/response types
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Email   string `json:"email,omitempty"`
	Message string `json:"message,omitempty"`
}

// Admin event with score info for dashboard
type AdminEventResponse struct {
	Guid      string   `json:"guid"`
	Title     string   `json:"title"`
	Link      string   `json:"link"`
	Venue     string   `json:"venue"`
	VenueLink string   `json:"venueLink"`
	Age       string   `json:"age"`
	Region    string   `json:"region"`
	Date      string   `json:"date"`
	Time      string   `json:"time"`
	StartTime string   `json:"startTime"`
	DoorsTime string   `json:"doorsTime,omitempty"`
	CoverCost string   `json:"coverCost"`
	Tags      []string `json:"tags"`
	Email     string   `json:"email"`
	Score     int      `json:"score"`
}

type AdminEventDayResponse struct {
	Day    string               `json:"day"`
	Events []AdminEventResponse `json:"events"`
}

type AdminDashboardResponse struct {
	Email            string                  `json:"email"`
	AwaitingApproval []AdminEventDayResponse `json:"awaitingApproval"`
	Upcoming         []AdminEventDayResponse `json:"upcoming"`
}

type VoteRequest struct {
	Guid string `json:"guid"`
	Vote string `json:"vote"`
}

type VoteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// Admin add event request (same fields but no email verification needed)
type AdminAddEventRequest struct {
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
}

type AdminAddEventResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// Admin update event request - all fields optional except guid
type AdminUpdateEventRequest struct {
	Guid      string   `json:"guid"`
	Title     *string  `json:"title,omitempty"`
	Link      *string  `json:"link,omitempty"`
	CoverCost *string  `json:"coverCost,omitempty"`
	Venue     *string  `json:"venue,omitempty"`
	Region    *string  `json:"region,omitempty"`
	Age       *string  `json:"age,omitempty"`
	Date      *string  `json:"date,omitempty"`
	StartTime *string  `json:"startTime,omitempty"`
	DoorsTime *string  `json:"doorsTime,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

type AdminUpdateEventResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type IngestSourceResponse struct {
	Source       string   `json:"source"`
	Status       string   `json:"status"`
	ScrapedCount int      `json:"scrapedCount"`
	SavedCount   int      `json:"savedCount"`
	FailedCount  int      `json:"failedCount"`
	Errors       []string `json:"errors"`
}

type IngestRunsResponse struct {
	Runs    []IngestRunResponse `json:"runs"`
	HasMore bool                `json:"hasMore"`
}

type IngestRunResponse struct {
	ID          int64                  `json:"id"`
	StartedAt   int64                  `json:"startedAt"`
	FinishedAt  int64                  `json:"finishedAt"`
	DedupeError string                 `json:"dedupeError,omitempty"`
	Scraped     int                    `json:"scraped"`
	Saved       int                    `json:"saved"`
	Failed      int                    `json:"failed"`
	Sources     []IngestSourceResponse `json:"sources"`
}

type StartIngestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
