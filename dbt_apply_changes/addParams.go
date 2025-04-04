package main

import (
	"fmt"

	"github.com/nickwells/dbtools/internal/dbtcommon"
	"github.com/nickwells/param.mod/v6/paction"
	"github.com/nickwells/param.mod/v6/param"
	"github.com/nickwells/param.mod/v6/psetter"
)

const (
	paramNameShowRelease = "show-releases"
	paramNameRelease     = "release"
)

func addParams(prog *Prog) param.PSetOptFunc {
	return func(ps *param.PSet) error {
		var flagCounter paction.Counter

		ps.Add(paramNameRelease,
			psetter.String[string]{
				Value: &prog.releaseName,
			},
			"this gives the name of the release to be applied to the database."+
				" The name refers to a sub-directory of the "+
				dbtcommon.ReleaseScriptsBaseName+" directory."+
				" This directory should contain a manifest file ("+
				dbtcommon.ReleaseManifestFileName+
				"), the scripts to be run and optionally,"+
				" a file describing the release ("+
				dbtcommon.ReleaseReadMeFileName+
				"), a file describing any concerns that you should"+
				" address before applying the changes ("+
				dbtcommon.ReleaseWarningFileName+
				") and a sub-directory called "+
				dbtcommon.ReleaseSQLDirName+
				" containing SQL files",
			param.AltNames("rel", "r"),
			param.PostAction(flagCounter.MakeActionFunc()))

		ps.Add(paramNameShowRelease, psetter.Bool{Value: &prog.doNotApply},
			"print a message showing the available releases",
			param.Attrs(param.CommandLineOnly),
			param.PostAction(flagCounter.MakeActionFunc()))

		ps.Add("quiet", psetter.Bool{Value: &prog.quiet},
			"don't show the messages which announce the packages being"+
				" applied and don't show the contents of the "+
				dbtcommon.ReleaseReadMeFileName+
				" file (if it exists)")

		ps.Add("no-warn", psetter.Bool{Value: &prog.noWarn},
			"If there is a "+
				dbtcommon.ReleaseWarningFileName+
				" file don't show its contents and don't ask if you"+
				" want to proceed")

		dbtcommon.AddParamPsqlPath(prog.dbp, ps)

		ps.AddFinalCheck(func() error {
			if flagCounter.Count() == 0 {
				return fmt.Errorf(
					"you must set either the %q or the %q parameter",
					paramNameRelease, paramNameShowRelease)
			}

			return nil
		})
		ps.AddFinalCheck(func() error {
			if flagCounter.Count() > 1 {
				return fmt.Errorf(
					"you must only set one of the %q or %q parameters",
					paramNameRelease, paramNameShowRelease)
			}

			return nil
		})

		return nil
	}
}
