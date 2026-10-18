package util

import (
	"fmt"
	"time"
)

// EventGuid is the identity of an event.
//
// With a non-empty id from the event's source it is "<venue>-<id>", so the event keeps
// its identity when its title or time change. Otherwise it hashes title, start time and
// venue, so a change to any of them makes a new event.
//
// The hash includes the start time's offset, so callers must pass the time in the zone
// they always have. Changing it changes the guid of every event without an id.
func EventGuid(venue string, title string, start time.Time, id *string) string {
	if id != nil && *id != "" {
		return fmt.Sprintf("%s-%s", venue, *id)
	}
	return Base64EncodeString(title + start.String() + venue)
}
