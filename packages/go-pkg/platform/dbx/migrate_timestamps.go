package dbx

import "gorm.io/gorm"

const timestampMigrationKey = "timestamp_utc_fix"

// MigrateTimestamps reinterprets existing event timestamps (stored as fake-UTC
// Central time values) as actual UTC. Idempotent via migration_flags table.
func (d *DatabaseClient) MigrateTimestamps(tx *gorm.DB) error {
	flag, err := d.GetMigrationFlag(timestampMigrationKey, tx)
	if err != nil {
		return err
	}
	if flag.ID != 0 {
		d.Logger.Infof("timestamp migration already applied, skipping")
		return nil
	}

	d.Logger.Infof("applying timestamp UTC migration...")

	if err := tx.Exec(`
		UPDATE events
		SET start_at = (start_at::timestamp AT TIME ZONE 'America/Chicago')
	`).Error; err != nil {
		return err
	}

	if err := tx.Exec(`
		UPDATE events
		SET doors_at = (doors_at::timestamp AT TIME ZONE 'America/Chicago')
		WHERE doors_at IS NOT NULL
	`).Error; err != nil {
		return err
	}

	if err := d.CreateMigrationFlag(timestampMigrationKey, tx); err != nil {
		return err
	}

	d.Logger.Infof("timestamp UTC migration applied successfully")
	return nil
}
