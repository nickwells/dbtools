package main

// dbt_make_dirs

import (
	"regexp"

	"github.com/nickwells/check.mod/v2/check"
	"github.com/nickwells/dbtools/internal/dbtcommon"
	"github.com/nickwells/param.mod/v5/param"
	"github.com/nickwells/param.mod/v5/param/psetter"
)

func addParams(prog *Prog) param.PSetOptFunc {
	return func(ps *param.PSet) error {
		dbtcommon.AddParamDBName(prog.dbp, ps)

		ps.Add("only-check", psetter.Bool{Value: &prog.onlyCheck},
			"only check if the directories are present - don't create them."+
				" If this is set then the exit status will be set to 1 if"+
				" the directories are not all present and to zero otherwise")

		ps.Add("schema-names",
			psetter.StrList{
				Value: &prog.schemaNames,
				Checks: []check.StringSlice{
					check.SliceAll[[]string](
						check.StringMatchesPattern[string](
							regexp.MustCompile(`[a-z][a-z0-9_]*`),
							"a schema name: a leading lowercase character"+
								" followed by zero or more lowercase"+
								" letters, digits or underscores")),
					check.SliceLength[[]string](check.ValGT(0)),
					check.SliceHasNoDups[[]string, string],
				},
			},
			"a list of schemas to create the directories for",
			param.AltNames("schemas"),
			param.Attrs(param.MustBeSet))

		return nil
	}
}
