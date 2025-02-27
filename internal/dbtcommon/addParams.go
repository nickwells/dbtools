package dbtcommon

import (
	"os"
	"regexp"

	"github.com/nickwells/check.mod/v2/check"
	"github.com/nickwells/filecheck.mod/filecheck"
	"github.com/nickwells/location.mod/location"
	"github.com/nickwells/param.mod/v6/param"
	"github.com/nickwells/param.mod/v6/psetter"
)

// DBParams holds all of the common values for the DB tools
type DBParams struct {
	// BaseDirName is the name of the directory where the dbtools directories
	// and files are to be found
	BaseDirName string

	// PsqlPath is the name of the postgresql command line tool
	PsqlPath string

	// DbName is the name of the postgresql database to use
	DbName string
}

// NewDBParams returns a pointer to a properly initialised DBParams object
func NewDBParams() *DBParams {
	return &DBParams{
		PsqlPath: "psql",
	}
}

const (
	// DbtEnvPrefix is the prefix that is applied to any parameters to be set
	// in the environment rather than the command line
	DbtEnvPrefix = "DBTOOLS_"

	// DbtBaseDirParamName is the name of the parameter that is used to set
	// the BaseDirName value
	DbtBaseDirParamName = "base-dir"

	// DbtPsqlPathParamName is the name of the parameter that is used to
	// override the name of the postgresql command line tool
	DbtPsqlPathParamName = "psql-path"
)

// setBaseDirEnvVar sets the value of the environment variable for the
// Base-directory to the value in the DBParams object
func setBaseDirEnvVar(dbp *DBParams) param.ActionFunc {
	return func(_ location.L, p *param.ByName, _ []string) error {
		envVarName := DbtEnvPrefix +
			param.ConvertParamNameToEnvVarName(p.Name())
		return os.Setenv(envVarName, dbp.BaseDirName)
	}
}

// setPsqlPathEnvVar sets the value of the environment variable for the
// psql pathname to the value in the DBParams object
func setPsqlPathEnvVar(dbp *DBParams) param.ActionFunc {
	return func(_ location.L, p *param.ByName, _ []string) error {
		envVarName := DbtEnvPrefix +
			param.ConvertParamNameToEnvVarName(p.Name())
		return os.Setenv(envVarName, dbp.PsqlPath)
	}
}

// AddParams will add the db tools params to the given param set This should
// be called before the PSet is parsed
func AddParams(dbp *DBParams) param.PSetOptFunc {
	return func(ps *param.PSet) error {
		const paramGroupName = "pkg.dbtcommon"

		ps.AddGroup(paramGroupName,
			`parameters for any of the commands in the dbTools family`)

		ps.SetEnvPrefix(DbtEnvPrefix)
		_ = setGlobalConfigFileForGroupPkgDbtcommon(ps)
		_ = setConfigFileForGroupPkgDbtcommon(ps)

		ps.Add(DbtBaseDirParamName,
			psetter.Pathname{
				Value:       &dbp.BaseDirName,
				Expectation: filecheck.DirExists(),
			},
			"the name of the directory under which the database directories"+
				" will be found. These directories all exist under a db"+
				" directory so the path will be: <base-dir>/"+DbtDirName+"/...",
			param.Attrs(param.MustBeSet),
			param.GroupName(paramGroupName),
			param.PostAction(setBaseDirEnvVar(dbp)))

		return nil
	}
}

// AddParamDBName adds the standard db parameter. Not all commands need this
// and so it is not added in the AddParams function above
func AddParamDBName(dbp *DBParams, ps *param.PSet, opts ...param.OptFunc) {
	opts = append(opts, param.AltNames("db"),
		param.Attrs(param.MustBeSet))
	ps.Add("db-name",
		psetter.String[string]{
			Value: &dbp.DbName,
			Checks: []check.ValCk[string]{
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
func AddParamPsqlPath(dbp *DBParams, ps *param.PSet, opts ...param.OptFunc) {
	opts = append(opts, param.PostAction(setPsqlPathEnvVar(dbp)))
	ps.Add(DbtPsqlPathParamName,
		psetter.Pathname{
			Value:       &dbp.PsqlPath,
			Expectation: filecheck.FileExists(),
		},
		"the pathname of the psql command - you will need to give this"+
			" parameter if the psql command is not in the PATH",
		opts...)
}
