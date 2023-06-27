package main

// dbt_load_schema

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nickwells/dbtools/internal/dbtcommon"
	"github.com/nickwells/filecheck.mod/filecheck"
	"github.com/nickwells/location.mod/location"
	"github.com/nickwells/macros.mod/macros"
	"github.com/nickwells/verbose.mod/verbose"
)

// Created: Thu Apr 20 22:39:40 2017

const (
	dfltSchema = "public"
)

// checkDBSchemaExists checks that the given database / schema directory
// exists in the DBS base directory
func (prog *Prog) checkDBSchemaExists() error {
	if prog.dbp.BaseDirName == "" ||
		prog.dbp.DbName == "" ||
		prog.schemaName == "" {
		return nil
	}

	exists := filecheck.DirExists()
	dirName := dbtcommon.DbtDirDBSchema(
		prog.dbp.BaseDirName, prog.dbp.DbName, prog.schemaName)
	return exists.StatusCheck(dirName)
}

// makeFileLists converts the slice of names into a slice of file names that
// exist in the DB.schema directory under dirName. If any of the files does
// not exist then the error is returned in the errs slice
func (prog *Prog) makeFileLists() {
	errs := []error{}

	for dirName, s := range prog.schemas {
		sdName := filepath.Join(dbtcommon.DbtDirDBSchema(
			prog.dbp.BaseDirName,
			prog.dbp.DbName,
			prog.schemaName),
			dirName)

		existence := filecheck.FileExists()

		for _, name := range s.names {
			if !strings.HasSuffix(name, ".sql") {
				name += ".sql"
			}
			name = filepath.Join(sdName, name)
			err := existence.StatusCheck(name)
			if err != nil {
				errs = append(errs, err)
			} else {
				s.files = append(s.files, name)
			}
		}

		if len(errs) != 0 {
			for _, err := range errs {
				fmt.Fprintln(os.Stderr, err)
			}
			os.Exit(1)
		}
	}
}

// applyFile applies the file to the schema
func (prog *Prog) applyFile(f string) error {
	verbose.Println("applying schema file: ",
		strings.Replace(f, prog.dbp.BaseDirName, "[base-dir]", 1))
	sql, err := prog.translateFile(f)
	if err != nil {
		return err
	}
	err = prog.applySQL(sql)
	return err
}

// translateFile reads the file applying any macros found
func (prog *Prog) translateFile(f string) (string, error) {
	sqlFile, err := os.Open(f)
	if err != nil {
		return "", err
	}
	defer sqlFile.Close()

	sql := "SET search_path TO " + prog.schemaName + ";\n"
	scanner := bufio.NewScanner(sqlFile)
	loc := location.New(f)
	for scanner.Scan() {
		loc.Incr()
		line, err := prog.macroCache.Substitute(scanner.Text(), loc)
		if err != nil {
			return "", err
		}
		sql += line + "\n"
	}
	err = scanner.Err()
	if err != nil {
		return "", err
	}
	return sql, nil
}

// applySQL writes the passed buffer to the sql command
func (prog *Prog) applySQL(sql string) error {
	if prog.displayOnly {
		fmt.Println(sql)
		return nil
	}

	cmd := dbtcommon.SQLCommand(prog.dbp, "-")

	cmdIn, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer cmdIn.Close()
		_, _ = io.WriteString(cmdIn, sql)
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", out)
	}

	return err
}

// generateAuditTable generates the audit table for the named table
func (prog *Prog) generateAuditTable(tbl string) {
	if !prog.createAuditTables {
		return
	}
	auditTbl := tbl + "_aud"

	sql := "SET search_path TO " + prog.schemaName + ";\n" +
		"CREATE TABLE " + auditTbl + " AS SELECT * FROM " + tbl +
		" WITH NO DATA;\n"
	err := prog.applySQL(sql)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"Could not generate the audit table (%s) for table %s\n",
			auditTbl, tbl)
		fmt.Fprintln(os.Stderr, "sql failed:\n", sql)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// applyAllFiles applies the files from the schema directories to the
// database. It exits on the first failure to apply a file.
func (prog *Prog) applyAllFiles() {
	verbose.Println("applying files")
	for schemaPart, s := range prog.schemas {
		verbose.Println("\t", schemaPart)
		for i, f := range s.files {
			verbose.Println("\t\t", f)
			err := prog.applyFile(f)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Couldn't apply the schema %q file: %s\n",
					schemaPart, f)
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			if schemaPart == dbtcommon.SchemaSubDirTables {
				prog.generateAuditTable(s.names[i])
			}
		}
	}
}

// makeMacroCache constructs the macro cache. It adds the default macro
// directory to the list of directories first.
func (prog *Prog) makeMacroCache() {
	verbose.Println("construct the Macro cache")
	prog.macroDirs = append(prog.macroDirs,
		dbtcommon.DbtDirMacros(prog.dbp.BaseDirName))

	mc, err := macros.NewCache(
		macros.Dirs(prog.macroDirs...),
		macros.Suffix(".sql"))
	if err != nil {
		fmt.Println("Couldn't construct the macro cache: ", err)
		os.Exit(1)
	}
	prog.macroCache = mc
}

// schema holds the list of schema part names and their associated files
type schema struct {
	names []string
	files []string
}

// Prog holds program parameters and status
type Prog struct {
	// parameters
	schemaName string

	createAuditTables bool
	displayOnly       bool

	schemas map[string]*schema
	dbp     *dbtcommon.DBParams

	macroDirs  []string
	macroCache *macros.Cache
}

// NewProg returns a new Prog instance with the default values set
func NewProg() *Prog {
	return &Prog{
		schemaName: dfltSchema,
		schemas:    make(map[string]*schema),
		dbp:        dbtcommon.NewDBParams(),
	}
}

func main() {
	prog := NewProg()
	ps := makeParamSet(prog)
	ps.Parse()

	prog.makeMacroCache()

	prog.makeFileLists()

	prog.applyAllFiles()
}
