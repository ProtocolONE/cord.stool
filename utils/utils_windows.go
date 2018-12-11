// +build windows

package utils

import (
	"os"
)

// CompleteUpgrade ...
func CompleteUpgrade(fnames []string) error {

	var args []string
	args = append(args, os.Args[0])
	args = append(args, "upgrade_complete")

	for _, f := range fnames {
		args = append(args, "-f")
		args = append(args, f)
	}

	procAttr := new(os.ProcAttr)
	procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}

	os.StartProcess(os.Args[0], args, procAttr)

	return nil
}
