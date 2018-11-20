package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func tst(a ...int) {
	if len(a) == 0 {
		return
	}

	println(a[0], len(a))
	tst(a[1:]...)
}

func Test_UpdateInfoPack(t *testing.T) {
	expectedResult := `<?xml version="1.0" encoding="UTF-8"?>
<UpdateFileList version="3.5.1">
  <files>
    <file path="qwe" crc="qqqq" rawLength="123" archiveLength="13" check="false"></file>
    <file path="qwe1\\qwe" crc="qqqq" rawLength="123" archiveLength="13" check="true"></file>
  </files>
</UpdateFileList>`

	info := UpdateInfo{

		Version1: "3.5.1",
		FilesInfo: FileInfoList{
			Files: []FileInfo{
				FileInfo{
					Path:          "qwe",
					Hash:          "qqqq",
					RawLength:     123,
					ArchiveLength: 13,
					Check:         false,
				},

				FileInfo{
					Path:          `qwe1\\qwe`,
					Hash:          "qqqq",
					RawLength:     123,
					ArchiveLength: 13,
					Check:         true,
				},
			},
		},
	}

	s, e := info.Pack()

	assert.True(t, e == nil)
	assert.True(t, s == expectedResult)
}
