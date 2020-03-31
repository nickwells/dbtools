package main

import (
	"fmt"

	"github.com/nickwells/dbtcommon.mod/dbtcommon"
	"github.com/nickwells/param.mod/v4/param"
	"github.com/nickwells/param.mod/v4/param/paction"
	"github.com/nickwells/param.mod/v4/param/psetter"
)

const (
	pNameShowRelease = "show-releases"
	pNameRelease     = "release"
)

func addParams(ps *param.PSet) error {
	var flagCounter paction.Counter

	ps.Add(pNameRelease,
		psetter.String{
			Value: &releaseName,
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
			"containing SQL files",
		param.AltName("rel"),
		param.AltName("r"),
		param.PostAction(flagCounter.MakeActionFunc()))

	ps.Add(pNameShowRelease, psetter.Bool{Value: &showReleasesFlag},
		"print a message showing the available releases",
		param.Attrs(param.CommandLineOnly),
		param.PostAction(flagCounter.MakeActionFunc()))

	ps.Add("quiet", psetter.Bool{Value: &quiet},
		"don't show the messages which announce the packages being"+
			" applied and don't show the contents of the "+
			dbtcommon.ReleaseReadMeFileName+
			" file (if it exists)")

	ps.Add("no-warn", psetter.Bool{Value: &noWarn},
		"If there is a "+
			dbtcommon.ReleaseWarningFileName+
			" file don't show its contents and don't ask if you"+
			" want to proceed")

	dbtcommon.AddParamPsqlPath(ps)

	ps.AddFinalCheck(func() error {
		if flagCounter.Count() == 0 {
			return fmt.Errorf("you must set either the %q or the %q parameter",
				pNameRelease, pNameShowRelease)
		}
		return nil
	})
	ps.AddFinalCheck(func() error {
		if flagCounter.Count() > 1 {
			return fmt.Errorf("you must only set one of the %q or %q parameters",
				pNameRelease, pNameShowRelease)
		}
		return nil
	})

	return nil
}
