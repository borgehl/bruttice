package passwordsx

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

var PwCost = bcrypt.DefaultCost

func init() {
	pw := []byte("nice and secret Passw00rd Long")
	for {
		ts := time.Now()
		if _, err := bcrypt.GenerateFromPassword(pw, PwCost); err != nil {
			panic(err)
		}
		if float64(time.Since(ts)) >= 1000.0 {
			break
		}
		PwCost++
	}
}

func GetPasswordHash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), PwCost)
}

func ValidPassword(password string, hashedPassword []byte) bool {
	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)) == nil
}
