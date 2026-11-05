package dbx

import "gorm.io/gorm"

const dedupEventsSQL = `
with
	delete_ids as (
		select
			distinct least(e1.id, e2.id) as id
			from events as e1
			inner join events e2
				on
					e1.venue_id = e2.venue_id
					and e1.venue_id in (?)
					and e1.start_at::date = e2.start_at::date
					and e1.id != e2.id
					and (
						(
							e1.title ilike '%' || e2.title || '%'
							or e2.title ilike '%' || e1.title || '%'
						)
						or (
							abs(extract(epoch from
								(e1.start_at - e2.start_at)
							)) / 3600 < 1
						)
					)
			where
				e1.deleted_at is null and e2.deleted_at is null
				and e1.cancelled_at is null and e2.cancelled_at is null
	)
	update events
	set deleted_at = now()
	where id in (select id from delete_ids)
`

// DeduplicateEvents soft-deletes duplicate events, considering only events
// belonging to the given venues. Cancelled events are left out: one kept next to the
// show that replaced it, or the older of two, must not take the live event with it.
func (d *DatabaseClient) DeduplicateEvents(venueIDs []int64, tx *gorm.DB) error {
	if len(venueIDs) == 0 {
		return nil
	}
	for {
		r := tx.Exec(dedupEventsSQL, venueIDs)
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected == 0 {
			break
		}
	}
	return nil
}
