package models

import (
	"encoding/json"
	"net/http"
)

type CookieModel struct {
	BaseModel

	Value      string `gorm:"unique"`
	CookieBody []byte
}

type SessionCookie struct {
	UserID int64
	http.Cookie
}

// update the cookie model with the new cookie data
func (c *CookieModel) UpdateCookie(cookie SessionCookie) error {
	body, err := json.Marshal(cookie)
	if err != nil {
		return err
	}
	c.CookieBody = body
	return nil
}

// unmarshal the stored cookie in the model body
func (c *CookieModel) GetCookie() (SessionCookie, error) {
	ret := SessionCookie{}
	err := json.Unmarshal(c.CookieBody, &ret)
	return ret, err
}
