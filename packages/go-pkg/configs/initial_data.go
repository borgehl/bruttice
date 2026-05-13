package configs

import "github.com/borgehl/bruttice/packages/go-pkg/models"

var RootUser = models.User{
	Email: "admin@example.com",
}

func SetRootEmail(email string) {
	RootUser.Email = email
}

const (
	ExampleVenueOneName   = "The Example Venue"
	ExampleVenueTwoName   = "Community Arts Space"
	ExampleVenueThreeName = "The Local Theater"
	UnlistedName          = "Unlisted (see link)"
)

var Venues = []models.Venue{
	{
		Name:          ExampleVenueOneName,
		Link:          "https://example.com/venue-one",
		IsScraped:     false,
		DefaultRegion: CentralRegion,
	},
	{
		Name:          ExampleVenueTwoName,
		Link:          "https://example.com/venue-two",
		IsScraped:     false,
		DefaultRegion: CentralRegion,
	},
	{
		Name:          ExampleVenueThreeName,
		Link:          "https://example.com/venue-three",
		IsScraped:     false,
		DefaultRegion: EasternRegion,
	},
	{
		Name:      UnlistedName,
		IsScraped: false,
	},
}

const (
	TwentyOnePlusAge = "21+"
	EighteenPlusAge  = "18+"
	SixteenPlusAge   = "16+"
	AllAges          = "all-ages"
	UnknownAges      = "unknown age limit"
)

var AgeGroups = map[string]uint8{
	UnknownAges:      255,
	TwentyOnePlusAge: 21,
	EighteenPlusAge:  18,
	SixteenPlusAge:   16,
	AllAges:          1,
}

const (
	CentralRegion = "central"
	EasternRegion = "eastern"
	WesternRegion = "western"
)

var Regions = map[string]string{
	CentralRegion: "America/Chicago",
	EasternRegion: "America/New_York",
	WesternRegion: "America/Los_Angeles",
}

const (
	ComedyTag     = "comedy"
	CommunityTag  = "community"
	DIYTag        = "diy"
	FreeTag       = "free"
	LiteratureTag = "literature"
	MusicTag      = "music"
	RecurringTag  = "recurring"
	TheaterTag    = "theater"
)

var Tags = []string{
	ComedyTag,
	CommunityTag,
	MusicTag,
	DIYTag,
	FreeTag,
	LiteratureTag,
	RecurringTag,
	TheaterTag,
}
