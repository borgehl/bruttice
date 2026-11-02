package dto

import "github.com/borgehl/bruttice/packages/go-pkg/platform/util"

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
	// Notifications are the emails the signed in admin chose, set once they are signed in
	Notifications *AdminNotifications `json:"notifications,omitempty"`
}

// AdminNotifications are the emails an admin gets: whether they chose each one.
type AdminNotifications struct {
	NewEvents    bool `json:"new_events"`
	SystemErrors bool `json:"system_errors"`
	Inbox        bool `json:"inbox"`
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
	// DeliveredAt is unix seconds for an email an admin sent, 0 if Resend has not said it was delivered.
	DeliveredAt int64 `json:"deliveredAt"`
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

// AdminSettingsRequest is what an admin can change about themself. A field that is absent is
// left as it is.
type AdminSettingsRequest struct {
	// Name is how other admins see them. Empty goes back to the email.
	Name          util.Optional[string]   `json:"name"`
	Notifications AdminNotificationsPatch `json:"notifications"`
}

// AdminNotificationsPatch holds the notifications a request changes; absent fields are not written.
type AdminNotificationsPatch struct {
	NewEvents    util.Optional[bool] `json:"new_events"`
	SystemErrors util.Optional[bool] `json:"system_errors"`
	Inbox        util.Optional[bool] `json:"inbox"`
}

type InboxReplyRequest struct {
	Body string `json:"body"`
}

// InboxComposeRequest is a new email to someone, starting a conversation.
type InboxComposeRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// InboxComposeResponse names the conversation the email started.
type InboxComposeResponse struct {
	ThreadID string `json:"threadId"`
}

// ConfigRegion is a region as the admin settings list it.
type ConfigRegion struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Events int64  `json:"events"`
	Venues int64  `json:"venues"`
	// Retired regions are no longer offered. MergedInto is where their events and old name went.
	Retired    bool   `json:"retired"`
	MergedInto string `json:"mergedInto,omitempty"`
}

type ConfigTag struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Events int64  `json:"events"`
}

type ConfigAge struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Value  uint8  `json:"value"`
	Events int64  `json:"events"`
}

type ConfigVenue struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Link   string `json:"link"`
	Region string `json:"region"`
	// Scraped venues are kept up to date by their scraper and can't be edited here.
	Scraped bool  `json:"scraped"`
	Events  int64 `json:"events"`
}

// ConfigResponse is every list an admin can change.
type ConfigResponse struct {
	Regions []ConfigRegion `json:"regions"`
	Tags    []ConfigTag    `json:"tags"`
	Ages    []ConfigAge    `json:"ages"`
	Venues  []ConfigVenue  `json:"venues"`
}

type ConfigNameRequest struct {
	Name string `json:"name"`
}

type ConfigAgeRequest struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type ConfigVenueRequest struct {
	Name   string `json:"name"`
	Link   string `json:"link"`
	Region string `json:"region"`
}

// ConfigRetireRegionRequest says what happens to the events of a retired region: they move to
// MoveToID, or they are deleted. Exactly one of the two.
type ConfigRetireRegionRequest struct {
	MoveToID     *int64 `json:"moveToId"`
	DeleteEvents bool   `json:"deleteEvents"`
}

// ConfigChangeResponse tells what a change did to events.
type ConfigChangeResponse struct {
	EventsMoved   int64 `json:"eventsMoved"`
	EventsDeleted int64 `json:"eventsDeleted"`
}

// InboxDeleteArchivedResponse says how many messages the removal of the archived conversations deleted.
type InboxDeleteArchivedResponse struct {
	Deleted int64 `json:"deleted"`
}
