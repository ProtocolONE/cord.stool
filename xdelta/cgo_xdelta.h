#pragma once

#ifdef __cplusplus
extern "C" {
#endif

typedef enum {

  CGO_XD3_FROM_FILE_OPEN_FAILED = -1,
  CGO_XD3_TO_FILE_OPEN_FAILED = -2,
  CGO_XD3_TO_FILE_CREATE_FAILED = -3,
  CGO_XD3_DIFF_FILE_CREATE_FAILED = -4,
  CGO_XD3_DIFF_FILE_OPEN_FAILED = -5,

  /* An application must be prepared to handle these five return
   * values from either xd3_encode_input or xd3_decode_input, except
   * in the case of no-source compression, in which case XD3_GETSRCBLK
   * is never returned.  More detailed comments for these are given in
   * xd3_encode_input and xd3_decode_input comments, below. */
  CGO_XD3_INPUT = -17703, /* need input */
  CGO_XD3_OUTPUT = -17704, /* have output */
  CGO_XD3_GETSRCBLK = -17705, /* need a block of source input (with no
         * xd3_getblk function), a chance to do
         * non-blocking read. */
  CGO_XD3_GOTHEADER = -17706, /* (decode-only) after the initial VCDIFF &
           first window header */
  CGO_XD3_WINSTART = -17707, /* notification: returned before a window is
         * processed, giving a chance to
         * XD3_SKIP_WINDOW or not XD3_SKIP_EMIT that
         * window. */
  CGO_XD3_WINFINISH = -17708, /* notification: returned after
            encode/decode & output for a window */
  CGO_XD3_TOOFARBACK = -17709, /* (encoder only) may be returned by
            getblk() if the block is too old */
  CGO_XD3_INTERNAL = -17710, /* internal error */
  CGO_XD3_INVALID = -17711, /* invalid config */
  CGO_XD3_INVALID_INPUT = -17712, /* invalid input/decoder error */
  CGO_XD3_NOSECOND = -17713, /* when secondary compression finds no
             improvement. */
  CGO_XD3_UNIMPLEMENTED = -17714  /* currently VCD_TARGET, VCD_CODETABLE */
} cgo_xd3_rvalues;

int encodeDiff(unsigned int from, unsigned int to, unsigned int diff);
int decodeDiff(unsigned int from, unsigned int to, unsigned int diff);

#ifdef __cplusplus
}
#endif