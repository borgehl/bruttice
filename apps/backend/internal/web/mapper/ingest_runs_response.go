package mapper

import (
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

// IngestRunResponses converts runs, totalling the counts of each run's sources.
func IngestRunResponses(runs []models.IngestRun) []dto.IngestRunResponse {
	ret := make([]dto.IngestRunResponse, 0, len(runs))
	for _, run := range runs {
		resp := dto.IngestRunResponse{
			ID:          run.ID,
			StartedAt:   run.StartedAt,
			FinishedAt:  run.FinishedAt,
			DedupeError: run.DedupeError,
			Sources:     make([]dto.IngestSourceResponse, 0, len(run.Sources)),
		}
		for _, s := range run.Sources {
			resp.Scraped += s.ScrapedCount
			resp.Saved += s.SavedCount
			resp.Failed += s.FailedCount
			errs := s.Errors
			if errs == nil {
				errs = []string{}
			}
			resp.Sources = append(resp.Sources, dto.IngestSourceResponse{
				Source:       s.Source,
				Status:       s.Status,
				ScrapedCount: s.ScrapedCount,
				SavedCount:   s.SavedCount,
				FailedCount:  s.FailedCount,
				Errors:       errs,
			})
		}
		ret = append(ret, resp)
	}
	return ret
}
