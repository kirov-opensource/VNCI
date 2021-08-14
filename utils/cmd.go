package utils

import (
	"os/exec"

	"github.com/google/logger"
)

func RunCmd(name string, args ...string) bool {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Fatalln("cmd.Run() failed with %s\n", err)
		return false
	}
	logger.Info(out)
	return true
}
