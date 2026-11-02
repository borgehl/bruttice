package dto

// IngestRunCreatedResponse is returned when a scraper opens a run.
type IngestRunCreatedResponse struct {
	ID int64 `json:"id"`
}

// IngestResultResponse reports how a payload was saved. Errors holds the payload's own
// errors followed by one entry for each event that failed to save.
type IngestResultResponse struct {
	Saved  int      `json:"saved"`
	Errors []string `json:"errors"`
}
