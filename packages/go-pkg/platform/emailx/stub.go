package emailx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/configs"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

const errDateFmt = "Mon Jan _2"

const (
	secureProtocol   = "https://"
	unsecureProtocol = "http://"
)

type EmailClient struct {
	Host          string
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
	auth := smtp.PlainAuth("", e.FromEmail, e.EmailPassword, e.MailServer)
	if err := smtp.SendMail(e.SendAddr, auth, e.FromEmail, []string{emailAddress}, msg); err != nil {
		e.logger.Errorf("failed to send email to %s: %s", emailAddress, err)
		return
	}
	e.logger.Infof("sent the email! %s", e.SendAddr)
}

func (e *EmailClient) GetErrorsEmail(errs []error) []byte {
	rootEmail := configs.RootUser.Email
	ret := fmt.Sprintf("To: %s\r\n", rootEmail)
	ret += fmt.Sprintf("From: %s\r\n", e.FromEmail)
	ret += fmt.Sprintf("Subject: %s Scraping Errors: %d\r\n\r\n", time.Now().Format(errDateFmt), len(errs))
	ret += "Please see recorded errors below\r\n"
	for i, err := range errs {
		ret += fmt.Sprintf("%d: %s\r\n", i, err)
	}
	return []byte(ret)
}

func (e *EmailClient) GetGenericEmail(subject string, msg string) []byte {
	rootEmail := configs.RootUser.Email
	ret := fmt.Sprintf("To: %s\r\n", rootEmail)
	ret += fmt.Sprintf("From: %s\r\n", e.FromEmail)
	ret += fmt.Sprintf("Subject: %s\r\n\r\n", subject)
	ret += fmt.Sprintf("%s\r\n", msg)
	return []byte(ret)
}

func (e *EmailClient) GetVerifyEventEmail(title string, email string, verificationCode string) []byte {
	host := strings.TrimPrefix(strings.TrimPrefix(e.Host, secureProtocol), unsecureProtocol)
	if strings.HasPrefix(e.Host, secureProtocol) {
		host = secureProtocol + host
	} else {
		host = unsecureProtocol + host
	}
	return []byte(
		fmt.Sprintf("To: %s\r\n", email) +
			fmt.Sprintf("From: %s\r\n", e.FromEmail) +
			"Subject: Please verify your event\r\n" +
			"\r\n" +
			fmt.Sprintf("Hello Author of %s!\r\n", title) +
			"Thanks for posting your event.\r\n" +
			"Please click on the following link to verify your email address:\r\n" +
			fmt.Sprintf("%s/verify-event?verifyKey=%s\r\n", host, verificationCode),
	)
}
