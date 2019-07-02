package update

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"cord.stool/utils"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
)

func CreatePatch(sourceOldDir string, sourceNewDir string, patchFile string) error {

	fmt.Println("Making patch ...")

	fullSourceOldDir, err := filepath.Abs(sourceOldDir)
	if err != nil {
		return err
	}

	fullSourceNewDir, err := filepath.Abs(sourceNewDir)
	if err != nil {
		return err
	}

	fullPatchFile, err := filepath.Abs(patchFile)
	if err != nil {
		return err
	}

	uiprogress.Start()
	_barTotal = uiprogress.AddBar(4).AppendCompleted().PrependElapsed()
	_barTotal.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(_totalTitle, 35)
	})
	_totalTitle = "Total progress"

	_bar = uiprogress.AddBar(100).AppendCompleted().PrependElapsed()
	_bar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(_curTitle, 35)
	})

	_curTitle = "Computing signature ..."

	signFile, err := ioutil.TempFile(os.TempDir(), "sign")
	if err != nil {
		return err
	}
	defer os.Remove(signFile.Name())
	signFile.Close()

	_barTotal.Incr()

	err = utils.CreateSignatureFile(fullSourceOldDir, signFile.Name(), newStateConsumer())
	if err != nil {
		return err
	}

	_barTotal.Incr()

	signInfo, err := utils.GetSignatureInfoFromFile(signFile.Name(), newStateConsumer())
	if err != nil {
		return err
	}

	_barTotal.Incr()

	_curTitle = "Creating patch ..."
	_bar.Set(0)

	err = utils.CreatePatchFile(fullSourceNewDir, fullPatchFile, signInfo, newStateConsumer())
	if err != nil {
		return err
	}

	_barTotal.Incr()

	_totalTitle = "Finished"

	uiprogress.Stop()
	fmt.Println("Patch completed.")

	return nil
}
