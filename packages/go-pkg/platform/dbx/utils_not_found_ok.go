package dbx

import (
	"errors"

	"gorm.io/gorm"
)

func isErrorNotFoundOK(err error) bool {
	if err == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	}
	return true
}
