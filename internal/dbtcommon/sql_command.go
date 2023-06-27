package dbtcommon

import "os/exec"

// SQLCommand returns the command to be run. The command is the sql runner
// (psql) with various standard flags applied and running the given filename.
func SQLCommand(dbp *DBParams, fileName string) *exec.Cmd {
	return exec.Command(dbp.PsqlPath,
		"-v", "ON_ERROR_STOP=1",
		"-q",
		"-d", dbp.DbName,
		"-f", fileName)
}
