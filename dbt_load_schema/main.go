// dbt_load_schema
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nickwells/check.mod/check"
	"github.com/nickwells/dbtcommon.mod/v2/dbtcommon"
	"github.com/nickwells/filecheck.mod/filecheck"
	"github.com/nickwells/location.mod/location"
	"github.com/nickwells/macros.mod/macros"
	"github.com/nickwells/param.mod/v5/param"
	"github.com/nickwells/param.mod/v5/param/paramset"
	"github.com/nickwells/verbose.mod/verbose"
)

// Created: Thu Apr 20 22:39:40 2017

var schemaName string
var schemaTypes []string
var schemaTables []string
var schemaFuncs []string
var schemaTriggers []string
var createAuditTables bool
var displayOnly bool

var macroDirs = make([]string, 0, 1)

// checkDBSchemaExists checks that the given database / schema directory
// exists in the DBS base directory
func checkDBSchemaExists() error {
	if dbtcommon.BaseDirName == "" ||
		dbtcommon.DbName == "" ||
		schemaName == "" {
		return nil
	}

	es := filecheck.Provisos{
		Checks:    []check.FileInfo{check.FileInfoIsDir},
		Existence: filecheck.MustExist,
	}
	dirName := dbtcommon.DbtDirDBSchema(dbtcommon.DbName, schemaName)
	return es.StatusCheck(dirName)
}

func init() {
	schemaTypes = make([]string, 0)
	schemaTables = make([]string, 0)
	schemaFuncs = make([]string, 0)
	schemaTriggers = make([]string, 0)
}

// makeFileList converts the slice of names into a slice of file names that
// exist in the DB.schema directory under dirName. If any of the files does
// not exist then the error is returned in the errs slice
func makeFileList(names []string, dirName string, errs *[]error) []string {
	fileList := make([]string, 0)
	if len(names) == 0 {
		return fileList
	}

	sdName := filepath.Join(dbtcommon.DbtDirDBSchema(dbtcommon.DbName,
		schemaName),
		dirName)

	es := filecheck.Provisos{
		Checks:    []check.FileInfo{check.FileInfoIsRegular},
		Existence: filecheck.MustExist,
	}

	for _, name := range names {
		if !strings.HasSuffix(name, ".sql") {
			name += ".sql"
		}
		name = filepath.Join(sdName, name)
		err := es.StatusCheck(name)
		if err != nil {
			*errs = append(*errs, err)
		} else {
			fileList = append(fileList, name)
		}
	}

	return fileList
}

// applyFile applies the file to the schema
func applyFile(f string, c *macros.Cache) error {
	verbose.Println("applying schema file: ",
		strings.Replace(f, dbtcommon.BaseDirName, "[base-dir]", 1))
	buf, err := translateFile(f, c)
	if err != nil {
		return err
	}
	err = applyBuffer(buf)
	return err
}

// translateFile reads the file applying any macros found
func translateFile(f string, c *macros.Cache) (*bytes.Buffer, error) {
	sqlFile, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer sqlFile.Close()
	var buf = &bytes.Buffer{}

	_, _ = (buf).WriteString("SET search_path TO " + schemaName + ";\n")
	scanner := bufio.NewScanner(sqlFile)
	loc := location.New(f)
	for scanner.Scan() {
		loc.Incr()
		line, err := c.Substitute(scanner.Text(), loc)
		if err != nil {
			return nil, err
		}
		_, _ = (buf).WriteString(line)
		_, _ = (buf).WriteString("\n") // add the newline stripped by Scan
	}
	err = scanner.Err()
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// applyBuffer writes the passed buffer to the sql command
func applyBuffer(buf *bytes.Buffer) error {
	if displayOnly {
		fmt.Println(buf.String())
		return nil
	}

	cmd := dbtcommon.SQLCommand("-")

	cmdIn, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer cmdIn.Close()
		(buf).WriteTo(cmdIn) //nolint: errcheck
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", out)
	}

	return err
}

// applyAllFiles applies the files from the schema directory to the
// database. It exits on the first failure to apply a file.
func applyAllFiles(files []string, typeName string, cache *macros.Cache) {
	for _, f := range files {
		err := applyFile(f, cache)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not apply the schema %s file: %s\n",
				typeName, f)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

// generateAuditTables ...
func generateAuditTables() {
	if !createAuditTables {
		return
	}

	for _, t := range schemaTables {
		tAud := t + "_aud"
		var buf = &bytes.Buffer{}

		_, _ = (buf).WriteString("SET search_path TO " + schemaName + ";\n")
		_, _ = (buf).WriteString(
			"CREATE TABLE " + tAud +
				" AS SELECT * FROM " + t +
				" WITH NO DATA;\n")
		err := applyBuffer(buf)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"Could not create the audit table for: %s\n", t)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

// makeMacroCache constructs the macro cache. It adds the default macro
// directory to the list of directories first.
func makeMacroCache() *macros.Cache {
	macroDirs = append(macroDirs, dbtcommon.DbtDirMacros())

	macroCache, err := macros.NewCache(
		macros.Dirs(macroDirs...),
		macros.Suffix(".sql"))
	if err != nil {
		fmt.Println("Couldn't construct the macro cache: ", err)
		os.Exit(1)
	}

	return macroCache
}

func main() {
	ps := paramset.NewOrDie(
		addParams,
		verbose.AddParams,
		dbtcommon.AddParams,
		param.SetProgramDescription("this will load the named schema files"))
	ps.Parse()

	macroCache := makeMacroCache()

	var errs = make([]error, 0)
	typeFiles := makeFileList(schemaTypes,
		dbtcommon.SchemaSubDirTypes, &errs)
	tableFiles := makeFileList(schemaTables,
		dbtcommon.SchemaSubDirTables, &errs)
	funcFiles := makeFileList(schemaFuncs,
		dbtcommon.SchemaSubDirFuncs, &errs)
	triggerFiles := makeFileList(schemaTriggers,
		dbtcommon.SchemaSubDirTriggers, &errs)

	if len(errs) != 0 {
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}

	applyAllFiles(typeFiles, "type", macroCache)
	applyAllFiles(tableFiles, "table", macroCache)
	generateAuditTables()
	applyAllFiles(funcFiles, "func", macroCache)
	applyAllFiles(triggerFiles, "trigger", macroCache)

}
