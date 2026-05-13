package models

type VisitorsMonthly struct {
	ViewBase

	NumVisitors int64
	VisitYear   int
	VisitMonth  int
}

func (a *VisitorsMonthly) CreateView() string {
	viewSql := `
		with base as (
			select
				visitor_id,
				extract(year from to_timestamp(created_at)) as visit_year,
				extract(month from to_timestamp(created_at)) as visit_month
			from visitors_view
		)

		select
			count(distinct visitor_id) as num_visitors,
			visit_year,
			visit_month
		from base
		group by visit_year, visit_month
		order by visit_year, visit_month
	`
	return a.BuildSql(viewSql, a)
}
