package cord

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"cord.stool/cordapi"
	"cord.stool/service/models"

	"github.com/itchio/savior/seeksource"
	"github.com/itchio/wharf/eos"
	"github.com/itchio/wharf/eos/option"
	"github.com/itchio/wharf/pools"
	"github.com/itchio/wharf/pwr"
	"github.com/itchio/wharf/state"
	"github.com/itchio/wharf/tlc"
	"github.com/itchio/wharf/wsync"
)

var ignoredPaths = []string{
	".git",
	".hg",
	".svn",
	".DS_Store",
	"__MACOSX",
	"._*",
	"Thumbs.db",
}

func filterPaths(fileInfo os.FileInfo) bool {
	name := fileInfo.Name()
	for _, pattern := range ignoredPaths {
		match, _ := filepath.Match(pattern, name)
		if match {
			return false
		}
	}

	return true
}

func compressionSettings() *pwr.CompressionSettings {

	return &pwr.CompressionSettings{
		Algorithm: pwr.CompressionAlgorithm_GZIP,
		Quality:   1,
	}
}

func newStateConsumer() *state.Consumer {
	return &state.Consumer{
		OnProgress:       Progress,
		OnProgressLabel:  ProgressLabel,
		OnPauseProgress:  PauseProgress,
		OnResumeProgress: ResumeProgress,
		OnMessage:        Logl,
	}
}

func ProgressLabel(label string) {

	_, fn := filepath.Split(label)
	_curTitle = fmt.Sprint("Creating patch: ", fn)
}

func PauseProgress() {
}

func ResumeProgress() {
}

func Progress(alpha float64) {

	_bar.Set(int(100 * alpha))
	_barTotal.Set(int(5*alpha) + (_barTotal.Total - 7))
}

func Logl(level string, msg string) {
}

func uploadWharf(api *cordapi.CordAPIManager, args Args, source string, manifest models.ConfigManifest) error {

	_bar.Total = 5
	_curTitle = "Getting files' info from server"

	signatureInfo, err := getSignatureInfo(api, args.SrcBuildID, manifest.Platform)
	if err != nil {
		return err
	}

	_barTotal.Incr()
	_curTitle = "Checking changed files"

	source = filepath.Join(source, manifest.LocalRoot)

	err = pwr.AssertValid(source, signatureInfo)
	_bar.Incr()
	if err == nil {
		return errors.New("No changes and not pushing anything")
	}

	_barTotal.Incr()
	_bar.Set(0)
	_bar.Total = 101
	_curTitle = "Creating patch"

	var sourceContainer *tlc.Container
	sourceContainer, err = tlc.WalkAny(source, &tlc.WalkOpts{Filter: filterPaths})
	if err != nil {
		return errors.New("Walking source as directory failed: " + err.Error())
	}

	var sourcePool wsync.Pool
	sourcePool, err = pools.New(sourceContainer, source)
	if err != nil {
		return errors.New("Walking source as directory failed: " + err.Error())
	}

	patchFile, err := ioutil.TempFile(os.TempDir(), "patch")
	if err != nil {
		return errors.New("Cannot get temp file, error: " + err.Error())
	}
	defer os.Remove(patchFile.Name())
	patchFile.Close()

	patchWriter, err := os.Create(patchFile.Name())
	if err != nil {
		return errors.New("Cannot create patch file, error: " + err.Error())
	}
	defer patchWriter.Close()

	signFile, err := ioutil.TempFile(os.TempDir(), "sign")
	if err != nil {
		return errors.New("Cannot get temp file, error: " + err.Error())
	}
	defer os.Remove(signFile.Name())
	signFile.Close()

	signatureWriter, err := os.Create(signFile.Name())
	if err != nil {
		return errors.New("Cannot create sign file, error: " + err.Error())
	}
	defer signatureWriter.Close()

	dctx := &pwr.DiffContext{
		SourceContainer: sourceContainer,
		Pool:            sourcePool,

		TargetContainer: signatureInfo.Container,
		TargetSignature: signatureInfo.Hashes,

		Consumer:    newStateConsumer(),
		Compression: compressionSettings(),
	}

	_bar.Incr()
	_barTotal.Incr()

	err = dctx.WritePatch(context.Background(), patchWriter, signatureWriter)
	if err != nil {
		return errors.New("Computing and writing patch and signature failed: " + err.Error())
	}

	_barTotal.Set(_barTotal.Total - 1)

	return applyPatch(api, args.BuildID, args.SrcBuildID, patchFile.Name(), manifest.Platform)
}

func applyPatch(api *cordapi.CordAPIManager, buildID string, srcbuildID string, patch string, platform string) error {

	_bar.Set(0)
	_bar.Total = 3
	_curTitle = "Applying patch on server"

	filedata, err := ioutil.ReadFile(patch)
	if err != nil {
		return errors.New("Cannot read file: " + err.Error())
	}

	_bar.Incr()

	err = api.ApplyPatch(&models.ApplyPatchCmd{BuildID: buildID, SrcBuildID: srcbuildID, FileData: filedata, Platform: platform})
	if err != nil {
		return errors.New("Applying patch failed: " + err.Error())
	}

	_bar.Set(_bar.Total)
	_barTotal.Incr()

	return nil
}

func getSignatureInfo(api *cordapi.CordAPIManager, buildID string, platform string) (*pwr.SignatureInfo, error) {

	singRes, err := api.GetSignature(buildID, platform)
	if err != nil {
		return nil, err
	}

	_bar.Incr()

	singFile, err := ioutil.TempFile(os.TempDir(), "sign")
	if err != nil {
		return nil, errors.New("Cannot get temp file, error: " + err.Error())
	}
	defer os.Remove(singFile.Name())
	singFile.Close()

	err = ioutil.WriteFile(singFile.Name(), singRes.FileData, 0777)
	if err != nil {
		return nil, errors.New("Cannot write to patch, error: " + err.Error())
	}
	singFile.Close()

	_bar.Incr()

	sigReader, err := eos.Open(singFile.Name(), option.WithConsumer(newStateConsumer()))
	if err != nil {
		return nil, errors.New("Cannot open sign file, error: " + err.Error())
	}
	defer sigReader.Close()

	sigSource := seeksource.FromFile(sigReader)
	_, err = sigSource.Resume(nil)
	if err != nil {
		return nil, errors.New("Creating source for signature failed: " + err.Error())
	}

	_bar.Incr()

	signatureInfo, err := pwr.ReadSignature(context.Background(), sigSource)
	if err != nil {
		return nil, errors.New("Decoding signature failed: " + err.Error())
	}

	_bar.Incr()

	return signatureInfo, nil
}
