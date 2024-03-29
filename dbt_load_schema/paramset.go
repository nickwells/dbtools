package main

import (
	"github.com/nickwells/dbtools/internal/dbtcommon"
	"github.com/nickwells/param.mod/v6/param"
	"github.com/nickwells/param.mod/v6/paramset"
	"github.com/nickwells/verbose.mod/verbose"
	"github.com/nickwells/versionparams.mod/versionparams"
)

// makeParamSet generates the param set ready for parsing
func makeParamSet(prog *Prog) *param.PSet {
	return paramset.NewOrPanic(
		addParams(prog),
		verbose.AddParams,
		versionparams.AddParams,
		dbtcommon.AddParams(prog.dbp),
		param.SetProgramDescription("this will load the named schema files"),
	)
}
