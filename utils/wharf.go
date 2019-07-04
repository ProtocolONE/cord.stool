package utils

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/itchio/headway/state"
	"github.com/itchio/httpkit/eos"
	"github.com/itchio/httpkit/eos/option"
	"github.com/itchio/lake"
	"github.com/itchio/lake/pools"
	"github.com/itchio/lake/tlc"
	"github.com/itchio/savior/seeksource"
	"github.com/itchio/wharf/pwr"
	"github.com/itchio/wharf/wire"
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

func CreateSignatureFile(source string, signatureFile string, consumer *state.Consumer) error {

	container, err := tlc.WalkAny(source, &tlc.WalkOpts{Filter: filterPaths})
	if err != nil {
		return err
	}

	pool, err := pools.New(container, source)
	if err != nil {
		return err
	}

	signatureWriter, err := os.Create(signatureFile)
	if err != nil {
		return err
	}
	defer signatureWriter.Close()

	rawSigWire := wire.NewWriteContext(signatureWriter)
	rawSigWire.WriteMagic(pwr.SignatureMagic)
	rawSigWire.WriteMessage(&pwr.SignatureHeader{Compression: compressionSettings()})

	sigWire, err := pwr.CompressWire(rawSigWire, compressionSettings())
	if err != nil {
		return err
	}

	sigWire.WriteMessage(container)

	err = pwr.ComputeSignatureToWriter(context.Background(), container, pool, consumer, func(hash wsync.BlockHash) error {
		return sigWire.WriteMessage(&pwr.BlockHash{
			WeakHash:   hash.WeakHash,
			StrongHash: hash.StrongHash,
		})
	})

	if err != nil {
		return err
	}

	err = sigWire.Close()
	if err != nil {
		return err
	}

	return nil
}

func CreateSignatureData(source string, consumer *state.Consumer) ([]byte, error) {

	signFile, err := ioutil.TempFile(os.TempDir(), "sign")
	if err != nil {
		return nil, err
	}
	defer os.Remove(signFile.Name())
	signFile.Close()

	err = CreateSignatureFile(source, signFile.Name(), consumer)
	if err != nil {
		return nil, err
	}

	signData, err := ioutil.ReadFile(signFile.Name())
	if err != nil {
		return nil, err
	}

	return signData, nil
}

func GetSignatureInfoFromFile(signatureFile string, consumer *state.Consumer) (*pwr.SignatureInfo, error) {

	sigReader, err := eos.Open(signatureFile, option.WithConsumer(consumer))
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

func GetSignatureInfoFromData(singData []byte, consumer *state.Consumer) (*pwr.SignatureInfo, error) {

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

	signatureInfo, err := GetSignatureInfoFromFile(singFile.Name(), consumer)
	if err != nil {
		return nil, err
	}

	return signatureInfo, nil
}

func CreatePatchFile(source string, patchFile string, signatureInfo *pwr.SignatureInfo, consumer *state.Consumer) error {

	sourceContainer, err := tlc.WalkAny(source, &tlc.WalkOpts{Filter: filterPaths})
	if err != nil {
		return err
	}

	var sourcePool lake.Pool
	sourcePool, err = pools.New(sourceContainer, source)
	if err != nil {
		return err
	}

	patchWriter, err := os.Create(patchFile)
	if err != nil {
		return err
	}
	defer patchWriter.Close()

	signFile, err := ioutil.TempFile(os.TempDir(), "sign")
	if err != nil {
		return err
	}
	defer os.Remove(signFile.Name())
	signFile.Close()

	signatureWriter, err := os.Create(signFile.Name())
	if err != nil {
		return err
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
		return err
	}

	return nil
}

func CreatePatchData(source string, signatureInfo *pwr.SignatureInfo, consumer *state.Consumer) ([]byte, error) {

	patchFile, err := ioutil.TempFile(os.TempDir(), "patch")
	if err != nil {
		return nil, err
	}
	defer os.Remove(patchFile.Name())
	patchFile.Close()

	err = CreatePatchFile(source, patchFile.Name(), signatureInfo, consumer)
	if err != nil {
		return nil, err
	}

	patchData, err := ioutil.ReadFile(patchFile.Name())
	if err != nil {
		return nil, err
	}

	return patchData, nil
}

func ApplyPatchFile(target string, output string, patchFile string, consumer *state.Consumer) error {

	patchReader, err := eos.Open(patchFile)
	if err != nil {
		return err
	}

	actx := &pwr.ApplyContext{
		TargetPath: target,
		OutputPath: output,
		DryRun:     false,
		InPlace:    target == output,
		Signature:  nil,
		WoundsPath: "",
		StagePath:  "",
		HealPath:   "",
		Consumer:   consumer,
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

func ApplyPatchData(target string, output string, patchData []byte, consumer *state.Consumer) error {

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

	err = ApplyPatchFile(target, output, patchFile.Name(), consumer)
	if err != nil {
		return err
	}

	return nil
}
