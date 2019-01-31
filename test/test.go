package test

import (
	"fmt"
	"context"
	"os"
	"errors"
	//"time"
	//"github.com/aristanetworks/goarista/monotime"

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

func Test() {
	
	output := "D:\\Temp\\test.old"
	signature := "D:\\Temp\\test.sign" 

	//monotime.Now()
	fmt.Println("Test")

	container, err := tlc.WalkAny(output, &tlc.WalkOpts{Filter: filtering.FilterPaths})
	if err != nil {
		panic(fmt.Srintf("walking directory to sign: %s", err.Error()))
	}

	pool, err := pools.New(container, output)
	if err != nil {
		panic(fmt.Srintf("creating pool for directory to sign: %s", err.Error()))
	}

	signatureWriter, err := os.Create(signature)
	if err != nil {
		panic(fmt.Srintf("creating signature file: %s", err.Error()))
	}
	defer signatureWriter.Close()

	rawSigWire := wire.NewWriteContext(signatureWriter)
	rawSigWire.WriteMagic(pwr.SignatureMagic)

	rawSigWire.WriteMessage(&pwr.SignatureHeader{
		Compression: &compression,
	})

	sigWire, err := pwr.CompressWire(rawSigWire, &compression)
	if err != nil {
		panic(fmt.Srintf("setting up compression for signature file: %s", err.Error()))
	}
	sigWire.WriteMessage(container)

	err = pwr.ComputeSignatureToWriter(context.Background(), container, pool, newStateConsumer(), func(hash wsync.BlockHash) error {
		return sigWire.WriteMessage(&pwr.BlockHash{
			WeakHash:   hash.WeakHash,
			StrongHash: hash.StrongHash,
		})
	})
	
	if err != nil {
		panic(fmt.Srintf("computing signature: %s", err.Error()))
	}

	err = sigWire.Close()
	if err != nil {
		panic(fmt.Srintf("finalizing signature file: %s", err.Error()))
	}
}
