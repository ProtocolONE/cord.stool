//#define WINVER  0x0602

#include <stdio.h>
#include <sys/stat.h>
#include "xdelta3.h"
#include "cgo_xdelat.h"

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

int encodeDiff(const char* from, const char* to, const char* diff)
{
  FILE* fromFile = NULL;
  FILE* toFile = NULL;
  FILE* diffFile = NULL;
  errno_t err = 0;
  int result = 0;

  fromFile = fopen(from, "rb");
  if (!fromFile) {
    result = CGO_XD3_FROM_FILE_OPEN_FAILED;
    goto cleanup;
  }

  toFile = fopen(to, "rb");
  if (!toFile) {
    result = CGO_XD3_TO_FILE_OPEN_FAILED;
    goto cleanup;
  }

  diffFile = fopen(diff, "wb");
  if (!diffFile) {
    result = CGO_XD3_DIFF_FILE_CREATE_FAILED;
    goto cleanup;
  }

  result = code(1, toFile, fromFile, diffFile, 0x1000);
  if (result) {
    goto cleanup;
  }

cleanup:
  if (fromFile)
    fclose(fromFile);
  
  if (toFile)
    fclose(toFile);

  if (diffFile)
    fclose(diffFile);

  return result;
}

int decodeDiff(const char* from, const char* to, const char* diff)
{
  FILE* fromFile = NULL;
  FILE* toFile = NULL;
  FILE* diffFile = NULL;
  errno_t err = 0;
  int result = 0;

  fromFile = fopen(from, "rb");
  if (!fromFile) {
    result = CGO_XD3_FROM_FILE_OPEN_FAILED;
    goto cleanup;
  }

  toFile = fopen(to, "wb");
  if (!toFile) {
    result = CGO_XD3_TO_FILE_CREATE_FAILED;
    goto cleanup;
  }

  diffFile = fopen(diff, "rb");
  if (!diffFile) {
    result = CGO_XD3_DIFF_FILE_OPEN_FAILED;
    goto cleanup;
  }

  result = code(0, diffFile, fromFile, toFile, 0x1000);
  if (result) {
    goto cleanup;
  }

cleanup:
  if (fromFile)
    fclose(fromFile);

  if (toFile)
    fclose(toFile);

  if (diffFile)
    fclose(diffFile);

  return result;
}
