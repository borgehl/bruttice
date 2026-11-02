package geoipx

import (
	"context"

	ipgeolocation "github.com/IPGeolocation/ip-geolocation-go-sdk/sdk"
)

const (
	includeFields = "location.state_code, location.city, location.country_code2"
	language      = "en"
	output        = "json"
	// hit the 1k limit for the day, when does it reset?
	tooManyRequests = 429
)

type IpGeolocationIoClient struct {
	apiKey string
	client *ipgeolocation.APIClient
	ctx    context.Context
	Locked bool
}

func NewIPGeolocationIoClient(config GeoIpConfig) *IpGeolocationIoClient {
	return &IpGeolocationIoClient{
		Locked: false,
		apiKey: config.IpGeolocationIoApiKey,
		client: ipgeolocation.NewAPIClient(ipgeolocation.NewConfiguration()),
		ctx:    ipgeolocation.WithAPIKey(context.Background(), config.IpGeolocationIoApiKey),
	}
}

func (c *IpGeolocationIoClient) IsLocked() bool {
	return c.Locked
}

func (c *IpGeolocationIoClient) GetLocation(ipAddr string) (ret GeoLocation, err error) {
	ret.IpAddr = ipAddr
	resp, r, err := c.client.IPGeolocationAPI.GetIpGeolocation(c.ctx).Ip(ipAddr).Lang(language).Fields(includeFields).Output(output).Execute()
	if err != nil {
		return
	}
	if r.StatusCode == tooManyRequests {
		c.Locked = true
		err = ErrClientLocked
		return
	}
	ret.Country = *resp.Location.CountryCode2
	ret.StateCode = *resp.Location.StateCode
	ret.City = *resp.Location.City
	return
}
