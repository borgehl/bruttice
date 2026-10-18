package geoipx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
)

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
