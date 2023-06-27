package main

import (
	"github.com/nickwells/dbtools/internal/dbtcommon"
	"github.com/nickwells/param.mod/v5/param"
	"github.com/nickwells/param.mod/v5/param/paramset"
	"github.com/nickwells/versionparams.mod/versionparams"
)

// makeParamSet generates the param set ready for parsing
func makeParamSet(prog *Prog) *param.PSet {
	return paramset.NewOrPanic(
		addParams(prog),
		versionparams.AddParams,
		dbtcommon.AddParams(prog.dbp),
		param.SetProgramDescription("this will apply a set of scripts"+
			" (typically shell scripts but any executable can be run)."+
			" The contents of the release directory is checked against"+
			" a Manifest file which also defines the order in which"+
			" they should be applied"),
	)
}
