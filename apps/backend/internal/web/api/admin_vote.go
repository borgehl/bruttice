package api

import (
	"fmt"
	"github.com/borgehl/bruttice/packages/go-pkg/models"

	"gorm.io/gorm"
)

func getIsApproval(vote string) int {
	switch vote {
	case "Approve", "Keep":
		return 1
	case "Deny", "Remove":
		return -1
	}
	return 0
}

func (h *APIHandler) processVote(vote string, guid string, user models.User, tx *gorm.DB) error {
	event, err := h.database.GetEventByGuid(guid, tx)
	if err != nil {
		return err
	}
	if event.ID == 0 {
		return fmt.Errorf("event for guid %s not found", guid)
	}

	isApproval := getIsApproval(vote)
	if isApproval == 0 {
		return fmt.Errorf("invalid vote of %s", vote)
	}

	judgement := models.EventAdminJudgement{UserID: user.ID, EventID: event.ID}
	judgement, err = h.database.GetJudgement(judgement, tx)
	if err != nil {
		return err
	}

	judgement.IsApproval = (isApproval == 1)
	if _, err = h.database.SaveJudgement(judgement, tx); err != nil {
		return err
	}

	// Check if we should approve or delete the event based on score
	event, err = h.database.GetEventByGuid(guid, tx)
	if err != nil {
		return err
	}

	score := event.Score()
	if score > 0 && event.ApprovedAt == nil {
		if err := h.database.ApproveEvent(event, tx); err != nil {
			return err
		}
	} else if score < 0 {
		if err := tx.Delete(&event).Error; err != nil {
			return err
		}
	}

	return nil
}
