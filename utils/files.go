package utils

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func ExpandHome(input string) string {
	if strings.HasPrefix(input, "~") {
		return strings.Replace(input, "~", os.Getenv("HOME"), 1)
	}
	return input
}

func Open(path string) ([]byte, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", path)
	} else if runtime.GOOS == "linux" {
		cmd = exec.Command("xdg-open", path)
	} 
	out, err := cmd.Output()
	return out, err
}
