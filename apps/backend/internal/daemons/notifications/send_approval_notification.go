package notifications

import (
	"fmt"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
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
		notify(log, db, emailClient, models.SettingNotifySystemErrors, emailClient.GenericMessage(
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

	notify(log, db, emailClient, models.SettingNotifyNewEvents, emailClient.GenericMessage(
		fmt.Sprintf("Events awaiting approval: %d", len(events)),
		msg,
	))
}

// notify sends a message to the admins who chose the notification (a models.SettingNotify key).
// If nobody did, or the lookup fails, it is not sent and a line says why.
func notify(log *logx.Logger, db *dbx.DatabaseClient, emailClient *emailx.EmailClient, kind string, msg emailx.Message) {
	to, err := db.NotificationRecipients(kind)
	if err != nil {
		log.Errorf("not sending %q: failed to look up who gets %s: %v", msg.Subject, kind, err)
		return
	}
	if len(to) == 0 {
		log.Infof("not sending %q: no admin chose %s", msg.Subject, kind)
		return
	}
	msg.To = to
	_, _ = emailClient.Send(msg)
}
