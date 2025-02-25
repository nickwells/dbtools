package main

// dbt_make_dirs

import (
	"fmt"
	"os"

	"github.com/nickwells/dbtools/internal/dbtcommon"
	"github.com/nickwells/verbose.mod/verbose"
)

// Created: Sat Apr  8 15:49:28 2017

// Prog holds program parameter values etc.
type Prog struct {
	onlyCheck   bool
	schemaNames []string
	dbp         *dbtcommon.DBParams
}

// NewProg returns a new Prog value, correctly initialised
func NewProg() *Prog {
	return &Prog{
		dbp: dbtcommon.NewDBParams(),
	}
}

func main() {
	prog := NewProg()
	ps := makeParamSet(prog)
	ps.Parse()

	var missingDirs bool

	verbose.Println("base dir: " + prog.dbp.BaseDirName)

	for _, schemaName := range prog.schemaNames {
		verbose.Println("db.schema: " + prog.dbp.DbName + "." + schemaName)

		if dbtcommon.CheckDirs(
			prog.dbp.BaseDirName, prog.dbp.DbName, schemaName) {
			verbose.Println("All required directories are already present")
		} else {
			verbose.Println("Some directories are missing")

			missingDirs = true

			if !prog.onlyCheck {
				err := dbtcommon.MakeMissingDirs(
					prog.dbp.BaseDirName, prog.dbp.DbName, schemaName)
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

	if prog.onlyCheck && missingDirs {
		os.Exit(1)
	}

	os.Exit(0)
}
