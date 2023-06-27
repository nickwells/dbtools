package main

// dbt_apply_changes

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nickwells/cli.mod/cli/responder"
	"github.com/nickwells/dbtools/internal/dbtcommon"
	"github.com/nickwells/location.mod/location"
)

// Created: Wed Apr 12 21:29:46 2017

type action int

const (
	doNothing action = iota
	confirm
	abort
)

const errorPrefix = "*** Error ***"

// reportErrors checks if there are any errors and if so prints them and exits
func reportErrors(errors ...error) {
	if len(errors) == 0 {
		return
	}

	for _, err := range errors {
		if err != nil {
			fmt.Println(errorPrefix, err)
		}
	}
	os.Exit(1)
}

// printFileHeader prints the header for the printFile func below
func printFileHeader(title, sep string) {
	fmt.Println()
	fmt.Print(sep)
	fmt.Printf("\t\t%s\n", title)
	fmt.Print(sep)
}

// printAlert prints a message with a surrounding alert box
func printAlert(msg string) action {
	box := strings.Repeat("*", 40)

	fmt.Println(box)
	fmt.Println("* " + errorPrefix)
	fmt.Println("* " + msg)
	fmt.Println(box)

	return abort
}

// printFile prints the file if it exists and is not empty and returns a
// value indicating what to do next
func printFile(fileName, title, sep string) action {
	nextAction := doNothing

	fStat, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nextAction
		}
		return printAlert(
			fmt.Sprintf("Couldn't open the file: %q: %s", fileName, err))
	}

	if !fStat.Mode().IsRegular() {
		return printAlert(
			fmt.Sprintf("%q exists but it is not a regular file",
				fileName))
	}

	fd, err := os.Open(fileName)
	if err != nil {
		return printAlert(
			fmt.Sprintf("Couldn't open %q: %s\n", fileName, err))
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		if nextAction == doNothing {
			printFileHeader(title, sep)
			nextAction = confirm
		}
		line := scanner.Text()
		fmt.Println(line)
	}
	if err = scanner.Err(); err != nil {
		return printAlert(
			fmt.Sprintf("%s While reading %q: %s\n",
				errorPrefix, fileName, err))
	}

	if nextAction == confirm {
		fmt.Print(sep)
		fmt.Println()
	}
	return nextAction
}

// showReadMe prints the ReadMe file if any unless the quiet flag has been set
func (prog *Prog) showReadMe() {
	if prog.quiet {
		return
	}

	printFile(dbtcommon.DbtFileReleaseReadMe(
		prog.dbp.BaseDirName, prog.releaseName),
		"Note", "========================================\n")
}

// showWarning prints the Warnings file (if any) unless the noWarn flag has
// been set
func (prog *Prog) showWarning() {
	if prog.noWarn {
		return
	}

	r := responder.NewOrPanic(
		"Do you want to continue",
		map[rune]string{
			'y': "apply the changes",
			'n': "abort the changes",
		},
		responder.SetMaxReprompts(5))

	switch printFile(
		dbtcommon.DbtFileReleaseWarning(prog.dbp.BaseDirName, prog.releaseName),
		"Warning", "#################################################\n") {
	case confirm:
		if r.GetResponseOrDie() == 'y' {
			return
		}
		fallthrough
	case abort:
		os.Exit(1)
	}
}

// checkReleaseDir checks that the release directory exists and reports an
// error and a list of releases and exits if it doesn't
func (prog *Prog) checkReleaseDir() {
	err := prog.releaseDirIsOK()
	if err != nil {
		fmt.Printf("%s Bad release: %s\n", errorPrefix, prog.releaseName)
		fmt.Printf("\t%s\n", err)
		prog.showReleases("\t", "\t\t")
		os.Exit(1)
	}
}

// applyRelease runs each of the files in the manifest in the specified
// order. If the file is in the SQL directory then it is applied with the
// standard SQL command directly. Otherwise the file is executed as a
// command itself. It reports any errors
func (prog *Prog) applyRelease() error {
	sqlPrefix := dbtcommon.DbtDirReleaseSQL(
		prog.dbp.BaseDirName, prog.releaseName)

	var cmd *exec.Cmd

	for _, f := range prog.fileList {
		if !prog.quiet {
			fmt.Println("running:", f)
		}
		if strings.HasPrefix(f, sqlPrefix) {
			cmd = dbtcommon.SQLCommand(prog.dbp, f)
		} else {
			cmd = exec.Command(f)
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("running %s: %s", f, err)
		}
	}
	return nil
}

// Prog holds parameter values etc
type Prog struct {
	quiet      bool
	noWarn     bool
	doNotApply bool

	releaseName string

	dbp *dbtcommon.DBParams

	manifestMap map[string]location.L
	fileList    []string
}

// NewProg returns a new Prog value, correctly initialised
func NewProg() *Prog {
	return &Prog{
		manifestMap: map[string]location.L{},
		dbp:         dbtcommon.NewDBParams(),
	}
}

func main() {
	prog := NewProg()
	ps := makeParamSet(prog)
	ps.Parse()

	prog.checkReleaseDir()

	if prog.doNotApply {
		prog.showReleases("", "\t")
		os.Exit(0)
	}

	prog.showReadMe()
	prog.showWarning()

	errors := prog.parseManifest()
	reportErrors(errors...)

	errors = prog.checkForUnusedFiles()
	reportErrors(errors...)

	err := prog.applyRelease()
	reportErrors(err)
}
