package update

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	//"time"
	"os"

	"cord.stool/utils"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
	//"github.com/itchio/wharf/state"
)

/*var _bar *uiprogress.Bar
var _barTotal *uiprogress.Bar
var _curTitle string
var _totalTitle string

func newStateConsumer() *state.Consumer {
	return &state.Consumer{
		OnProgress:       progress,
		OnProgressLabel:  progressLabel,
		OnPauseProgress:  pauseProgress,
		OnResumeProgress: resumeProgress,
		OnMessage:        logl,
	}
}

func progressLabel(label string) {
}

func pauseProgress() {
}

func resumeProgress() {
}

func progress(alpha float64) {

	_bar.Set(int(100 * alpha))
	_barTotal.Set(int(5*alpha) + (_barTotal.Total - 7))
}

func logl(level string, msg string) {
}*/

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

	//time.Sleep(2 * time.Second)
	//_bar.Incr()
	_barTotal.Incr()

	err = utils.CreateSignatureFile(fullSourceOldDir, signFile.Name(), newStateConsumer())
	if err != nil {
		return err
	}

	//time.Sleep(2 * time.Second)
	//_bar.Incr()
	_barTotal.Incr()

	signInfo, err := utils.GetSignatureInfoFromFile(signFile.Name(), newStateConsumer())
	if err != nil {
		return err
	}

	//time.Sleep(2 * time.Second)
	//_bar.Incr()
	_barTotal.Incr()

	_curTitle = "Creating patch ..."
	_bar.Set(0)
	//_bar.Total = 1

	err = utils.CreatePatchFile(fullSourceNewDir, fullPatchFile, signInfo, newStateConsumer())
	if err != nil {
		return err
	}

	//time.Sleep(2 * time.Second)
	//_bar.Incr()
	_barTotal.Incr()

	_totalTitle = "Finished"

	uiprogress.Stop()
	fmt.Println("Patch completed.")

	return nil
}
