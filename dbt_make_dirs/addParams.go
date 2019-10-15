// dbt_make_dirs
package main

import (
	"regexp"

	"github.com/nickwells/check.mod/check"
	"github.com/nickwells/dbtcommon.mod/dbtcommon"
	"github.com/nickwells/param.mod/v3/param"
	"github.com/nickwells/param.mod/v3/param/psetter"
)

// Created: Sat Apr  8 15:49:28 2017

var onlyCheck bool
var schemaNames = make([]string, 0)

func addParams(ps *param.PSet) error {
	dbtcommon.AddParamDBName(ps)

	ps.Add("only-check", psetter.Bool{Value: &onlyCheck},
		"only check if the directories are present - don't create them."+
			" If this is set then the exit status will be set to 1 if"+
			" the directories are not all present and to zero otherwise")

	ps.Add("schema-names",
		psetter.StrList{
			Value: &schemaNames,
			Checks: []check.StringSlice{
				check.StringSliceStringCheck(
					check.StringMatchesPattern(
						regexp.MustCompile(`[a-z][a-z0-9_]*`),
						"a schema name: a leading lowercase character"+
							" followed by zero or more lowercase"+
							" letters, digits or underscores")),
				check.StringSliceLenGT(0),
				check.StringSliceNoDups,
			},
		},
		"a list of schemas to create the directories for",
		param.AltName("schemas"),
		param.Attrs(param.MustBeSet))

	return nil
}
