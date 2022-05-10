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
	"github.com/nickwells/param.mod/v5/param"
	"github.com/nickwells/param.mod/v5/param/paramset"
)

// Created: Wed Apr 12 21:29:46 2017

type action int

const (
	doNothing action = iota
	confirm
	abort
)

var (
	quiet            bool
	noWarn           bool
	showReleasesFlag bool
)

var releaseName string

const errorPrefix = "*** Error ***"

// reportErrors checks if there are any errors and if so prints them and exits
func reportErrors(errors ...error) {
	if len(errors) != 0 {
		for _, err := range errors {
			if err != nil {
				fmt.Println(errorPrefix, err)
			}
		}
		os.Exit(1)
	}
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
// boolean to indicate whether anything was printed
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

	if fStat.Mode()&os.ModeType != 0 {
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
func showReadMe(rel string) {
	if quiet {
		return
	}

	printFile(dbtcommon.DbtFileReleaseReadMe(rel),
		"Note", "========================================\n")
}

// showWarning prints the Warnings file (if any) unless the noWarn flag has
// been set
func showWarning(rel string) {
	if noWarn {
		return
	}

	r := responder.NewOrPanic(
		"Do you want to continue",
		map[rune]string{
			'y': "apply the changes",
			'n': "abort the changes",
		},
		responder.SetMaxReprompts(5))

	switch printFile(dbtcommon.DbtFileReleaseWarning(rel),
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
func checkReleaseDir() {
	err := releaseDirExists(releaseName)
	if err != nil {
		fmt.Printf("%s Bad release: %s\n", errorPrefix, releaseName)
		fmt.Printf("\t%s\n", err)
		releases, err := findReleases()
		if err != nil {
			fmt.Printf("\t%s\n", err)
		} else if len(releases) == 0 {
			fmt.Println("\tThere are no releases to apply")
		} else {
			fmt.Println("\tPossible releases are:")
			for _, r := range releases {
				fmt.Println("\t\t", r)
			}
		}
		os.Exit(1)
	}
}

// applyRelease runs each of the files in the manifest in the specified
// order. If the file is in the SQL directory then it is applied with the
// standard SQL command directly. Otherwise the file is executed as a
// command itself. It reports any errors
func applyRelease(fileList []string) error {
	sqlPrefix := dbtcommon.DbtDirReleaseSQL(releaseName)

	var cmd *exec.Cmd

	for _, f := range fileList {
		if !quiet {
			fmt.Println("running:", f)
		}
		if strings.HasPrefix(f, sqlPrefix) {
			cmd = dbtcommon.SQLCommand(f)
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

func main() {
	ps := paramset.NewOrDie(
		addParams,
		dbtcommon.AddParams,
		param.SetProgramDescription("this will apply a set of scripts"+
			" (typically shell scripts but any executable can be run)."+
			" The contents of the release directory is checked against"+
			" a Manifest file which also defines the order in which"+
			" they should be applied"))
	ps.Parse()

	checkReleaseDir()

	if showReleasesFlag {
		showReleases()
		os.Exit(0)
	}

	showReadMe(releaseName)
	showWarning(releaseName)

	manifestMap, fileList, errors := parseManifest(releaseName)
	reportErrors(errors...)

	errors = checkForUnusedFiles(releaseName, manifestMap)
	reportErrors(errors...)

	err := applyRelease(fileList)
	reportErrors(err)
}
