package ingest

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"fmt"
)

type VenueInfo struct {
	Name          string `json:"name"`
	Link          string `json:"link"`
	IsAutoIngest  bool   `json:"is_auto_ingest"`
	DefaultRegion string `json:"default_region"`
	SkipVenueSave bool   `json:"skip_venue_save"`
	Timezone      string `json:"timezone"`
}

type IngestionOutput struct {
	VenueInfo VenueInfo      `json:"venue_info"`
	Events    []models.Event `json:"events"`
	Errors    []string       `json:"errors"`
}

func (io *IngestionOutput) AppendErrors(errs []error) {
	for _, err := range errs {
		io.Errors = append(io.Errors, fmt.Sprintf("%s: %s", io.VenueInfo.Name, err))
	}
}

func (io *IngestionOutput) AppendError(err error) {
	if err == nil {
		return
	}
	io.Errors = append(io.Errors, fmt.Sprintf("%s: %s", io.VenueInfo.Name, err))
}

func (io *IngestionOutput) GetErrors() []error {
	ret := make([]error, 0, len(io.Errors))
	for _, e := range io.Errors {
		ret = append(ret, fmt.Errorf("%s", e))
	}
	return ret
}
