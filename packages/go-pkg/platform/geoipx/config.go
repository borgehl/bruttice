package geoipx

import (
	"os"
)

type GeoIpConfig struct {
	IpGeolocationIoApiKey string
}

func NewGeoIpConfig() GeoIpConfig {
	return GeoIpConfig{
		IpGeolocationIoApiKey: os.Getenv("IP_GEOLOCATION_IO_API_KEY"),
	}
}
