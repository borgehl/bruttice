package middleware

import (
	"context"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/dbx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/geoipx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const VisitorIDKey contextKey = "visitor-id"

// MiddlewareVisitorTracker logs page accesses for visitors who have consented
// (identified by the presence of a valid "visitor-id" cookie). Requests without
// a valid cookie are counted anonymously and passed through.
func MiddlewareVisitorTracker(db *dbx.DatabaseClient, log *logx.Logger, geoClient *geoipx.GeoIpClient, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("visitor-id")
		if err != nil {
			// No cookie — user has not consented. Count the hit anonymously.
			if ierr := db.IncrementUntrackedHit(r.URL.Path); ierr != nil {
				log.Errorf("visitor_tracker: failed to increment untracked hit: %v", ierr)
			}
			handler(w, r)
			return
		}

		visitorID, err := uuid.Parse(cookie.Value)
		if err != nil {
			log.Errorf("invalid cookie (%s): %+v", err, cookie.Value)
			// Cookie value is not a valid UUID — count as untracked and pass through.
			if ierr := db.IncrementUntrackedHit(r.URL.Path); ierr != nil {
				log.Errorf("visitor_tracker: failed to increment untracked hit: %v", ierr)
			}
			handler(w, r)
			return
		}

		visitorIDStr := visitorID.String()
		// Store visitor ID in context so handlers can read it if needed.
		ctx := context.WithValue(r.Context(), VisitorIDKey, visitorIDStr)
		r = r.WithContext(ctx)

		accessLog := models.NewAccessLog(r, visitorIDStr)

		tx := db.DB.Begin()
		defer tx.Commit()

		if err := db.SaveVisitorLog(accessLog, models.VisitorLocation{VisitorID: visitorIDStr}, tx); err != nil {
			tx.Rollback()
			log.Errorf("visitor_tracker: failed to save access log for visitor %s: %v", visitorIDStr, err)
		}

		handler(w, r)
	}
}
