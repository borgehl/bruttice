package models

import (
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"fmt"
)

type ViewBase struct{}

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
