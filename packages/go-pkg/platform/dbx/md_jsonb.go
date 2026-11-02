package dbx

import (
	"encoding/json"

	"gorm.io/gorm"
)

// MergeJSONB merges patch into the jsonb column of one row as a single statement, so keys absent
// from patch are never read or rewritten and a concurrent write of another key survives. The
// merge is shallow: a key in patch replaces the whole value under that key, and no key is ever
// removed (write the zero value instead). It also touches updated_at, and scans the column as it
// is after the merge into dest, which fails with sql.ErrNoRows if there is no such row. The column
// must not be null: a null merged with an object stays null. table and column are repository
// constants, never request input.
func MergeJSONB(db *gorm.DB, table, column string, id int64, patch map[string]any, dest any) error {
	raw, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	return db.Raw(
		"UPDATE "+table+" SET "+column+" = "+column+" || ?::jsonb, updated_at = extract(epoch from now())::bigint "+
			"WHERE id = ? AND deleted_at IS NULL RETURNING "+column,
		string(raw), id,
	).Row().Scan(dest)
}
