package dbtcommon

import (
	"os"
	"regexp"

	"github.com/nickwells/check.mod/v2/check"
	"github.com/nickwells/filecheck.mod/filecheck"
	"github.com/nickwells/location.mod/location"
	"github.com/nickwells/param.mod/v5/param"
	"github.com/nickwells/param.mod/v5/param/psetter"
)

// BaseDirName is the name of the directory where the dbtools directories and
// files are to be found
var BaseDirName string

// PsqlPath is the name of the postgresql command line tool
var PsqlPath = "psql"

// DbName is the name of the postgresql database to use
var DbName string

// DbtEnvPrefix is the prefix that is applied to any parameters to be set in
// the environment rather than the command line
var DbtEnvPrefix = "DBTOOLS_"

// DbtBaseDirParamName is the name of the parameter that is used to set the
// BaseDirName value
var DbtBaseDirParamName = "base-dir"

// DbtPsqlPathParamName is the name of the parameter that is used to override
// the name of the postgresql command line tool
var DbtPsqlPathParamName = "psql-path"

// setBaseDirEnvVar ...
func setBaseDirEnvVar(_ location.L, p *param.ByName, _ []string) error {
	envVarName := DbtEnvPrefix +
		param.ConvertParamNameToEnvVarName(p.Name())
	return os.Setenv(envVarName, BaseDirName)
}

// setPsqlPathEnvVar ...
func setPsqlPathEnvVar(_ location.L, p *param.ByName, _ []string) error {
	envVarName := DbtEnvPrefix +
		param.ConvertParamNameToEnvVarName(p.Name())
	return os.Setenv(envVarName, PsqlPath)
}

// AddParams will add the db tools params to the given param set This should
// be called before the PSet is parsed
func AddParams(ps *param.PSet) error {
	const paramGroupName = "pkg.dbtcommon"
	ps.AddGroup(paramGroupName,
		`parameters for any of the commands in the dbTools family`)

	ps.SetEnvPrefix(DbtEnvPrefix)
	_ = setGlobalConfigFileForGroupPkgDbtcommon(ps)
	_ = setConfigFileForGroupPkgDbtcommon(ps)

	ps.Add(DbtBaseDirParamName,
		psetter.Pathname{
			Value:       &BaseDirName,
			Expectation: filecheck.DirExists(),
		},
		"the name of the directory under which the database directories"+
			" will be found. These directories all exist under a db"+
			" directory so the path will be: <base-dir>/"+DbtDirName+"/...",
		param.Attrs(param.MustBeSet),
		param.GroupName(paramGroupName),
		param.PostAction(setBaseDirEnvVar))

	return nil
}

// AddParamDBName adds the standard db parameter. Not all commands need this
// and so it is not added in the AddParams function above
func AddParamDBName(ps *param.PSet, opts ...param.OptFunc) {
	opts = append(opts, param.AltName("db"),
		param.Attrs(param.MustBeSet))
	ps.Add("db-name",
		psetter.String{
			Value: &DbName,
			Checks: []check.String{
				check.StringMatchesPattern[string](
					regexp.MustCompile(`[a-z][a-z0-9_]*`),
					"a database name: a leading lowercase character"+
						" followed by zero or more lowercase letters,"+
						" digits or underscores"),
			},
		},
		"the name of the database",
		opts...)
}

// AddParamPsqlPath adds the standard psql-name parameter. Not all commands
// need this and so it is not added in the AddParams function above
func AddParamPsqlPath(ps *param.PSet, opts ...param.OptFunc) {
	opts = append(opts, param.PostAction(setPsqlPathEnvVar))
	ps.Add(DbtPsqlPathParamName,
		psetter.Pathname{
			Value:       &PsqlPath,
			Expectation: filecheck.FileExists(),
		},
		"the pathname of the psql command - you will need to do this"+
			" if the psql command was not in the PATH",
		opts...)
}
