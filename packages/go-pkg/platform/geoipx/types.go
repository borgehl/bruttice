package geoipx

import (
	"fmt"
)

var ErrClientLocked = fmt.Errorf("client locked")

type GeoLocation struct {
	IpAddr string
	// ex. "US-CA"
	StateCode string
	City      string
	Country   string
}

type SdkInterface interface {
	GetLocation(string) (GeoLocation, error)
	IsLocked() bool
}
