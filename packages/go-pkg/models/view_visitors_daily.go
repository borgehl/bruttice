package models

import "time"

type VisitorsDaily struct {
	ViewBase

	NumVisitors int64
	VisitDate   time.Time
}

func (a *VisitorsDaily) CreateView() string {
	viewSql := `
		with base as (
			select
				visitor_id,
				to_timestamp(created_at)::date as visit_date
			from visitors_view
		)

		select
			count(distinct visitor_id) as num_visitors,
			visit_date
		from base
		group by visit_date
			order by visit_date
	`
	return a.BuildSql(viewSql, a)
}
