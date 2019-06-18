package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"context"

	"github.com/itchio/savior/seeksource"
	"github.com/itchio/wharf/eos"
	"github.com/itchio/wharf/eos/option"
	"github.com/itchio/wharf/pools"
	"github.com/itchio/wharf/pwr"
	"github.com/itchio/wharf/state"
	"github.com/itchio/wharf/tlc"
	"github.com/itchio/wharf/wsync"
	"github.com/itchio/wharf/wire"
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

func CreateSignature(source string, consumer *state.Consumer) ([]byte, error) {

	container, err := tlc.WalkAny(source, &tlc.WalkOpts{Filter: filterPaths})
	if err != nil {
		return nil, err
	}

	pool, err := pools.New(container, source)
	if err != nil {
		return nil, err
	}

	signFile, err := ioutil.TempFile(os.TempDir(), "sign")
	if err != nil {
		return nil, err
	}
	defer os.Remove(signFile.Name())
	signFile.Close()

	signatureWriter, err := os.Create(signFile.Name())
	if err != nil {
		return nil, err
	}
	defer signatureWriter.Close()

	rawSigWire := wire.NewWriteContext(signatureWriter)
	rawSigWire.WriteMagic(pwr.SignatureMagic)
	rawSigWire.WriteMessage(&pwr.SignatureHeader{Compression: compressionSettings()})

	sigWire, err := pwr.CompressWire(rawSigWire, compressionSettings())
	if err != nil {
		return nil, err
	}

	sigWire.WriteMessage(container)

	err = pwr.ComputeSignatureToWriter(context.Background(), container, pool, consumer, func(hash wsync.BlockHash) error {
		return sigWire.WriteMessage(&pwr.BlockHash{
			WeakHash:   hash.WeakHash,
			StrongHash: hash.StrongHash,
		})
	})

	if err != nil {
		return nil, err
	}

	err = sigWire.Close()
	if err != nil {
		return nil, err
	}

	signData, err := ioutil.ReadFile(signFile.Name())
	if err != nil {
		return nil, err
	}

	return signData, nil
}

func GetSignatureInfo(singData []byte, consumer *state.Consumer) (*pwr.SignatureInfo, error) {

	singFile, err := ioutil.TempFile(os.TempDir(), "sign")
	if err != nil {
		return nil, err
	}
	defer os.Remove(singFile.Name())
	singFile.Close()

	err = ioutil.WriteFile(singFile.Name(), singData, 0777)
	if err != nil {
		return nil, err
	}
	singFile.Close()

	sigReader, err := eos.Open(singFile.Name(), option.WithConsumer(consumer))
	if err != nil {
		return nil, err
	}
	defer sigReader.Close()

	sigSource := seeksource.FromFile(sigReader)
	_, err = sigSource.Resume(nil)
	if err != nil {
		return nil, err
	}

	signatureInfo, err := pwr.ReadSignature(context.Background(), sigSource)
	if err != nil {
		return nil, err
	}

	return signatureInfo, nil
}

func CreatePatch(source string, signatureInfo *pwr.SignatureInfo, consumer *state.Consumer) ([]byte, error) {

	sourceContainer, err := tlc.WalkAny(source, &tlc.WalkOpts{Filter: filterPaths})
	if err != nil {
		return nil, err
	}

	var sourcePool wsync.Pool
	sourcePool, err = pools.New(sourceContainer, source)
	if err != nil {
		return nil, err
	}

	patchFile, err := ioutil.TempFile(os.TempDir(), "patch")
	if err != nil {
		return nil, err
	}
	defer os.Remove(patchFile.Name())
	patchFile.Close()

	patchWriter, err := os.Create(patchFile.Name())
	if err != nil {
		return nil, err
	}
	defer patchWriter.Close()

	signFile, err := ioutil.TempFile(os.TempDir(), "sign")
	if err != nil {
		return nil, err
	}
	defer os.Remove(signFile.Name())
	signFile.Close()

	signatureWriter, err := os.Create(signFile.Name())
	if err != nil {
		return nil, err
	}
	defer signatureWriter.Close()

	dctx := &pwr.DiffContext{
		SourceContainer: sourceContainer,
		Pool:            sourcePool,

		TargetContainer: signatureInfo.Container,
		TargetSignature: signatureInfo.Hashes,

		Consumer:    consumer,
		Compression: compressionSettings(),
	}

	err = dctx.WritePatch(context.Background(), patchWriter, signatureWriter)
	if err != nil {
		return nil, err
	}

	patchData, err := ioutil.ReadFile(patchFile.Name())
	if err != nil {
		return nil, err
	}

	return patchData, nil
}

func ApplyPatch(target string, patchData []byte, consumer *state.Consumer) error {

	patchFile, err := ioutil.TempFile(os.TempDir(), "patch")
	if err != nil {
		return err
	}
	defer os.Remove(patchFile.Name())
	patchFile.Close()

	err = ioutil.WriteFile(patchFile.Name(), patchData, 0777)
	if err != nil {
		return err
	}
	patchFile.Close()

	patchReader, err := eos.Open(patchFile.Name())
	if err != nil {
		return err
	}

	actx := &pwr.ApplyContext{
		TargetPath: target,
		OutputPath: target,
		DryRun:     false,
		InPlace:    false,
		Signature:  nil,
		WoundsPath: "",
		StagePath:  "",
		HealPath:   "",
		Consumer: consumer,
	}

	patchSource := seeksource.FromFile(patchReader)

	_, err = patchSource.Resume(nil)
	if err != nil {
		return err
	}

	err = actx.ApplyPatch(patchSource)
	if err != nil {
		return err
	}

	return nil
}