package api

import (
	"fmt"
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/apps/backend/internal/web/mapper"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/passwordsx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"net/http"
	"strings"
	"time"
)

func (h *APIHandler) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	if req.Email == "" || req.Password == "" {
		h.writeError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	tx := h.database.DB.Begin()
	defer tx.Commit()

	user, err := h.database.GetUser(models.User{Email: req.Email}, tx)
	if err != nil || user.ID == 0 {
		h.writeError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}
	if !passwordsx.ValidPassword(req.Password, user.PasswordHash) {
		h.writeError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Set session cookie
	cookie := h.cookieSessions.CreateSessionCookie(user, tx)
	if cookie == nil {
		h.logger.Errorf("failed to create session cookie for user: %s", req.Email)
		h.writeError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}
	http.SetCookie(w, cookie)
	h.writeJSON(w, http.StatusOK, dto.LoginResponse{Success: true, Email: user.Email, Name: user.Name, Notifications: mapper.Notifications(user.Settings)})
}

func (h *APIHandler) handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "user-session-cookie",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	h.writeJSON(w, http.StatusOK, dto.SuccessResponse{Success: true})
}

func (h *APIHandler) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	user := h.getAuthenticatedUser(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	tx := h.database.DB.Begin()
	defer tx.Commit()

	// Get events awaiting approval
	awaitingApprovalEvents, err := h.database.GetAwaitingApprovalEvents(tx)
	if err != nil {
		h.logger.Errorf("failed to get awaiting approval events: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to fetch events")
		return
	}

	// Get upcoming approved events
	upcomingFilter := models.CalendarFilterBackend{}
	upcomingFilter.UpdateDates()
	upcomingEvents, err := h.database.GetIndexEvents(upcomingFilter, tx)
	if err != nil {
		h.logger.Errorf("failed to get upcoming events: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to fetch events")
		return
	}

	unread, err := h.inbox.CountUnreadInboxThreads()
	if err != nil {
		// the dashboard is still useful without the count
		h.logger.Errorf("failed to count unread inbox threads: %v", err)
	}

	h.writeJSON(w, http.StatusOK, dto.AdminDashboardResponse{
		Email:            user.Email,
		Name:             user.Name,
		UnreadMessages:   unread,
		AwaitingApproval: mapper.AdminEventDayResponses(awaitingApprovalEvents),
		Upcoming:         mapper.AdminEventDayResponses(upcomingEvents),
	})
}

func (h *APIHandler) handleAdminVote(w http.ResponseWriter, r *http.Request) {
	user := h.getAuthenticatedUser(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req dto.VoteRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	if req.Guid == "" || req.Vote == "" {
		h.writeError(w, http.StatusBadRequest, "Guid and vote are required")
		return
	}

	validVotes := map[string]bool{"Approve": true, "Deny": true, "Keep": true, "Remove": true}
	if !validVotes[req.Vote] {
		h.writeError(w, http.StatusBadRequest, "Invalid vote value")
		return
	}

	tx := h.database.DB.Begin()
	defer tx.Commit()

	if err := h.processVote(req.Vote, req.Guid, *user, tx); err != nil {
		h.logger.Errorf("failed to process vote: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to process vote")
		return
	}
	h.writeJSON(w, http.StatusOK, dto.VoteResponse{Success: true})
}

func (h *APIHandler) handleAdminCheckSession(w http.ResponseWriter, r *http.Request) {
	user := h.getAuthenticatedUser(r)
	if user == nil {
		h.writeJSON(w, http.StatusOK, dto.LoginResponse{Success: false})
		return
	}
	h.writeJSON(w, http.StatusOK, dto.LoginResponse{Success: true, Email: user.Email, Name: user.Name, Notifications: mapper.Notifications(user.Settings)})
}

// Helper to get authenticated user from request
func (h *APIHandler) getAuthenticatedUser(r *http.Request) *models.User {
	// tests stand in for the session cookie lookup, which needs a database
	if h.authenticate != nil {
		return h.authenticate(r)
	}
	tx := h.database.DB.Begin()
	defer tx.Commit()
	r, err := h.cookieSessions.RequestWithCookieUser(r, tx)
	if err != nil || r == nil {
		return nil
	}
	return h.cookieSessions.GetUserFromRequest(r)
}

func (h *APIHandler) handleAdminAddEvent(w http.ResponseWriter, r *http.Request) {
	user := h.getAuthenticatedUser(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req dto.AdminAddEventRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}

	// Validate required fields
	if req.Title == "" {
		h.writeError(w, http.StatusBadRequest, "Title is required")
		return
	}
	if req.Link == "" {
		h.writeError(w, http.StatusBadRequest, "Event URL is required")
		return
	}
	if req.Venue == "" {
		h.writeError(w, http.StatusBadRequest, "Venue is required")
		return
	}
	if req.Date == "" {
		h.writeError(w, http.StatusBadRequest, "Date is required")
		return
	}
	if req.StartTime == "" {
		h.writeError(w, http.StatusBadRequest, "Start time is required")
		return
	}

	event, err := mapper.AdminAddEventRequest(req, time.Now())
	if err != nil {
		h.writeError(w, http.StatusBadRequest, eventFormErrorMessage(err))
		return
	}

	tx := h.database.DB.Begin()
	if tx.Error != nil {
		h.logger.Errorf("failed to begin transaction: %v", tx.Error)
		h.writeError(w, http.StatusInternalServerError, "Failed to save event")
		return
	}
	defer func() {
		if err := tx.Commit().Error; err != nil {
			h.logger.Errorf("failed to commit transaction: %v", err)
		}
	}()

	event, err = h.database.SaveEvent(event, false, tx)
	if err != nil {
		h.logger.Errorf("failed to save admin event: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to save event")
		return
	}

	h.writeJSON(w, http.StatusOK, dto.AdminAddEventResponse{Success: true, Message: "Event created successfully."})
}

func (h *APIHandler) handleAdminGetAges(w http.ResponseWriter, r *http.Request) {
	user := h.getAuthenticatedUser(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	tx := h.database.DB.Begin()
	defer tx.Commit()

	// Get ages
	ageModels, err := h.database.GetAges(tx)
	if err != nil {
		h.logger.Errorf("failed to get ages: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get ages")
		return
	}

	h.writeJSON(w, http.StatusOK, mapper.AgeResponses(ageModels))
}

func (h *APIHandler) handleAdminUpdateEvent(w http.ResponseWriter, r *http.Request) {
	user := h.getAuthenticatedUser(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req dto.AdminUpdateEventRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	if req.Guid == "" {
		h.writeError(w, http.StatusBadRequest, "Event GUID is required")
		return
	}

	tx := h.database.DB.Begin()
	if tx.Error != nil {
		h.logger.Errorf("failed to begin transaction: %v", tx.Error)
		h.writeError(w, http.StatusInternalServerError, "Failed to update event")
		return
	}
	defer func() {
		if err := tx.Commit().Error; err != nil {
			h.logger.Errorf("failed to commit transaction: %v", err)
		}
	}()

	// Edits of one event run one after the other. Without the lock two edits of different fields both
	// read the old row, and the one that commits last writes the old value of the other's field back.
	if err := h.database.LockEventByGuid(req.Guid, tx); err != nil {
		h.logger.Errorf("failed to lock event %s: %v", req.Guid, err)
		h.writeError(w, http.StatusInternalServerError, "Failed to update event")
		return
	}

	event, err := h.database.GetEventByGuid(req.Guid, tx)
	if err != nil {
		h.logger.Errorf("failed to get event by guid: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to find event")
		return
	}
	if event.ID == 0 {
		h.writeError(w, http.StatusNotFound, "Event not found")
		return
	}

	// Update fields if provided
	updated := false
	if req.Title != nil && *req.Title != event.Title {
		event.Title = *req.Title
		updated = true
	}
	if req.Link != nil && *req.Link != event.Link {
		event.Link = mapper.NormalizeURL(*req.Link)
		updated = true
	}
	if req.CoverCost != nil && *req.CoverCost != event.CoverCost {
		event.CoverCost = *req.CoverCost
		updated = true
	}
	if req.FlyerURL != nil {
		flyer := mapper.FlyerURL(*req.FlyerURL)
		if flyer == "" && strings.TrimSpace(*req.FlyerURL) != "" {
			h.writeError(w, http.StatusBadRequest, "Flyer must be an http or https link")
			return
		}
		if flyer != event.FlyerURL {
			event.FlyerURL = flyer
			updated = true
		}
	}
	if req.Venue != nil && *req.Venue != event.Venue.Name {
		venue, err := h.database.GetVenue(models.Venue{Name: *req.Venue}, tx)
		if err != nil || venue.ID == 0 {
			h.writeError(w, http.StatusBadRequest, "Venue not found")
			return
		}
		event.VenueID = venue.ID
		event.Venue = venue
		updated = true
	}
	if req.Region != nil && *req.Region != event.Region.Name {
		region, err := h.database.GetRegion(models.Region{Name: *req.Region}, tx)
		if err != nil || region.ID == 0 {
			h.writeError(w, http.StatusBadRequest, "Region not found")
			return
		}
		event.RegionID = region.ID
		event.Region = region
		updated = true
	}
	if req.Age != nil && *req.Age != event.Age.Name {
		age, err := h.database.GetAge(models.Age{Name: *req.Age}, tx)
		if err != nil || age.ID == 0 {
			h.writeError(w, http.StatusBadRequest, "Age not found")
			return
		}
		event.AgeID = age.ID
		event.Age = age
		updated = true
	}
	if req.Date != nil {
		timeStr := event.StartAt.In(util.TZLocation).Format("15:04")
		newStart, err := time.ParseInLocation(mapper.FormDateTimeFormat, fmt.Sprintf("%s %s", *req.Date, timeStr), util.TZLocation)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid date format")
			return
		}
		if !newStart.Equal(event.StartAt) {
			if event.DoorsAt != nil {
				doorsTimeStr := event.DoorsAt.In(util.TZLocation).Format("15:04")
				if newDoors, err := time.ParseInLocation(mapper.FormDateTimeFormat, fmt.Sprintf("%s %s", *req.Date, doorsTimeStr), util.TZLocation); err == nil {
					event.DoorsAt = &newDoors
				}
			}
			event.StartAt = newStart
			updated = true
		}
	}
	if req.StartTime != nil {
		dateStr := event.StartAt.In(util.TZLocation).Format("2006-01-02")
		newStart, err := time.ParseInLocation(mapper.FormDateTimeFormat, fmt.Sprintf("%s %s", dateStr, *req.StartTime), util.TZLocation)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid start time format")
			return
		}
		if !newStart.Equal(event.StartAt) {
			event.StartAt = newStart
			updated = true
		}
	}
	if req.DoorsTime != nil {
		if *req.DoorsTime == "" {
			event.DoorsAt = nil
			updated = true
		} else {
			dateStr := event.StartAt.In(util.TZLocation).Format("2006-01-02")
			newDoors, err := time.ParseInLocation(mapper.FormDateTimeFormat, fmt.Sprintf("%s %s", dateStr, *req.DoorsTime), util.TZLocation)
			if err != nil {
				h.writeError(w, http.StatusBadRequest, "Invalid doors time format")
				return
			}
			if event.DoorsAt == nil || !newDoors.Equal(*event.DoorsAt) {
				event.DoorsAt = &newDoors
				updated = true
			}
		}
	}
	if req.Tags != nil {
		newTags := make([]models.Tag, 0, len(req.Tags))
		for _, tagName := range req.Tags {
			newTags = append(newTags, models.Tag{Name: tagName})
		}
		event.Tags = newTags
		updated = true
	}

	if !updated {
		h.writeJSON(w, http.StatusOK, dto.AdminUpdateEventResponse{Success: true, Message: "No changes detected."})
		return
	}

	// Save the updated event (resolves tags by name and replaces associations)
	if _, err := h.database.UpdateEvent(event, tx); err != nil {
		h.logger.Errorf("failed to save updated event: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to update event")
		return
	}

	h.writeJSON(w, http.StatusOK, dto.AdminUpdateEventResponse{Success: true, Message: "Event updated successfully."})
}
