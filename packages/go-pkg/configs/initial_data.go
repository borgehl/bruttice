package configs

import "github.com/borgehl/bruttice/packages/go-pkg/models"

var RootUser = models.User{
	Email: "admin@example.com",
}

func SetRootEmail(email string) {
	RootUser.Email = email
}

const (
	// UnlistedName is the reserved venue name for events whose venue is not in the DB.
	// Region enforcement is skipped for this venue.
	UnlistedName = "Unlisted (see link)"

	// FreeTag is automatically applied to events with a zero cover cost.
	FreeTag = "free"
)
