package xdelta

/*
#include <stdlib.h> // for C.free
#include "cgo_xdelta.h"
*/
import "C"
import (
	"errors"
	"os"
	"strconv"
)

// EncodeDiff ...
func EncodeDiff(from string, to string, diff string) error {

	fileFrom, err := os.Open(from)
	if err != nil {
		return errors.New("Cannot open source_old file: " + err.Error())
	}
	defer fileFrom.Close()

	fileTo, err := os.Open(to)
	if err != nil {
		return errors.New("Cannot open source_new file: " + err.Error())
	}
	defer fileTo.Close()

	filediff, err := os.Create(diff)
	if err != nil {
		return errors.New("Cannot create diff file: " + err.Error())
	}
	defer filediff.Close()

	ret := C.encodeDiff(C.uint(fileFrom.Fd()), C.uint(fileTo.Fd()), C.uint(filediff.Fd()))
	if ret != 0 {
		return errors.New("xdelta error: " + strconv.Itoa(int(ret)))
	}

	return nil
}

// DecodeDiff ...
func DecodeDiff(from string, to string, diff string) error {

	fileFrom, err := os.Open(from)
	if err != nil {
		return errors.New("Cannot open source_old file: " + err.Error())
	}
	defer fileFrom.Close()

	fileTo, err := os.Create(to)
	if err != nil {
		return errors.New("Cannot create source_new file: " + err.Error())
	}
	defer fileTo.Close()

	filediff, err := os.Open(diff)
	if err != nil {
		return errors.New("Cannot open diff file: " + err.Error())
	}
	defer filediff.Close()

	ret := C.decodeDiff(C.uint(fileFrom.Fd()), C.uint(fileTo.Fd()), C.uint(filediff.Fd()))
	if ret != 0 {
		return errors.New("xdelta error: " + strconv.Itoa(int(ret)))
	}

	return nil
}
