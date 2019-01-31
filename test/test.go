package test

import (
	"fmt"
	"context"
	"os"
	"path/filepath"

	"github.com/itchio/wharf/pools"
	"github.com/itchio/wharf/pwr"
	"github.com/itchio/wharf/tlc"
	"github.com/itchio/wharf/wire"
	"github.com/itchio/wharf/wsync"
	"github.com/itchio/wharf/state"
)

func newStateConsumer() *state.Consumer {
	return &state.Consumer{
		OnProgress:       Progress,
		OnProgressLabel:  ProgressLabel,
		OnPauseProgress:  PauseProgress,
		OnResumeProgress: ResumeProgress,
		OnMessage:        Logl,
	}
}

var IgnoredPaths = []string{
	".git",
	".hg",
	".svn",
	".DS_Store",
	"__MACOSX",
	"._*",
	"Thumbs.db",
	".itch",
}

// FilterPaths filters out known bad folder/files
// which butler should just ignore
func FilterPaths(fileInfo os.FileInfo) bool {
	name := fileInfo.Name()
	for _, pattern := range IgnoredPaths {
		match, _ := filepath.Match(pattern, name)
		if match {
			return false
		}
	}

	return true
}

func CompressionSettings() pwr.CompressionSettings {

	return pwr.CompressionSettings{
		Algorithm: pwr.CompressionAlgorithm_GZIP,
		Quality:   int32(5),
	}
}

func ProgressLabel(label string) {
}

func PauseProgress() {
}

func ResumeProgress() {
}

func Progress(alpha float64) {
}

func Logl(level string, msg string) {
}

// Test ...
func Test() {
	
	output := "D:\\Temp\\test.old"
	signature := "D:\\Temp\\test.sign" 

	//monotime.Now()
	fmt.Println("Test")

	container, err := tlc.WalkAny(output, &tlc.WalkOpts{Filter: FilterPaths})
	if err != nil {
		panic(fmt.Sprintf("walking directory to sign: %s", err.Error()))
	}

	pool, err := pools.New(container, output)
	if err != nil {
		panic(fmt.Sprintf("creating pool for directory to sign: %s", err.Error()))
	}

	signatureWriter, err := os.Create(signature)
	if err != nil {
		panic(fmt.Sprintf("creating signature file: %s", err.Error()))
	}
	defer signatureWriter.Close()

	rawSigWire := wire.NewWriteContext(signatureWriter)
	rawSigWire.WriteMagic(pwr.SignatureMagic)

	rawSigWire.WriteMessage(&pwr.SignatureHeader{Compression: &CompressionSettings())

	sigWire, err := pwr.CompressWire(rawSigWire, &CompressionSettings())

	if err != nil {
		panic(fmt.Sprintf("setting up compression for signature file: %s", err.Error()))
	}
	sigWire.WriteMessage(container)

	err = pwr.ComputeSignatureToWriter(context.Background(), container, pool, newStateConsumer(), func(hash wsync.BlockHash) error {
		return sigWire.WriteMessage(&pwr.BlockHash{
			WeakHash:   hash.WeakHash,
			StrongHash: hash.StrongHash,
		})
	})
	
	if err != nil {
		panic(fmt.Sprintf("computing signature: %s", err.Error()))
	}

	err = sigWire.Close()
	if err != nil {
		panic(fmt.Sprintf("finalizing signature file: %s", err.Error()))
	}
}
