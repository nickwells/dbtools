package dbtcommon

import (
	"fmt"
	"os"
	"path/filepath"
)

var pBits os.FileMode = 0o755

// DirSpec holds the details for a hierarchy of directories
type DirSpec struct {
	name          string
	ignoreContent bool
	subDirs       []DirSpec
}

// Various names of directories and files
const (
	DbtDirName = "db.postgres"

	ReleaseScriptsBaseName  = "releaseScripts"
	ReleaseArchiveDirName   = "Archive"
	ReleaseSQLDirName       = "SQL.files"
	ReleaseManifestFileName = "Manifest"
	ReleaseReadMeFileName   = "ReadMe"
	ReleaseWarningFileName  = "Warning"

	MacrosDirName   = "macros"
	DBSchemaDirName = "db.schema"

	SchemaSubDirTypes    = "types"
	SchemaSubDirTables   = "tables"
	SchemaSubDirFuncs    = "funcs"
	SchemaSubDirTriggers = "triggers"
)

var dirHierarchy = []DirSpec{
	{
		name: DbtDirName,
		subDirs: []DirSpec{
			{
				name: ReleaseScriptsBaseName,
				subDirs: []DirSpec{
					{
						name:          ReleaseArchiveDirName,
						ignoreContent: true,
					},
				},
			},
			{
				name:          MacrosDirName,
				ignoreContent: true,
			},
			{
				name:          DBSchemaDirName,
				ignoreContent: true,
			},
		},
	},
}

var schemaDirs = []DirSpec{
	{
		name:          SchemaSubDirTypes,
		ignoreContent: true,
	},
	{
		name:          SchemaSubDirTables,
		ignoreContent: true,
	},
	{
		name:          SchemaSubDirFuncs,
		ignoreContent: true,
	},
	{
		name:          SchemaSubDirTriggers,
		ignoreContent: true,
	},
}

// DbtDirStart returns the name of the starting directory
func DbtDirStart(basename string) string {
	return filepath.Join(basename, DbtDirName)
}

// DbtDirMacros returns the fullname of the macros directory
func DbtDirMacros(basename string) string {
	return filepath.Join(DbtDirStart(basename), MacrosDirName)
}

// DbtDirDBSchemaBase returns the full base name of the DB.schema directories
func DbtDirDBSchemaBase(basename string) string {
	return filepath.Join(DbtDirStart(basename), DBSchemaDirName)
}

// DbtDirDBSchema returns the full name of the directory for the given database
// and schema
func DbtDirDBSchema(basename, dbName, schemaName string) string {
	return filepath.Join(DbtDirDBSchemaBase(basename), dbName+"."+schemaName)
}

// DbtDirReleaseBase returns the full name of the release scripts directory
func DbtDirReleaseBase(basename string) string {
	return filepath.Join(DbtDirStart(basename), ReleaseScriptsBaseName)
}

// DbtDirRelease returns the full name of the release directory
func DbtDirRelease(basename, rel string) string {
	return filepath.Join(DbtDirReleaseBase(basename), rel)
}

// DbtDirReleaseSQL returns the full name of the release SQL.files directory
func DbtDirReleaseSQL(basename, rel string) string {
	return filepath.Join(DbtDirReleaseBase(basename), rel, ReleaseSQLDirName)
}

// DbtFileReleaseManifest returns the full name of the release manifest file
func DbtFileReleaseManifest(basename, rel string) string {
	return filepath.Join(DbtDirRelease(basename, rel), ReleaseManifestFileName)
}

// DbtFileReleaseReadMe returns the full name of the release ReadMe file
func DbtFileReleaseReadMe(basename, rel string) string {
	return filepath.Join(DbtDirRelease(basename, rel), ReleaseReadMeFileName)
}

// DbtFileReleaseWarning returns the full name of the release Warning file
func DbtFileReleaseWarning(basename, rel string) string {
	return filepath.Join(DbtDirRelease(basename, rel), ReleaseWarningFileName)
}

// checkSubDirs recursively checks the dirs exist in base
func checkSubDirs(base string, dirs []DirSpec) bool {
	for _, d := range dirs {
		dirName := filepath.Join(base, d.name)

		info, err := os.Stat(dirName)
		if err != nil {
			return false
		}

		if !info.Mode().IsDir() {
			return false
		}

		if d.ignoreContent {
			continue
		}

		if !checkSubDirs(dirName, d.subDirs) {
			return false
		}
	}

	return true
}

// CheckDirs confirms that the necessary directories are present
func CheckDirs(basename, dbName, schemaName string) bool {
	if !checkSubDirs(basename, dirHierarchy) {
		return false
	}

	return checkSubDirs(DbtDirDBSchema(basename, dbName, schemaName), schemaDirs)
}

// makeDirIfMissing will create a directory if it is not present and will
// report any errors on the way
func makeDirIfMissing(dirName string) error {
	info, err := os.Stat(dirName)
	if os.IsNotExist(err) {
		err = os.Mkdir(dirName, pBits)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if !info.Mode().IsDir() {
		return fmt.Errorf("Couldn't create the directory %q"+
			" - it already exists and is not a directory",
			dirName)
	}

	return nil
}

// makeMissingSubDirs recursively makes the dirs missing from base. It stops
// at the first error
func makeMissingSubDirs(base string, dirs []DirSpec) error {
	for _, d := range dirs {
		dirName := filepath.Join(base, d.name)

		err := makeDirIfMissing(dirName)
		if err != nil {
			return err
		}

		if !d.ignoreContent {
			if err = makeMissingSubDirs(dirName, d.subDirs); err != nil {
				return err
			}
		}
	}

	return nil
}

// MakeMissingDirs this will make any directories that are needed
// and not present. There can be errors if the process doesn't have
// the right permissions, if there is a file-system object such as
// a file that is masking the directory, the file system is full
// etc. The attempt will stop at the first error
func MakeMissingDirs(basename, dbName, schemaName string) error {
	err := makeMissingSubDirs(basename, dirHierarchy)
	if err != nil {
		return err
	}

	dirName := DbtDirDBSchemaBase(basename)

	err = makeDirIfMissing(dirName)
	if err != nil {
		return err
	}

	dirName = DbtDirDBSchema(basename, dbName, schemaName)

	err = makeDirIfMissing(dirName)
	if err != nil {
		return err
	}

	return makeMissingSubDirs(dirName, schemaDirs)
}
