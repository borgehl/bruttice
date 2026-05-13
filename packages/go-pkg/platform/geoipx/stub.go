package geoipx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
	"context"
	"fmt"
	"os"

	ipgeolocation "github.com/IPGeolocation/ip-geolocation-go-sdk/sdk"
)

var ErrClientLocked = fmt.Errorf("client locked")

type GeoLocation struct {
	IpAddr    string
	StateCode string
	City      string
	Country   string
}

type SdkInterface interface {
	GetLocation(string) (GeoLocation, error)
	IsLocked() bool
}

type GeoIpConfig struct {
	IpGeolocationIoApiKey string
}

func NewGeoIpConfig() GeoIpConfig {
	return GeoIpConfig{
		IpGeolocationIoApiKey: os.Getenv("IP_GEOLOCATION_IO_API_KEY"),
	}
}

const (
	includeFields   = "location.state_code, location.city, location.country_code2"
	language        = "en"
	output          = "json"
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

type GeoIpClient struct {
	clients []SdkInterface
}

func NewGeoIpClient(log *logx.Logger) *GeoIpClient {
	config := NewGeoIpConfig()
	log.Infof("Config: %+v\n", config)
	ret := GeoIpClient{}
	ret.clients = append(ret.clients, NewIPGeolocationIoClient(config))
	return &ret
}

func (c *GeoIpClient) GetLocation(ipAddr string) (ret GeoLocation, err error) {
	for _, client := range c.clients {
		if client.IsLocked() {
			continue
		}
		ret, err = client.GetLocation(ipAddr)
		if err == nil {
			break
		}
	}
	return
}
