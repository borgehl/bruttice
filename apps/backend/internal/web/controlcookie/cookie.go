package controlcookie

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"time"
)

// 30 minutes of inactivity will expire the session
const expirationSeconds = 30 * 60

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

// Reset expiration to start from now
func RefreshCookieExpiration(cookie *http.Cookie) {
	cookie.Expires = getNewExpiration()
}

// Checks cookie for
//   - expiration
//   - signature
//
// Return false for any reasons the cookie is invalid including
// errors in parsing cookie.Value
func (c *CookieStore) IsValidCookie(cookie *http.Cookie) bool {
	if cookie == nil {
		c.logger.Errorf("nil cookie")
		return false
	}
	// expired
	if time.Now().After(cookie.Expires) {
		c.logger.Errorf("expired cookie, at %+v", cookie.Expires)
		return false
	}
	// false if cannot decode
	decodedValueByte, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		c.logger.Errorf("couldn't decode cookie")
		return false
	}
	// false if cookie not long enough
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
