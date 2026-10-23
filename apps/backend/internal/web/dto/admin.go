package dto

// Admin login request/response types
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Email   string `json:"email,omitempty"`
	Name    string `json:"name,omitempty"`
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
	FlyerURL  string   `json:"flyerUrl,omitempty"`
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
	Name             string                  `json:"name,omitempty"`
	AwaitingApproval []AdminEventDayResponse `json:"awaitingApproval"`
	Upcoming         []AdminEventDayResponse `json:"upcoming"`
	// UnreadMessages is how many inbox conversations have a message nobody has read.
	UnreadMessages int64 `json:"unreadMessages"`
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
	Guid      string  `json:"guid"`
	Title     *string `json:"title,omitempty"`
	Link      *string `json:"link,omitempty"`
	CoverCost *string `json:"coverCost,omitempty"`
	// FlyerURL replaces the flyer link, an empty string removes the flyer.
	FlyerURL  *string  `json:"flyerUrl,omitempty"`
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

// InboxMessageResponse is one message of a conversation.
type InboxMessageResponse struct {
	ID int64 `json:"id"`
	// Source is contact_form, email or admin_reply.
	Source    string `json:"source"`
	FromName  string `json:"fromName"`
	FromEmail string `json:"fromEmail"`
	// Body is plain text.
	Body      string `json:"body"`
	CreatedAt int64  `json:"createdAt"`
	Unread    bool   `json:"unread"`
}

// InboxThreadResponse is a conversation, its messages oldest first.
type InboxThreadResponse struct {
	ThreadID string `json:"threadId"`
	Subject  string `json:"subject"`
	// FromName and FromEmail are who the conversation is with.
	FromName  string `json:"fromName"`
	FromEmail string `json:"fromEmail"`
	Unread    bool   `json:"unread"`
	Archived  bool   `json:"archived"`
	// LastAt is the time of the latest message, as unix seconds.
	LastAt   int64                  `json:"lastAt"`
	Messages []InboxMessageResponse `json:"messages"`
}

// AdminSettingsRequest is what an admin can change about themself.
type AdminSettingsRequest struct {
	// Name is how other admins see them. Empty goes back to the email.
	Name string `json:"name"`
}

type InboxReplyRequest struct {
	Body string `json:"body"`
}
