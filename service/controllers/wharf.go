package controllers

import (
	"cord.stool/service/core/utils"
	"cord.stool/service/models"

	context2 "context"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/itchio/savior/seeksource"
	"github.com/itchio/wharf/eos"
	"github.com/itchio/wharf/pools"
	"github.com/itchio/wharf/pwr"
	"github.com/itchio/wharf/tlc"
	"github.com/itchio/wharf/wire"
	"github.com/itchio/wharf/wsync"

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

	gameId := context.QueryParam("gameId")
	if gameId == "" {
		return utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "gameId is required")
	}

	branch := context.QueryParam("branch")
	if branch == "" {
		return utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "branch is required")
	}

	fpath, err := utils.GetUserBuildPath(context.Request().Header.Get("ClientID"), gameId, branch)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorGetUserStorage, err.Error())
	}

	container, err := tlc.WalkAny(fpath, &tlc.WalkOpts{Filter: filterPaths})
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorWharfLibrary, err.Error())
	}

	pool, err := pools.New(container, fpath)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorWharfLibrary, err.Error())
	}

	signFile, err := ioutil.TempFile(os.TempDir(), "sign")
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorWharfLibrary, err.Error())
	}
	defer os.Remove(signFile.Name())
	signFile.Close()

	signatureWriter, err := os.Create(signFile.Name())
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorWharfLibrary, err.Error())
	}
	defer signatureWriter.Close()

	rawSigWire := wire.NewWriteContext(signatureWriter)
	rawSigWire.WriteMagic(pwr.SignatureMagic)
	rawSigWire.WriteMessage(&pwr.SignatureHeader{Compression: compressionSettings()})

	sigWire, err := pwr.CompressWire(rawSigWire, compressionSettings())
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorWharfLibrary, err.Error())
	}

	sigWire.WriteMessage(container)

	err = pwr.ComputeSignatureToWriter(context2.Background(), container, pool, nil, func(hash wsync.BlockHash) error {
		return sigWire.WriteMessage(&pwr.BlockHash{
			WeakHash:   hash.WeakHash,
			StrongHash: hash.StrongHash,
		})
	})

	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorWharfLibrary, err.Error())
	}

	err = sigWire.Close()
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorWharfLibrary, err.Error())
	}

	signData, err := ioutil.ReadFile(signFile.Name())
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorFileIOFailure, err.Error())
	}

	return context.JSON(http.StatusOK, models.SignatureCmdResult{FileData: signData})
}

func ApplyPatchCmd(context echo.Context) error {

	reqCmp := &models.ApplyPatchCmd{}
	err := context.Bind(reqCmp)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorInvalidJSONFormat, err.Error())
	}

	fpath, err := utils.GetUserBuildPath(context.Request().Header.Get("ClientID"), reqCmp.GameID, reqCmp.BranchName)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorGetUserStorage, err.Error())
	}

	patchFile, err := ioutil.TempFile(os.TempDir(), "patch")
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorFileIOFailure, err.Error())
	}
	defer os.Remove(patchFile.Name())
	patchFile.Close()

	err = ioutil.WriteFile(patchFile.Name(), reqCmp.FileData, 0777)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorFileIOFailure, err.Error())
	}
	patchFile.Close()

	patchReader, err := eos.Open(patchFile.Name())
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorFileIOFailure, err.Error())
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
		return utils.BuildInternalServerError(context, models.ErrorWharfLibrary, err.Error())
	}

	err = actx.ApplyPatch(patchSource)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorWharfLibrary, err.Error())
	}

	return context.NoContent(http.StatusOK)
}
