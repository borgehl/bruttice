package models

const (
	IngestSourceStatusOk     = "ok"
	IngestSourceStatusErrors = "errors"
	IngestSourceStatusFailed = "failed"
)

// one execution of IngestorClient.RunAll
type IngestRun struct {
	BaseModel
	StartedAt   int64
	FinishedAt  int64
	DedupeError string
	Sources     []IngestRunSource `gorm:"foreignKey:IngestRunID"`
}

// the result of a single scraper within an IngestRun
type IngestRunSource struct {
	BaseModel
	IngestRunID int64 `gorm:"index"`
	Source      string
	Status      string
	// events the scraper returned
	ScrapedCount int
	// events that were saved without error
	SavedCount int
	// events that failed to save
	FailedCount int
	Errors      []string `gorm:"serializer:json"`
}
