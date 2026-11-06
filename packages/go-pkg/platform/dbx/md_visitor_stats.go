package dbx

import "time"

// VisitorDayCount is the number of distinct visitors on one day.
type VisitorDayCount struct {
	VisitDate   time.Time
	NumVisitors int64
}

// VisitorLocationCount is the number of distinct visitors from one place.
// Visitors whose location could not be found have empty fields.
type VisitorLocationCount struct {
	Country     string
	State       string
	City        string
	NumVisitors int64
}

// GetVisitorsPerDay returns one row for each of the last `days` days, today
// included, with 0 for the days nobody visited.
func (d *DatabaseClient) GetVisitorsPerDay(days int) ([]VisitorDayCount, error) {
	ret := make([]VisitorDayCount, 0, days)
	err := d.DB.Raw(`
		select
			series.day::date as visit_date,
			coalesce(daily.num_visitors, 0) as num_visitors
		from generate_series(
			current_date - (cast(? as int) - 1),
			current_date,
			interval '1 day'
		) as series(day)
		left join visitors_daily_view as daily
			on daily.visit_date = series.day::date
		order by visit_date
	`, days).Scan(&ret).Error
	return ret, err
}

// GetVisitorsPerLocation returns the busiest places over the last `days` days,
// today included. A visitor who came on several days is counted once.
func (d *DatabaseClient) GetVisitorsPerLocation(days, limit int) ([]VisitorLocationCount, error) {
	ret := make([]VisitorLocationCount, 0)
	err := d.DB.Raw(`
		select
			coalesce(country, '') as country,
			coalesce(state, '') as state,
			coalesce(city, '') as city,
			count(distinct visitor_id) as num_visitors
		from visitors_view
		where to_timestamp(created_at)::date > current_date - cast(? as int)
		group by 1, 2, 3
		order by num_visitors desc, country, state, city
		limit ?
	`, days, limit).Scan(&ret).Error
	return ret, err
}
