package mapper

import (
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

func AgeResponses(ages []models.Age) []dto.AgeResponse {
	ret := make([]dto.AgeResponse, 0, len(ages))
	for _, age := range ages {
		ret = append(ret, dto.AgeResponse{ID: int64(age.ID), Name: age.Name, Value: age.Value})
	}
	return ret
}
