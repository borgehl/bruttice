package configs

const (
	// UnlistedName is the reserved venue name for events whose venue is not in the DB.
	// Region enforcement is skipped for this venue.
	UnlistedName = "Unlisted (see link)"

	// FreeTag is automatically applied to events with a zero cover cost.
	FreeTag = "free"
)
