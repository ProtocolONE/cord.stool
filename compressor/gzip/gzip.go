package gzip

import (
	"compress/gzip"
	"io"

	"github.com/itchio/savior"
	"github.com/itchio/savior/gzipsource"
	"github.com/itchio/wharf/pwr"
)

type gzipCompressor struct{}

func (gc *gzipCompressor) Apply(writer io.Writer, quality int32) (io.Writer, error) {
	return gzip.NewWriterLevel(writer, int(quality))
}

type gzipDecompressor struct{}

func (bc *gzipDecompressor) Apply(source savior.Source) (savior.Source, error) {
	return gzipsource.New(source), nil
}

func Init() {
	pwr.RegisterCompressor(pwr.CompressionAlgorithm_GZIP, &gzipCompressor{})
	pwr.RegisterDecompressor(pwr.CompressionAlgorithm_GZIP, &gzipDecompressor{})
}
