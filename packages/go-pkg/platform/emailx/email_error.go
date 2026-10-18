package emailx

import (
	"fmt"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/configs"
)

const errDateFmt = "Mon Jan _2"

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
