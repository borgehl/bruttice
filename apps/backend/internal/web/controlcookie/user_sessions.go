package controlcookie

import (
	"fmt"
	"net/http"

	"github.com/borgehl/bruttice/apps/backend/internal/web/controlctx"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"gorm.io/gorm"
)

const UserSessionCookieName = "user-session-cookie"

var (
	ErrInvalidCookie  = fmt.Errorf("invalid cookie")
	ErrCookieNotFound = fmt.Errorf("cookie not found")
)

func (c *CookieStore) NewUserSessionCookie(user models.User, tx *gorm.DB) (http.Cookie, error) {
	ret := c.MakeCookie(UserSessionCookieName, user.Email)
	_, err := c.database.SaveCookie(models.SessionCookie{UserID: user.ID, Cookie: ret}, tx)
	return ret, err
}

// CreateSessionCookie creates a new session cookie for the user and returns a pointer to it
func (c *CookieStore) CreateSessionCookie(user models.User, tx *gorm.DB) *http.Cookie {
	cookie, err := c.NewUserSessionCookie(user, tx)
	if err != nil {
		c.logger.Errorf("failed to create session cookie: %v", err)
		return nil
	}
	return &cookie
}

// GetUserFromRequest gets the user from the request context after RequestWithCookieUser was called
func (c *CookieStore) GetUserFromRequest(r *http.Request) *models.User {
	return controlctx.GetUserFromRequestContext(r)
}

// gets cookie from the request
// checks existence and validity
func (c *CookieStore) RequestWithCookieUser(r *http.Request, tx *gorm.DB) (*http.Request, error) {
	cookie := getCookie(r, UserSessionCookieName)
	if cookie == nil {
		return nil, ErrCookieNotFound
	}
	cookieModel, err := c.database.GetCookieByValue(cookie.Value, tx)
	if err != nil {
		return nil, err
	}
	if cookieModel.ID == 0 {
		return nil, ErrCookieNotFound
	}
	sessionCookie, err := cookieModel.GetCookie()
	if err != nil {
		return nil, err
	}
	if !c.IsValidCookie(&sessionCookie.Cookie) {
		c.logger.Errorf("invalid cookie %+v", sessionCookie.Cookie)
		return nil, ErrInvalidCookie
	}
	user := models.User{BaseModel: models.BaseModel{ID: sessionCookie.UserID}}
	user, err = c.database.GetUser(user, tx)
	if err != nil {
		return nil, err
	}
	return controlctx.RequestWithUser(r, &user), nil
}
