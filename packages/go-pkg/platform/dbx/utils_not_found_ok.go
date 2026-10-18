package dbx

import (
	"errors"

	"gorm.io/gorm"
)

// false if
//   - err is nil
//   - is a record not found error from gorm
//
// maybe not helpful... se get artist logic
func isErrorNotFoundOK(err error) bool {
	if err == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	}
	return true
}
