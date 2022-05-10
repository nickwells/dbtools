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
func findReleases() ([]string, error) {
	dir, err := os.Open(dbtcommon.DbtDirReleaseBase())
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

	if fStat.Mode()&os.ModeType != 0 {
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
func parseManifest(rel string) (map[string]location.L, []string, []error) {
	manifest := dbtcommon.DbtFileReleaseManifest(rel)
	relDir := dbtcommon.DbtDirRelease(rel)
	errors := make([]error, 0)

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
		return nil, nil, errors
	}

	if mfStat.Mode()&os.ModeType != 0 {
		errors = append(errors,
			fmt.Errorf(
				"The release directory (%s) contains %q but it is"+
					" not a regular file",
				relDir, dbtcommon.ReleaseManifestFileName))
		return nil, nil, errors
	}

	manifestMap := make(map[string]location.L)
	fileList := make([]string, 0)

	mfp := manifestFileParser{
		fileMap:    manifestMap,
		fileList:   &fileList,
		releaseDir: relDir,
	}

	fp := fileparse.New("Manifest", &mfp)
	fp.SetCommentIntro("#")
	fp.SetInclKeyWord("")
	errors = append(errors, fp.Parse(manifest)...)

	if len(fileList) == 0 {
		errors = append(errors,
			fmt.Errorf(
				"the manifest is empty - all the lines are empty or comments"))
	}

	return manifestMap, fileList, errors
}

// checkForUnusedFiles checks that all the files in the release dir are
// referenced in the manifest file
func checkForUnusedFiles(rel string, manifestMap map[string]location.L) []error {
	errors := make([]error, 0)

	relDir := dbtcommon.DbtDirRelease(rel)
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
		if _, ok := manifestMap[entry.Name()]; !ok {
			errors = append(errors,
				fmt.Errorf("the release directory (%s) contains %q"+
					" which is not in the Manifest file",
					relDir, entry.Name()))
		}
	}

	return errors
}

// releaseDirExists checks that the release directory exists and returns an
// error if it does not
func releaseDirExists(rel string) error {
	if rel == dbtcommon.ReleaseArchiveDirName {
		return fmt.Errorf(
			"the %s directory cannot be used as a release directory", rel)
	}

	relDir := dbtcommon.DbtDirRelease(rel)
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
