package notifications

import (
	"fmt"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/dbx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/emailx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
	"time"
)

// sencd approval notes every hour
func GoRunSendApprovalNotifications(emailClient *emailx.EmailClient) {
	for {
		SendApprovalNotification(emailClient)
		now := time.Now()
		next := now.Add(time.Hour)
		time.Sleep(next.Sub(now))
	}
}

func SendApprovalNotification(emailClient *emailx.EmailClient) {
	log := logx.NewLogger(nil, nil)
	db := dbx.NewDatabaseClient(log)

	tx := db.DB.Begin()
	defer tx.Commit()

	events, err := db.GetAwaitingApprovalEvents(tx)
	if err != nil {
		_, _ = emailClient.Send(emailClient.GenericMessage(
			"Failed to get awaiting approval events",
			fmt.Sprintf("Received database error: %s", err),
		))
		return
	}

	if len(events) == 0 {
		log.Info("No events awaiting approval.")
		return
	}

	msg := "Events awaiting approval:\n"
	for _, event := range events {
		msg += fmt.Sprintf("%s at %s by %s\n", event.Title, event.Venue.Name, *event.VerificationEmail)
	}

	_, _ = emailClient.Send(emailClient.GenericMessage(
		fmt.Sprintf("Events awaiting approval: %d", len(events)),
		msg,
	))
}
