package emailx

import (
	"fmt"
	"strings"
)

const (
	secureProtocol   = "https://"
	unsecureProtocol = "http://"
)

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
