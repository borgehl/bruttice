package controlcookie

import (
	"github.com/borgehl/bruttice/apps/backend/internal/web/controlctx"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/dbx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"
)

const UserSessionCookieName = "user-session-cookie"
const expirationSeconds = 30 * 60

var (
	ErrInvalidCookie  = fmt.Errorf("invalid cookie")
	ErrCookieNotFound = fmt.Errorf("cookie not found")
)

type CookieStore struct {
	database       *dbx.DatabaseClient
	logger         *logx.Logger
	SessionCookies map[string]struct{}
	cookieSalt     string
}

func NewCookieStore(database *dbx.DatabaseClient, cookieSalt string) *CookieStore {
	return &CookieStore{
		database:       database,
		logger:         database.Logger,
		SessionCookies: make(map[string]struct{}),
		cookieSalt:     cookieSalt,
	}
}

func getCookie(r *http.Request, cookieName string) *http.Cookie {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil
		}
		panic(err)
	}
	return cookie
}

func WriteCookie(w http.ResponseWriter, cookie *http.Cookie) {
	http.SetCookie(w, cookie)
}

func getNewExpiration() time.Time {
	return time.Now().Add(expirationSeconds * time.Second)
}

func (c *CookieStore) cookieSignature(name string, value string) []byte {
	mac := hmac.New(sha256.New, []byte(c.cookieSalt))
	mac.Write([]byte(name))
	mac.Write([]byte(value))
	return mac.Sum(nil)
}

func RefreshCookieExpiration(cookie *http.Cookie) {
	cookie.Expires = getNewExpiration()
}

func (c *CookieStore) IsValidCookie(cookie *http.Cookie) bool {
	if cookie == nil {
		c.logger.Errorf("nil cookie")
		return false
	}
	if time.Now().After(cookie.Expires) {
		c.logger.Errorf("expired cookie, at %+v", cookie.Expires)
		return false
	}
	decodedValueByte, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		c.logger.Errorf("couldn't decode cookie")
		return false
	}
	decodedValueString := string(decodedValueByte)
	if len(decodedValueString) < sha256.Size {
		c.logger.Errorf("decoded cookie not long enough")
		return false
	}
	signature := decodedValueString[:sha256.Size]
	value := decodedValueString[sha256.Size:]
	calculatedSignature := c.cookieSignature(cookie.Name, value)
	ret := hmac.Equal([]byte(signature), calculatedSignature)
	if !ret {
		c.logger.Errorf("cookie signature invalid")
	}
	return ret
}

func (c *CookieStore) MakeCookie(name string, value string) http.Cookie {
	ret := http.Cookie{
		Name:     name,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  getNewExpiration(),
	}
	ret.Value = base64.URLEncoding.EncodeToString(
		[]byte(string(c.cookieSignature(ret.Name, value)) + value),
	)
	return ret
}

func (c *CookieStore) NewUserSessionCookie(user models.User, tx *gorm.DB) (http.Cookie, error) {
	ret := c.MakeCookie(UserSessionCookieName, user.Email)
	_, err := c.database.SaveCookie(models.SessionCookie{UserID: user.ID, Cookie: ret}, tx)
	return ret, err
}

func (c *CookieStore) CreateSessionCookie(user models.User, tx *gorm.DB) *http.Cookie {
	cookie, err := c.NewUserSessionCookie(user, tx)
	if err != nil {
		c.logger.Errorf("failed to create session cookie: %v", err)
		return nil
	}
	return &cookie
}

func (c *CookieStore) GetUserFromRequest(r *http.Request) *models.User {
	return controlctx.GetUserFromRequestContext(r)
}

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
