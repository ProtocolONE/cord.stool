package controllers

import (
	"cord.stool/service/core/utils"
	"cord.stool/service/models"

	context2 "context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/itchio/wharf/eos"
	"github.com/itchio/wharf/pools"
	"github.com/itchio/wharf/pwr"
	"github.com/itchio/wharf/tlc"
	"github.com/itchio/wharf/wire"
	"github.com/itchio/wharf/wsync"
	//"github.com/itchio/wharf/state"
	//"github.com/itchio/wharf/eos/option"
	"github.com/itchio/savior/seeksource"

	"github.com/labstack/echo"
)

var ignoredPaths = []string{
	".git",
	".hg",
	".svn",
	".DS_Store",
	"__MACOSX",
	"._*",
	"Thumbs.db",
	".itch",
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

func SignatureCmd(context echo.Context) error {

	userRoot, err := utils.GetUserStorage(context.Request().Header.Get("ClientID"))
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorGetUserStorage, err.Error()})
	}

	pathParam := context.QueryParam("path")
	if pathParam == "" {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorInvalidRequest, "Invalid request params: " + err.Error()})
	}

	fpath := path.Join(userRoot, pathParam)

	container, err := tlc.WalkAny(fpath, &tlc.WalkOpts{Filter: filterPaths})
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorWharfLibrary, "Walking directory to sign failed: " + err.Error()})
	}

	pool, err := pools.New(container, fpath)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorWharfLibrary, "Creating pool for directory to sign failed: " + err.Error()})
	}

	signFile, err := ioutil.TempFile(os.TempDir(), "sign")
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorGenTempFile, fmt.Sprintf("Cannot get temp file, error: %s", err.Error())})
	}
	defer os.Remove(signFile.Name())
	signFile.Close()

	signatureWriter, err := os.Create(signFile.Name())
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorCreateFile, fmt.Sprintf("Cannot create signature file, error: %s", err.Error())})
	}
	defer signatureWriter.Close()

	rawSigWire := wire.NewWriteContext(signatureWriter)
	rawSigWire.WriteMagic(pwr.SignatureMagic)
	rawSigWire.WriteMessage(&pwr.SignatureHeader{Compression: compressionSettings()})

	sigWire, err := pwr.CompressWire(rawSigWire, compressionSettings())
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorWharfLibrary, "Setting up compression for signature file failed: " + err.Error()})
	}

	sigWire.WriteMessage(container)

	err = pwr.ComputeSignatureToWriter(context2.Background(), container, pool, nil, func(hash wsync.BlockHash) error {
		return sigWire.WriteMessage(&pwr.BlockHash{
			WeakHash:   hash.WeakHash,
			StrongHash: hash.StrongHash,
		})
	})

	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorWharfLibrary, "Computing signature failed: " + err.Error()})
	}

	err = sigWire.Close()
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorWharfLibrary, "Finalizing signature file failed: " + err.Error()})
	}

	signData, err := ioutil.ReadFile(signFile.Name())
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorReadFile, fmt.Sprintf("Cannot read signature file, error: %s", err.Error())})
	}

	return context.JSON(http.StatusOK, models.SignatureCmdResult{FileData: signData})
}

func ApplyPatchCmd(context echo.Context) error {

	reqCmp := &models.ApplyPatchCmd{}
	err := context.Bind(reqCmp)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorInvalidJSONFormat, "Invalid JSON format: " + err.Error()})
	}

	userRoot, err := utils.GetUserStorage(context.Request().Header.Get("ClientID"))
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorGetUserStorage, err.Error()})
	}

	fpath := path.Join(userRoot, reqCmp.Path)

	patchFile, err := ioutil.TempFile(os.TempDir(), "patch")
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorGenTempFile, fmt.Sprintf("Cannot get temp file, error: %s", err.Error())})
	}
	defer os.Remove(patchFile.Name())
	patchFile.Close()

	err = ioutil.WriteFile(patchFile.Name(), reqCmp.FileData, 0777)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorWriteFile, fmt.Sprintf("Cannot write to patch, error: %s", err.Error())})
	}
	patchFile.Close()

	patchReader, err := eos.Open(patchFile.Name())
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorOpenFile, fmt.Sprintf("Cannot open patch file, error: %s", err.Error())})
	}

	actx := &pwr.ApplyContext{
		TargetPath: fpath,
		OutputPath: fpath,
		DryRun:     false,
		InPlace:    true,
		Signature:  nil,
		WoundsPath: "",
		StagePath:  "",
		HealPath:   "",

		Consumer: nil, //newStateConsumer(),
	}

	patchSource := seeksource.FromFile(patchReader)

	_, err = patchSource.Resume(nil)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorWharfLibrary, "Resuming patch file failed: " + err.Error()})
	}

	err = actx.ApplyPatch(patchSource)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorWharfLibrary, "Applying patch file failed: " + err.Error()})
	}

	return context.NoContent(http.StatusOK)
}
