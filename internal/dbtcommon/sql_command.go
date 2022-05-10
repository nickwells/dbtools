package dbtcommon

import "os/exec"

// SQLCommand returns the command to be run
func SQLCommand(fileName string) *exec.Cmd {
	return exec.Command(PsqlPath,
		"-v", "ON_ERROR_STOP=1",
		"-q",
		"-d", DbName,
		"-f", fileName)
}
