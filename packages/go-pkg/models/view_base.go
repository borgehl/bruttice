package models

import (
	"fmt"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
)

type ViewBase struct{}

// drop and create view sql
//
// viewSql indentino expected to be two tab widths for pretty display
func (vb *ViewBase) BuildSql(viewSql string, v any) string {
	viewName := util.StructToSnakeCase(v) + "_view"
	ret := fmt.Sprintf("drop view if exists %s cascade ;\n", viewName)
	ret += fmt.Sprintf(
		`create view %s as (

%s

		);`,
		viewName,
		viewSql,
	)
	return ret
}
