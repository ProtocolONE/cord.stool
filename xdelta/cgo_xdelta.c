#include <stdio.h>
#include <sys/stat.h>
#include <io.h>
#include <Fcntl.h>

#include "xdelta3.h"
#include "cgo_xdelta.h"

int code(
  int encode,
  FILE*  InFile,
  FILE*  SrcFile,
  FILE* OutFile,
  int BufSize)
{
  int r, ret;
  struct stat statbuf;
  xd3_stream stream;
  xd3_config config;
  xd3_source source;
  void* Input_Buf;
  int Input_Buf_Read;

  if (BufSize < XD3_ALLOCSIZE)
    BufSize = XD3_ALLOCSIZE;

  memset(&stream, 0, sizeof(stream));
  memset(&source, 0, sizeof(source));

  xd3_init_config(&config, XD3_ADLER32);
  config.winsize = BufSize;
  xd3_config_stream(&stream, &config);

  if (SrcFile)
  {
    r = fstat(_fileno(SrcFile), &statbuf);
    if (r)
      return r;

    source.blksize = BufSize;
    source.curblk = malloc(source.blksize);

    /* Load 1st block of stream. */
    r = fseek(SrcFile, 0, SEEK_SET);
    if (r)
      return r;
    source.onblk = fread((void*)source.curblk, 1, source.blksize, SrcFile);
    source.curblkno = 0;
    /* Set the stream. */
    xd3_set_source(&stream, &source);
  }

  Input_Buf = malloc(BufSize);

  fseek(InFile, 0, SEEK_SET);
  do
  {
    Input_Buf_Read = fread(Input_Buf, 1, BufSize, InFile);
    if (Input_Buf_Read < BufSize)
    {
      xd3_set_flags(&stream, XD3_FLUSH | stream.flags);
    }
    xd3_avail_input(&stream, Input_Buf, Input_Buf_Read);

  process:
    if (encode)
      ret = xd3_encode_input(&stream);
    else
      ret = xd3_decode_input(&stream);

    switch (ret)
    {
    case XD3_INPUT:
    {
      continue;
    }

    case XD3_OUTPUT:
    {
      r = fwrite(stream.next_out, 1, stream.avail_out, OutFile);
      if (r != (int)stream.avail_out)
        return r;
      xd3_consume_output(&stream);
      goto process;
    }

    case XD3_GETSRCBLK:
    {
      if (SrcFile)
      {
        r = fseek(SrcFile, source.blksize * source.getblkno, SEEK_SET);
        if (r)
          return r;
        source.onblk = fread((void*)source.curblk, 1,
          source.blksize, SrcFile);
        source.curblkno = source.getblkno;
      }
      goto process;
    }

    case XD3_GOTHEADER:
    {
      goto process;
    }

    case XD3_WINSTART:
    {
      goto process;
    }

    case XD3_WINFINISH:
    {
      goto process;
    }

    default:
    {
      return ret;
    }

    }

  } while (Input_Buf_Read == BufSize);

  free(Input_Buf);

  free((void*)source.curblk);
  xd3_close_stream(&stream);
  xd3_free_stream(&stream);

  return 0;

};

int encodeDiff(unsigned int from, unsigned int to, unsigned int diff)
{
  int result = 0;

  int fdfrom = _open_osfhandle(from, _O_RDONLY|_O_BINARY);
  if (-1 == fdfrom) {
    result = CGO_XD3_FROM_FILE_OPEN_FAILED;
    goto cleanup;
  }

  FILE* fromFile = fdopen(fdfrom, "rb");
  if (!fromFile) {
    result = CGO_XD3_FROM_FILE_OPEN_FAILED;
    goto cleanup;
  }

  int fdTo = _open_osfhandle(to, _O_RDONLY|_O_BINARY);
  if (-1 == fdTo) {
    result = CGO_XD3_TO_FILE_OPEN_FAILED;
    goto cleanup;
  }

  FILE* toFile = fdopen(fdTo, "rb");
  if (!toFile) {
    result = CGO_XD3_TO_FILE_OPEN_FAILED;
    goto cleanup;
  }

  int fdDiff = _open_osfhandle(diff, _O_CREAT|_O_WRONLY|_O_BINARY);
  if (-1 == fdDiff) {
    result = CGO_XD3_DIFF_FILE_CREATE_FAILED;
    goto cleanup;
  }

  FILE* diffFile = fdopen(fdDiff, "wb");
  if (!diffFile) {
    result = CGO_XD3_DIFF_FILE_CREATE_FAILED;
    goto cleanup;
  }

  result = code(1, toFile, fromFile, diffFile, 0x1000);
  if (result)
    goto cleanup;

  fflush(diffFile);

cleanup:
  return result;
}

int decodeDiff(unsigned int from, unsigned int to, unsigned int diff)
{
  int result = 0;

  int fdfrom = _open_osfhandle(from, _O_RDONLY|_O_BINARY);
  if (-1 == fdfrom) {
    result = CGO_XD3_FROM_FILE_OPEN_FAILED;
    goto cleanup;
  }

  FILE* fromFile = fdopen(fdfrom, "rb");
  if (!fromFile) {
    result = CGO_XD3_FROM_FILE_OPEN_FAILED;
    return result;
  }

  int fdTo = _open_osfhandle(to, _O_CREAT|_O_WRONLY|_O_BINARY);
  if (-1 == fdTo) {
    result = CGO_XD3_TO_FILE_CREATE_FAILED;
    goto cleanup;
  }

  FILE* toFile = fdopen(fdTo, "wb");
  if (!toFile) {
    result = CGO_XD3_TO_FILE_CREATE_FAILED;
    return result;
  }

  int fdDiff = _open_osfhandle(diff, _O_RDONLY|_O_BINARY);
  if (-1 == fdDiff) {
    result = CGO_XD3_DIFF_FILE_OPEN_FAILED;
    goto cleanup;
  }

  FILE* diffFile = fdopen(fdDiff, "rb");
  if (!diffFile) {
    result = CGO_XD3_DIFF_FILE_OPEN_FAILED;
    return result;
  }

  result = code(0, diffFile, fromFile, toFile, 0x1000);
  if (result)
    goto cleanup;

  fflush(toFile);

cleanup:
  return result;
}
