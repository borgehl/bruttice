package emailx

import (
	"fmt"
	"net/smtp"
	"os"
	"strconv"

	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
)

type EmailClient struct {
	Host string
	// mail server credentials
	MailServer    string
	FromEmail     string
	EmailPassword string
	MailPort      int
	SendAddr      string

	logger *logx.Logger
}

func NewEmailClient(logger *logx.Logger) EmailClient {
	mailPortStr := os.Getenv("MAIL_PORT")
	if mailPortStr == "" {
		panic(fmt.Errorf("no mail port sent for sending emails"))
	}
	mailPort, err := strconv.Atoi(mailPortStr)
	if err != nil {
		panic(err)
	}
	ret := EmailClient{
		Host:          os.Getenv("HOST"),
		MailServer:    os.Getenv("MAIL_SERVER"),
		FromEmail:     os.Getenv("MAIL_EMAIL"),
		EmailPassword: os.Getenv("MAIL_PASSWORD"),
		MailPort:      mailPort,
		logger:        logger,
	}
	ret.SendAddr = fmt.Sprintf("%s:%d", ret.MailServer, ret.MailPort)
	return ret
}

func (e *EmailClient) SendEmail(emailAddress string, msg []byte) {
	if e.MailServer == "" {
		e.logger.Infof("MAIL_SERVER is not set, not sending email to %s", emailAddress)
		return
	}

	// auth
	auth := smtp.PlainAuth("", e.FromEmail, e.EmailPassword, e.MailServer)
	if err := smtp.SendMail(e.SendAddr, auth, e.FromEmail, []string{emailAddress}, msg); err != nil {
		e.logger.Errorf("failed to send email to %s: %s", emailAddress, err)
		return
	}
	e.logger.Infof("sent the email! %s", e.SendAddr)
}
