package updater

import (
	"os"
	"fmt"
	"path/filepath"

	"cord.stool/xdelta"
	"cord.stool/utils"

	"github.com/gosuri/uiprogress"
	"github.com/udhos/equalfile"
)

func CreateBinDiff(sourceOldDir string, sourceNewDir string, outputDiffDir string) error {

	fmt.Println("Making patch ...")
	uiprogress.Start()

	fullSourceOldDir, err := filepath.Abs(sourceOldDir)
	if err != nil {
		return err
	}

	fullSourceNewDir, err := filepath.Abs(sourceNewDir)
	if err != nil {
		return err
	}

	fullOutputDiffDir, err := filepath.Abs(outputDiffDir)
	if err != nil {
		return err
	}

	fc, err := utils.FileCount(fullSourceNewDir)
	if err != nil {
		return err
	}

	uiprogress.Start()
	barTotal := uiprogress.AddBar(fc + 1 ).AppendCompleted().PrependElapsed()
	barTotal.PrependFunc(func(b *uiprogress.Bar) string {
		return "Total progress"
	})

	stopCh := make(chan struct{})
	defer func() {
		select {
		case stopCh <- struct{}{}:
		default:
		}

		close(stopCh)
	}()

	os.RemoveAll(fullOutputDiffDir)

	f, e := utils.EnumFilesRecursive(fullSourceNewDir, stopCh)

	bar := uiprogress.AddBar(4).AppendCompleted().PrependElapsed()

	var curTitle string
	var title *string
	title = &curTitle

	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return *title
	})

	barTotal.Incr();

	for pathNewFile := range f {

		_, fn := filepath.Split(pathNewFile)
		curTitle = fmt.Sprint("Patching file: ", fn)

		barTotal.Incr();
		bar.Set(0);
		bar.Incr();

		relativePath, err := filepath.Rel(fullSourceNewDir, pathNewFile)
		if err != nil {
			return err
		}

		pathOldFile := filepath.Join(fullSourceOldDir, relativePath)
		pathDiffFile := filepath.Join(fullOutputDiffDir, relativePath)
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
				bar.Set(4);
				continue
			}
		}

		bar.Incr();

		pathDiff, _ := filepath.Split(pathDiffFile)
		if _, err := os.Stat(pathDiff); os.IsNotExist(err) {
			err = os.MkdirAll(pathDiff, 0777)
			if err != nil {
				return err
			}
		}

		bar.Incr();

		err = xdelta.EncodeDiff(pathOldFile, pathNewFile, pathDiffFile)
		if err != nil {
			return err
		}

		bar.Incr();
	}

	err = <-e
	if err != nil {
		return err
	}

	curTitle = "Finished"

	uiprogress.Stop()
	fmt.Println("Patch completed.")

	return nil
}
