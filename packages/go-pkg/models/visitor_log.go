package models

import "net/http"

type VisitorLog struct {
	BaseModel

	VisitorID string
	Hostname  string
	PathName  string
	FilterKey string

	// raw query string
	QueryParams string

	VisitorLocationID int64
	VisitorLocation   VisitorLocation
}

func NewAccessLog(r *http.Request, visitorID string) (ret VisitorLog) {
	ret.VisitorID = visitorID
	ret.Hostname = r.Host
	ret.PathName = r.URL.Path
	ret.QueryParams = r.URL.RawQuery
	ret.FilterKey = r.URL.Query().Get("filterKey")
	return ret
}
