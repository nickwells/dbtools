package main

// dbt_make_dirs

import (
	"fmt"
	"os"

	"github.com/nickwells/dbtools/internal/dbtcommon"
	"github.com/nickwells/param.mod/v5/param"
	"github.com/nickwells/param.mod/v5/param/paramset"
	"github.com/nickwells/verbose.mod/verbose"
	"github.com/nickwells/versionparams.mod/versionparams"
)

// Created: Sat Apr  8 15:49:28 2017

func main() {
	ps := paramset.NewOrDie(
		addParams,
		verbose.AddParams,
		versionparams.AddParams,
		dbtcommon.AddParams,
		param.SetProgramDescription(
			"this will create the database directories under"+
				" the base (project) directory"))
	ps.Parse()

	var missingDirs bool

	verbose.Println("base dir: " + dbtcommon.BaseDirName)
	for _, schemaName := range schemaNames {
		verbose.Println("db.schema: " + dbtcommon.DbName + "." + schemaName)
		if dbtcommon.CheckDirs(dbtcommon.DbName, schemaName) {
			verbose.Println("All required directories are already present")
		} else {
			verbose.Println("Some directories are missing")
			missingDirs = true

			if !onlyCheck {
				err := dbtcommon.MakeMissingDirs(dbtcommon.DbName, schemaName)
				if err != nil {
					fmt.Fprintf(os.Stderr,
						"Couldn't create all the missing subdirectories: %s\n",
						err)
					os.Exit(1)
				}
				verbose.Println("Missing directories created")
			}
		}
	}
	if onlyCheck && missingDirs {
		os.Exit(1)
	}
	os.Exit(0)
}
