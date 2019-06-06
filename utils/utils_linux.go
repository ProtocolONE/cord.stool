// +build linux

package utils

import (
	"os"
	"os/exec"
	
	"cord.stool/service/models"
)

// CompleteUpgrade ...
func CompleteUpgrade(fnames []string) error {

	for _, f := range fnames {
		os.Remove(f)
	}
	return nil
}

func AddRegKeys(manifest *models.ConfigManifest) error {

	return nil
}

func CheckCompletion(regKey models.ConfigRegistryKey) (bool, error) {

	return true, nil
}

func RunCommand(admin bool, name string, arg ...string) error {

	return exec.Command(name, arg...).Run()
}
