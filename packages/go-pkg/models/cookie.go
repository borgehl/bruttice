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

func (c *CookieModel) UpdateCookie(cookie SessionCookie) error {
	body, err := json.Marshal(cookie)
	if err != nil {
		return err
	}
	c.CookieBody = body
	return nil
}

func (c *CookieModel) GetCookie() (SessionCookie, error) {
	ret := SessionCookie{}
	err := json.Unmarshal(c.CookieBody, &ret)
	return ret, err
}
