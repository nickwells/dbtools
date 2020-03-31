// dbt_apply_changes
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nickwells/cli.mod/cli/responder"
	"github.com/nickwells/dbtcommon.mod/dbtcommon"
	"github.com/nickwells/param.mod/v4/param"
	"github.com/nickwells/param.mod/v4/param/paramset"
)

// Created: Wed Apr 12 21:29:46 2017

var quiet bool
var noWarn bool
var showReleasesFlag bool

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

// printFile prints the file if it exists and is not empty and returns a
// boolean to indicate whether anything was printed
func printFile(fileName, title, sep string) bool {
	fStat, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		fmt.Printf("Couldn't open the %s file: %s", fileName, err)
		os.Exit(1)
	}

	if fStat.Mode()&os.ModeType != 0 {
		fmt.Printf("%q exists but it is not a regular file", fileName)
		os.Exit(1)
	}

	fd, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Couldn't open %s: %s\n", fileName, err)
		os.Exit(1)
	}
	defer fd.Close()

	headerPrinted := false
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		if !headerPrinted {
			printFileHeader(title, sep)
			headerPrinted = true
		}
		line := scanner.Text()
		fmt.Println(line)
	}
	if err = scanner.Err(); err != nil {
		fmt.Printf("%s While reading %s: %s\n", errorPrefix, fileName, err)
		os.Exit(1)
	}

	if headerPrinted {
		fmt.Print(sep)
		fmt.Println()

		return true
	}
	return false
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

	if printFile(dbtcommon.DbtFileReleaseWarning(rel),
		"Warning", "#################################################\n") {
		if r.GetResponseOrDie() == 'y' {
			return
		}
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
			cmd = dbtcommon.SqlCommand(f)
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
