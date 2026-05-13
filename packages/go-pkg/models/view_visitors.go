package models

type Visitors struct {
	ViewBase

	VisitorID string
	PathName  string
	City      string
	State     string
	Country   string
}

func (a *Visitors) CreateView() string {
	viewSql := `
		select
			log.created_at,
			log.visitor_id,
			log.path_name,
			loc.city,
			split_part(loc.state_code, '-', 2) as state,
			loc.country
		from visitor_logs as log
		inner join visitor_locations as loc
			on log.visitor_location_id = loc.id
	`
	return a.BuildSql(viewSql, a)
}
