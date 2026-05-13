package api

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/passwordsx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Email   string `json:"email,omitempty"`
	Message string `json:"message,omitempty"`
}

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
	AwaitingApproval []AdminEventDayResponse `json:"awaitingApproval"`
	Upcoming         []AdminEventDayResponse `json:"upcoming"`
}

type VoteRequest struct {
	Guid string `json:"guid"`
	Vote string `json:"vote"`
}

type VoteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

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

type AdminUpdateEventRequest struct {
	Guid      string   `json:"guid"`
	Title     *string  `json:"title,omitempty"`
	Link      *string  `json:"link,omitempty"`
	CoverCost *string  `json:"coverCost,omitempty"`
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

func (h *APIHandler) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
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

	cookie := h.cookieSessions.CreateSessionCookie(user, tx)
	if cookie == nil {
		h.logger.Errorf("failed to create session cookie for user: %s", req.Email)
		h.writeError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}
	http.SetCookie(w, cookie)
	h.writeJSON(w, http.StatusOK, LoginResponse{Success: true, Email: user.Email})
}

func (h *APIHandler) handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "user-session-cookie",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	h.writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *APIHandler) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	user := h.getAuthenticatedUser(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	tx := h.database.DB.Begin()
	defer tx.Commit()

	awaitingApprovalEvents, err := h.database.GetAwaitingApprovalEvents(tx)
	if err != nil {
		h.logger.Errorf("failed to get awaiting approval events: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to fetch events")
		return
	}

	upcomingFilter := models.CalendarFilterBackend{}
	upcomingFilter.UpdateDates()
	upcomingEvents, err := h.database.GetIndexEvents(upcomingFilter, tx)
	if err != nil {
		h.logger.Errorf("failed to get upcoming events: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to fetch events")
		return
	}

	h.writeJSON(w, http.StatusOK, AdminDashboardResponse{
		Email:            user.Email,
		AwaitingApproval: convertToAdminEventDays(awaitingApprovalEvents),
		Upcoming:         convertToAdminEventDays(upcomingEvents),
	})
}

func (h *APIHandler) handleAdminVote(w http.ResponseWriter, r *http.Request) {
	user := h.getAuthenticatedUser(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
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
	h.writeJSON(w, http.StatusOK, VoteResponse{Success: true})
}

func (h *APIHandler) handleAdminCheckSession(w http.ResponseWriter, r *http.Request) {
	user := h.getAuthenticatedUser(r)
	if user == nil {
		h.writeJSON(w, http.StatusOK, LoginResponse{Success: false})
		return
	}
	h.writeJSON(w, http.StatusOK, LoginResponse{Success: true, Email: user.Email})
}

func (h *APIHandler) getAuthenticatedUser(r *http.Request) *models.User {
	tx := h.database.DB.Begin()
	defer tx.Commit()
	r, err := h.cookieSessions.RequestWithCookieUser(r, tx)
	if err != nil || r == nil {
		return nil
	}
	return h.cookieSessions.GetUserFromRequest(r)
}

func convertToAdminEventDays(events []models.Event) []AdminEventDayResponse {
	if len(events) == 0 {
		return []AdminEventDayResponse{}
	}
	dayMap := make(map[string]*AdminEventDayResponse)
	var days []string

	for _, event := range events {
		dayStr := event.StartAt.In(util.TZLocation).Format("Monday, January 2, 2006")
		if _, exists := dayMap[dayStr]; !exists {
			dayMap[dayStr] = &AdminEventDayResponse{Day: dayStr, Events: []AdminEventResponse{}}
			days = append(days, dayStr)
		}
		adminEvent := AdminEventResponse{
			Guid:      event.Guid,
			Title:     event.Title,
			Link:      event.Link,
			Venue:     event.Venue.Name,
			VenueLink: event.Venue.Link,
			Age:       event.Age.Name,
			Region:    event.Region.Name,
			Date:      event.StartAt.In(util.TZLocation).Format("2006-01-02"),
			Time:      formatEventTime(event),
			StartTime: event.StartAt.In(util.TZLocation).Format("15:04"),
			CoverCost: event.CoverCost,
			Tags:      getEventTags(event),
			Score:     getScore(event),
		}
		if event.DoorsAt != nil {
			adminEvent.DoorsTime = event.DoorsAt.In(util.TZLocation).Format("15:04")
		}
		if event.VerificationEmail != nil {
			adminEvent.Email = *event.VerificationEmail
		}
		dayMap[dayStr].Events = append(dayMap[dayStr].Events, adminEvent)
	}

	result := make([]AdminEventDayResponse, 0, len(days))
	for _, day := range days {
		result = append(result, *dayMap[day])
	}
	return result
}

func getScore(event models.Event) int {
	ret := 0
	for _, j := range event.EventAdminJudgements {
		if j.IsApproval {
			ret++
		} else {
			ret--
		}
	}
	return ret
}

func formatEventTime(event models.Event) string {
	if event.DoorsAt != nil && *event.DoorsAt != event.StartAt {
		return "Doors " + event.DoorsAt.In(util.TZLocation).Format("3:04pm") + ", Start " + event.StartAt.In(util.TZLocation).Format("3:04pm")
	}
	return event.StartAt.In(util.TZLocation).Format("3:04pm")
}

func getEventTags(event models.Event) []string {
	tags := make([]string, 0, len(event.Tags))
	for _, tag := range event.Tags {
		tags = append(tags, tag.Name)
	}
	return tags
}

func (h *APIHandler) handleAdminAddEvent(w http.ResponseWriter, r *http.Request) {
	user := h.getAuthenticatedUser(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req AdminAddEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

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

	var event models.Event
	event.Title = req.Title
	event.Link = normalizeURL(req.Link)
	event.CoverCost = req.CoverCost
	event.Venue.Name = req.Venue
	event.Region.Name = req.Region
	event.Age.Name = req.Age

	startAt, err := time.ParseInLocation(formDTFormat, fmt.Sprintf("%s %s", req.Date, req.StartTime), util.TZLocation)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid date or start time format")
		return
	}
	event.StartAt = startAt

	if req.DoorsTime != "" {
		doorsAt, err := time.ParseInLocation(formDTFormat, fmt.Sprintf("%s %s", req.Date, req.DoorsTime), util.TZLocation)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid doors time format")
			return
		}
		event.DoorsAt = &doorsAt
	}

	for _, tag := range req.Tags {
		event.Tags = append(event.Tags, models.Tag{Name: tag})
	}

	now := time.Now().Unix()
	event.VerifiedAt = &now
	event.ApprovedAt = &now
	event.UpdateGuid(nil)

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

	h.writeJSON(w, http.StatusOK, AdminAddEventResponse{Success: true, Message: "Event created successfully."})
}

func (h *APIHandler) handleAdminGetAges(w http.ResponseWriter, r *http.Request) {
	user := h.getAuthenticatedUser(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	tx := h.database.DB.Begin()
	defer tx.Commit()

	ageModels, err := h.database.GetAges(tx)
	if err != nil {
		h.logger.Errorf("failed to get ages: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get ages")
		return
	}

	ages := make([]AgeResponse, 0, len(ageModels))
	for _, age := range ageModels {
		ages = append(ages, AgeResponse{ID: int64(age.ID), Name: age.Name, Value: age.Value})
	}
	h.writeJSON(w, http.StatusOK, ages)
}

func (h *APIHandler) handleAdminUpdateEvent(w http.ResponseWriter, r *http.Request) {
	user := h.getAuthenticatedUser(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req AdminUpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
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

	updated := false
	if req.Title != nil && *req.Title != event.Title {
		event.Title = *req.Title
		updated = true
	}
	if req.Link != nil && *req.Link != event.Link {
		event.Link = normalizeURL(*req.Link)
		updated = true
	}
	if req.CoverCost != nil && *req.CoverCost != event.CoverCost {
		event.CoverCost = *req.CoverCost
		updated = true
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
		newStart, err := time.ParseInLocation(formDTFormat, fmt.Sprintf("%s %s", *req.Date, timeStr), util.TZLocation)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid date format")
			return
		}
		if !newStart.Equal(event.StartAt) {
			if event.DoorsAt != nil {
				doorsTimeStr := event.DoorsAt.In(util.TZLocation).Format("15:04")
				if newDoors, err := time.ParseInLocation(formDTFormat, fmt.Sprintf("%s %s", *req.Date, doorsTimeStr), util.TZLocation); err == nil {
					event.DoorsAt = &newDoors
				}
			}
			event.StartAt = newStart
			updated = true
		}
	}
	if req.StartTime != nil {
		dateStr := event.StartAt.In(util.TZLocation).Format("2006-01-02")
		newStart, err := time.ParseInLocation(formDTFormat, fmt.Sprintf("%s %s", dateStr, *req.StartTime), util.TZLocation)
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
			newDoors, err := time.ParseInLocation(formDTFormat, fmt.Sprintf("%s %s", dateStr, *req.DoorsTime), util.TZLocation)
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
		h.writeJSON(w, http.StatusOK, AdminUpdateEventResponse{Success: true, Message: "No changes detected."})
		return
	}

	if _, err := h.database.UpdateEvent(event, tx); err != nil {
		h.logger.Errorf("failed to save updated event: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to update event")
		return
	}

	h.writeJSON(w, http.StatusOK, AdminUpdateEventResponse{Success: true, Message: "Event updated successfully."})
}
