// +build linux

package utils

import (
	"os"
)
// CompleteUpgrade ...
func CompleteUpgrade(fnames []string) error {

	for _, f := range fnames {
		os.Remove(f)	
	}
	return nil
}
