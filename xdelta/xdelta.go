package xdelta

/*
#include <stdlib.h> // for C.free
#include "cgo_xdelat.h"
*/
import "C"
import (
	"errors"
	"strconv"
	"unsafe"
)

// EncodeDiff ...
func EncodeDiff(from string, to string, diff string) error {

	cfrom := C.CString(from)
	defer C.free(unsafe.Pointer(cfrom))
	cto := C.CString(to)
	defer C.free(unsafe.Pointer(cto))
	cdiff := C.CString(diff)
	defer C.free(unsafe.Pointer(cdiff))

	ret := C.encodeDiff(cfrom, cto, cdiff)
	if ret != 0 {
		return errors.New("xdelta error: " + strconv.Itoa(int(ret)))
	}

	return nil
}

// DecodeDiff ...
func DecodeDiff(from string, to string, diff string) error {

	cfrom := C.CString(from)
	defer C.free(unsafe.Pointer(cfrom))
	cto := C.CString(to)
	defer C.free(unsafe.Pointer(cto))
	cdiff := C.CString(diff)
	defer C.free(unsafe.Pointer(cdiff))

	ret := C.decodeDiff(cfrom, cto, cdiff)
	if ret != 0 {
		return errors.New("xdelta error: " + strconv.Itoa(int(ret)))
	}

	return nil
}
