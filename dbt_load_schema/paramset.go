package main

import (
	"github.com/nickwells/dbtools/internal/dbtcommon"
	"github.com/nickwells/param.mod/v7/param"
	"github.com/nickwells/param.mod/v7/paramset"
	"github.com/nickwells/verbose.mod/verbose"
	"github.com/nickwells/versionparams.mod/versionparams"
)

// makeParamSet generates the param set ready for parsing
func makeParamSet(prog *Prog) *param.PSet {
	return paramset.New(
		addParams(prog),
		verbose.AddParams,
		versionparams.AddParams,
		dbtcommon.AddParams(prog.dbp),
		param.SetProgramDescription("this will load the named schema files"),
	)
}
