package models

import (
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	BaseModel

	Guid string `gorm:"unique"`

	Title     string
	Link      string
	DoorsAt   *time.Time
	StartAt   time.Time
	CoverCost string
	Tags      []Tag `gorm:"many2many:event_tags"`

	VerificationKey      *string
	VerifiedAt           *int64 `gorm:"index"`
	VerificationEmail    *string
	ApprovedAt           *int64 `gorm:"index"`
	EventAdminJudgements []EventAdminJudgement

	AgeID int64
	Age   Age

	RegionID int64
	Region   Region

	VenueID int64
	Venue   Venue
}

// Idempotent guid updater. Does not overwrite if Guid is already set.
// If id is non-nil and non-empty, sets Guid to "<VenueName>-<id>".
// Otherwise hashes Title+StartAt+VenueName.
func (e *Event) UpdateGuid(id *string) {
	if e.Guid != "" {
		return
	}
	if id != nil && *id != "" {
		e.Guid = fmt.Sprintf("%s-%s", e.Venue.Name, *id)
		return
	}
	e.Guid = util.Base64EncodeString(e.Title + e.StartAt.String() + e.Venue.Name)
}

func (e *Event) CreateVerificationKey() string {
	key := util.Base64EncodeString(uuid.NewString())
	e.VerificationKey = &key
	return key
}

func (e *Event) Str() string {
	e.UpdateGuid(nil)
	doors := ""
	if e.DoorsAt != nil {
		doors = e.DoorsAt.Format("2006-1-2 3:04 pm")
	}
	return fmt.Sprintf(
		"%s @ %s --> %s\n%s\n%s %s: doors %s, start %s\n%+v",
		e.Title, e.Venue.Name, e.Link,
		e.Guid,
		e.Age.Name, e.CoverCost, doors, e.StartAt.Format("2006-1-2 3:04 pm"),
		e.Tags,
	)
}

func (e *Event) AddTagStr(tag string) {
	e.Tags = append(e.Tags, Tag{Name: tag})
}

func (e *Event) SetTimezoneCentral() {
	e.StartAt = util.DatetimeIsCentral(e.StartAt)
	if e.DoorsAt != nil {
		dt := util.DatetimeIsCentral(*e.DoorsAt)
		e.DoorsAt = &dt
	}
}
