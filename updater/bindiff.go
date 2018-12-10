package updater

import (
	"os"
	"path/filepath"

	"cord.stool/xdelta"
	"cord.stool/utils"

	"github.com/udhos/equalfile"
)

func CreateBinDiff(SourceOldDir string, SourceNewDir string, OutputDiffDir string) error {

	stopCh := make(chan struct{})
	defer func() {
		select {
		case stopCh <- struct{}{}:
		default:
		}

		close(stopCh)
	}()

	os.RemoveAll(OutputDiffDir)

	f, e := utils.EnumFilesRecursive(SourceNewDir, stopCh)

	for pathNewFile := range f {
		relativePath, err := filepath.Rel(SourceNewDir, pathNewFile)
		if err != nil {
			return err
		}

		pathOldFile := filepath.Join(SourceOldDir, relativePath)
		pathDiffFile := filepath.Join(OutputDiffDir, relativePath)
		pathDiffFile += ".diff"

		if _, err := os.Stat(pathOldFile); os.IsNotExist(err) { // the file is not exist

			pathOldFile = "NUL" // fake name

		} else { // the file is exist, try to compare

			cmp := equalfile.New(nil, equalfile.Options{})
			equal, err := cmp.CompareFile(pathOldFile, pathNewFile)
			if err != nil {
				return err
			}

			if equal {
				continue
			}
		}

		pathDiff, _ := filepath.Split(pathDiffFile)
		if _, err := os.Stat(pathDiff); os.IsNotExist(err) {
			err = os.MkdirAll(pathDiff, 0777)
			if err != nil {
				return err
			}
		}
		err = xdelta.EncodeDiff(pathOldFile, pathNewFile, pathDiffFile)
		if err != nil {
			return err
		}
	}

	err := <-e
	if err != nil {
		return err
	}

	return nil
}
