package notifications

import (
	"github.com/borgehl/bruttice/packages/go-pkg/configs"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/dbx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/emailx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
	"fmt"
	"time"
)

func GoRunSendApprovalNotifications() {
	for {
		SendApprovalNotification()
		now := time.Now()
		next := now.Add(time.Hour)
		time.Sleep(next.Sub(now))
	}
}

func SendApprovalNotification() {
	log := logx.NewLogger(nil, nil)
	db := dbx.NewDatabaseClient(log)
	emailClient := emailx.NewEmailClient(log)

	tx := db.DB.Begin()
	defer tx.Commit()

	events, err := db.GetAwaitingApprovalEvents(tx)
	if err != nil {
		emailClient.SendEmail(
			configs.RootUser.Email,
			emailClient.GetGenericEmail(
				"Failed to get awaiting approval events",
				fmt.Sprintf("Received database error: %s", err),
			),
		)
		return
	}

	if len(events) == 0 {
		log.Info("No events awaiting approval.")
		return
	}

	msg := "Events awaiting approval:\r\n"
	for _, event := range events {
		msg += fmt.Sprintf("%s at %s by %s\r\n", event.Title, event.Venue.Name, *event.VerificationEmail)
	}

	emailClient.SendEmail(
		configs.RootUser.Email,
		emailClient.GetGenericEmail(
			fmt.Sprintf("Events awaiting approval: %d", len(events)),
			msg,
		),
	)
}
