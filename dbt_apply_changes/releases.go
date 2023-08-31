package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/nickwells/dbtools/internal/dbtcommon"
	"github.com/nickwells/fileparse.mod/fileparse"
	"github.com/nickwells/location.mod/location"
)

// findReleases finds all the non-archived releases in the release directory
func (prog *Prog) findReleases() ([]string, error) {
	dir, err := os.Open(dbtcommon.DbtDirReleaseBase(prog.dbp.BaseDirName))
	if err != nil {
		return nil, err
	}

	contents, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	ignoreEntry := map[string]bool{
		".":                             true,
		"..":                            true,
		dbtcommon.ReleaseArchiveDirName: true,
	}

	relDirs := make([]string, 0)
	for _, entry := range contents {
		if ignoreEntry[entry.Name()] || !entry.IsDir() {
			continue
		}
		relDirs = append(relDirs, entry.Name())
	}
	sort.Strings(relDirs)

	return relDirs, nil
}

type manifestFileParser struct {
	fileMap    map[string]location.L
	fileList   *[]string
	releaseDir string
}

// (mfp *manifestFileParser) ParseLine parses a line from the manifest file
func (mfp *manifestFileParser) ParseLine(line string, loc *location.L) error {
	file := filepath.Join(mfp.releaseDir, line)

	fStat, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return loc.Errorf(
				"The release directory (%s) does not contain %q",
				mfp.releaseDir, line)
		}
		return loc.Error(err.Error())
	}

	if !fStat.Mode().IsRegular() {
		return loc.Errorf("The release directory (%s) contains %q"+
			" but it is not a regular file",
			mfp.releaseDir, line)
	}

	if prevLoc, ok := mfp.fileMap[line]; ok {
		return loc.Errorf("The file is already in the manifest at: %s",
			prevLoc)
	}
	mfp.fileMap[line] = *loc
	*mfp.fileList = append(*mfp.fileList, file)

	return nil
}

// parseManifest reads the Manifest file and constructs a map of files and
// a list of the files in the order they appear in the Manifest file. The
// files are checked to make sure they exist and an error is generated if
// they don't. A duplicate entry is also an error
func (prog *Prog) parseManifest() []error {
	manifest := dbtcommon.DbtFileReleaseManifest(
		prog.dbp.BaseDirName, prog.releaseName)
	relDir := dbtcommon.DbtDirRelease(
		prog.dbp.BaseDirName, prog.releaseName)
	var errors []error

	mfStat, err := os.Stat(manifest)
	if err != nil {
		errors = append(errors, err)
		if os.IsNotExist(err) {
			errors = append(errors,
				fmt.Errorf(
					"The release directory (%s) does not contain a"+
						" file called %q. This lists the release"+
						" files to apply and the order in which they"+
						" should be applied",
					relDir, dbtcommon.ReleaseManifestFileName))
		}
		return errors
	}

	if !mfStat.Mode().IsRegular() {
		errors = append(errors,
			fmt.Errorf(
				"The release directory (%s) contains %q but it is"+
					" not a regular file",
				relDir, dbtcommon.ReleaseManifestFileName))
		return errors
	}

	mfp := manifestFileParser{
		fileMap:    prog.manifestMap,
		fileList:   &prog.fileList,
		releaseDir: relDir,
	}

	fp := fileparse.New("Manifest", &mfp)
	fp.SetCommentIntro("#")
	fp.SetInclKeyWord("")
	errors = append(errors, fp.Parse(manifest)...)

	if len(prog.fileList) == 0 {
		errors = append(errors,
			fmt.Errorf(
				"the manifest is empty - all the lines are empty or comments"))
	}

	return errors
}

// checkForUnusedFiles checks that all the files in the release dir are
// referenced in the manifest file
func (prog *Prog) checkForUnusedFiles() []error {
	errors := make([]error, 0)

	relDir := dbtcommon.DbtDirRelease(prog.dbp.BaseDirName, prog.releaseName)
	dir, err := os.Open(relDir)
	if err != nil {
		return append(errors, err)
	}

	contents, err := dir.Readdir(-1)
	if err != nil {
		return append(errors, err)
	}

	ignoreEntry := map[string]bool{
		".":                               true,
		"..":                              true,
		dbtcommon.ReleaseSQLDirName:       true,
		dbtcommon.ReleaseManifestFileName: true,
		dbtcommon.ReleaseReadMeFileName:   true,
		dbtcommon.ReleaseWarningFileName:  true,
	}

	for _, entry := range contents {
		if ignoreEntry[entry.Name()] {
			continue
		}
		if _, ok := prog.manifestMap[entry.Name()]; !ok {
			errors = append(errors,
				fmt.Errorf("the release directory (%s) contains %q"+
					" which is not in the Manifest file",
					relDir, entry.Name()))
		}
	}

	return errors
}

// releaseDirIsOK checks that the release directory exists and returns an
// error if it does not
func (prog *Prog) releaseDirIsOK() error {
	if prog.releaseName == dbtcommon.ReleaseArchiveDirName {
		return fmt.Errorf(
			"the %s directory cannot be used as a release directory",
			prog.releaseName)
	}

	relDir := dbtcommon.DbtDirRelease(prog.dbp.BaseDirName, prog.releaseName)
	rdStat, err := os.Stat(relDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("The release directory (%s) does not exist",
				relDir)
		}
		return err
	}

	if !rdStat.IsDir() {
		return fmt.Errorf("%s is not a directory", relDir)
	}

	return nil
}
