package main

import (
	"fmt"
	"regexp"

	"github.com/nickwells/check.mod/check"
	"github.com/nickwells/dbtcommon.mod/v2/dbtcommon"
	"github.com/nickwells/param.mod/v4/param"
	"github.com/nickwells/param.mod/v4/param/paction"
	"github.com/nickwells/param.mod/v4/param/psetter"
)

func addParams(ps *param.PSet) error {
	var schemaObjParamCounter paction.Counter
	af := (&schemaObjParamCounter).MakeActionFunc()

	dbtcommon.AddParamDBName(ps, param.Attrs(param.MustBeSet))
	ps.AddFinalCheck(checkDBSchemaExists)
	dbtcommon.AddParamPsqlPath(ps)

	ps.Add("schema", psetter.String{Value: &schemaName},
		"this gives the name of the schema that is to be applied to the"+
			" database. The name refers to the name of a schema under "+
			dbtcommon.DBSchemaDirName+
			". This directory should contain the tables, triggers,"+
			" functions and types (in subdirectories of the same names)",
		param.Attrs(param.MustBeSet))
	ps.AddFinalCheck(checkDBSchemaExists)

	ps.Add("macro-dirs",
		psetter.StrList{
			Value: &macroDirs,
			Checks: []check.StringSlice{
				check.StringSliceLenGT(0),
				check.StringSliceNoDups,
			},
		},
		"a list of additional directories in which macros may be found")

	schemaObjCheck := check.StringSliceStringCheck(
		check.StringMatchesPattern(
			regexp.MustCompile(`[a-z_][a-z0-9_]*`),
			"a letter or underscore followed by 0 or more letters,"+
				" digits or undescores"))

	ps.Add("types",
		psetter.StrList{
			Value: &schemaTypes,
			Checks: []check.StringSlice{
				check.StringSliceLenGT(0),
				check.StringSliceNoDups,
				schemaObjCheck,
			},
		},
		"this gives the list of types to be applied to the schema",
		param.PostAction(af),
		param.AltName("type"))

	ps.Add("tables",
		psetter.StrList{
			Value: &schemaTables,
			Checks: []check.StringSlice{
				check.StringSliceLenGT(0),
				check.StringSliceNoDups,
				schemaObjCheck,
			},
		},
		"this gives the list of tables to be applied to the schema",
		param.PostAction(af),
		param.AltName("table"),
		param.AltName("tbl"))

	ps.Add("funcs",
		psetter.StrList{
			Value: &schemaFuncs,
			Checks: []check.StringSlice{
				check.StringSliceLenGT(0),
				check.StringSliceNoDups,
				schemaObjCheck,
			},
		},
		"this gives the list of funcs to be applied to the schema",
		param.PostAction(af),
		param.AltName("func"))

	ps.Add("triggers",
		psetter.StrList{
			Value: &schemaTriggers,
			Checks: []check.StringSlice{
				check.StringSliceLenGT(0),
				check.StringSliceNoDups,
				schemaObjCheck,
			},
		},
		"this gives the list of triggers to be applied to the schema",
		param.PostAction(af),
		param.AltName("trigger"))

	ps.Add("create-audit-tables", psetter.Bool{Value: &createAuditTables},
		"this will create audit tables for every table created")

	ps.Add("display-sql-only", psetter.Bool{Value: &displayOnly},
		"this will just print out the sql that would be applied"+
			" without changing the database",
		param.AltName("debug"))

	ps.AddFinalCheck(func() error {
		if schemaObjParamCounter.Count() == 0 {
			return fmt.Errorf(
				"You must give at least one type, table, trigger or func name")
		}
		return nil
	})

	return nil
}
