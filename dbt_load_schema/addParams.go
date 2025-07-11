package main

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/nickwells/check.mod/v2/check"
	"github.com/nickwells/dbtools/internal/dbtcommon"
	"github.com/nickwells/location.mod/location"
	"github.com/nickwells/param.mod/v6/paction"
	"github.com/nickwells/param.mod/v6/param"
	"github.com/nickwells/param.mod/v6/psetter"
)

const (
	paramNameTypes    = "types"
	paramNameTables   = "tables"
	paramNameFuncs    = "funcs"
	paramNameTriggers = "triggers"
)

func addParams(prog *Prog) param.PSetOptFunc {
	return func(ps *param.PSet) error {
		loadItemParams := []string{
			paramNameTypes,
			paramNameTables,
			paramNameFuncs,
			paramNameTriggers,
		}

		schemaObjParamCounter := paction.Counter{}
		countSchema := (&schemaObjParamCounter).MakeActionFunc()

		dbtcommon.AddParamDBName(prog.dbp, ps, param.Attrs(param.MustBeSet))
		ps.AddFinalCheck(prog.checkDBSchemaExists)
		dbtcommon.AddParamPsqlPath(prog.dbp, ps)

		ps.Add("schema", psetter.String[string]{Value: &prog.schemaName},
			"this gives the name of the schema that is to be applied to the"+
				" database. The name refers to the name of a schema under "+
				dbtcommon.DBSchemaDirName+
				". This directory should contain the tables, triggers,"+
				" functions and types (in subdirectories of the same names)",
		)

		ps.Add("macro-dirs",
			psetter.StrList[string]{
				Value: &prog.macroDirs,
				Checks: []check.ValCk[[]string]{
					check.SliceLength[[]string](check.ValGT(0)),
					check.SliceHasNoDups[[]string, string],
				},
			},
			"a list of additional directories in which macros may be found")

		schemaObjNameCheck := check.SliceAll[[]string](
			check.StringMatchesPattern[string](
				regexp.MustCompile(`[a-z_][a-z0-9_]*`),
				"a lowercase letter or underscore followed by 0 or more"+
					" lowercase letters, undescores or digits"))
		noDupsCheck := check.SliceHasNoDups[[]string, string]
		schemaObjChecks := []check.ValCk[[]string]{
			check.SliceLength[[]string](check.ValGT(0)),
			noDupsCheck,
			schemaObjNameCheck,
		}

		{
			var names []string

			ps.Add(paramNameTypes,
				psetter.StrList[string]{
					Value:  &names,
					Checks: schemaObjChecks,
				},
				"this gives the list of types to be applied to the schema",
				param.AltNames("type"), param.PostAction(countSchema),
				param.PostAction(
					func(_ location.L, _ *param.ByName, _ []string) error {
						s := prog.schemas[dbtcommon.SchemaSubDirTypes]
						if s == nil {
							s = &schema{}
							prog.schemas[dbtcommon.SchemaSubDirTypes] = s
						}

						s.names = append(s.names, names...)

						if err := noDupsCheck(s.names); err != nil {
							return fmt.Errorf("duplicate types: %w", err)
						}

						return nil
					}),
				param.SeeAlso(loadItemParams...),
			)
		}

		{
			var names []string

			ps.Add(paramNameTables,
				psetter.StrList[string]{
					Value:  &names,
					Checks: schemaObjChecks,
				},
				"this gives the list of tables to be applied to the schema",
				param.AltNames("table", "tbl"),
				param.PostAction(countSchema),
				param.PostAction(
					func(_ location.L, _ *param.ByName, _ []string) error {
						s := prog.schemas[dbtcommon.SchemaSubDirTables]
						if s == nil {
							s = &schema{}
							prog.schemas[dbtcommon.SchemaSubDirTables] = s
						}

						s.names = append(s.names, names...)

						if err := noDupsCheck(s.names); err != nil {
							return fmt.Errorf("duplicate tables: %w", err)
						}

						return nil
					}),
				param.SeeAlso(loadItemParams...),
			)
		}

		{
			var names []string

			ps.Add(paramNameFuncs,
				psetter.StrList[string]{
					Value:  &names,
					Checks: schemaObjChecks,
				},
				"this gives the list of funcs to be applied to the schema",
				param.AltNames("func"),
				param.PostAction(countSchema),
				param.PostAction(
					func(_ location.L, _ *param.ByName, _ []string) error {
						s := prog.schemas[dbtcommon.SchemaSubDirFuncs]
						if s == nil {
							s = &schema{}
							prog.schemas[dbtcommon.SchemaSubDirFuncs] = s
						}

						s.names = append(s.names, names...)

						if err := noDupsCheck(s.names); err != nil {
							return fmt.Errorf("duplicate funcs: %w", err)
						}

						return nil
					}),
				param.SeeAlso(loadItemParams...),
			)
		}

		{
			var names []string

			ps.Add(paramNameTriggers,
				psetter.StrList[string]{
					Value:  &names,
					Checks: schemaObjChecks,
				},
				"this gives the list of triggers to be applied to the schema",
				param.AltNames("trigger"),
				param.PostAction(countSchema),
				param.PostAction(
					func(_ location.L, _ *param.ByName, _ []string) error {
						s := prog.schemas[dbtcommon.SchemaSubDirTriggers]
						if s == nil {
							s = &schema{}
							prog.schemas[dbtcommon.SchemaSubDirTriggers] = s
						}

						s.names = append(s.names, names...)

						if err := noDupsCheck(s.names); err != nil {
							return fmt.Errorf("duplicate triggers: %w", err)
						}

						return nil
					}),
				param.SeeAlso(loadItemParams...),
			)
		}

		ps.Add("create-audit-tables",
			psetter.Bool{Value: &prog.createAuditTables},
			"this will create audit tables for every table created")

		ps.Add("display-sql-only", psetter.Bool{Value: &prog.displayOnly},
			"this will just print out the sql that would be applied"+
				" without changing the database",
			param.AltNames("debug", "dbg", "sql-only"))

		ps.AddFinalCheck(func() error {
			if schemaObjParamCounter.Count() == 0 {
				return errors.New("you must give at least one" +
					" type, table, trigger or func name")
			}

			return nil
		})

		return nil
	}
}
