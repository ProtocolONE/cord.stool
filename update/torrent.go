package update

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"cord.stool/libtorrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/gosuri/uiprogress"
)

type ResumeData struct {
	Filename string
	Size     int64
	ModTime  time.Time
}

func StartDownloadFile(torrentFile string, output string, bar *uiprogress.Bar, stats *DownloadStatistics) error {

	torrentData, err := ioutil.ReadFile(torrentFile)
	if err != nil {
		return err
	}

	return StartDownload(torrentData, output, bar, stats)
}

func StartDownload(torrentData []byte, output string, bar *uiprogress.Bar, stats *DownloadStatistics) error {

	bar.Set(0)

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	defer w.Close()
	f := func() { os.Stdout = old }
	defer f()

	return startDownload(torrentData, output, bar, stats)
}

func startDownload(torrentData []byte, output string, bar *uiprogress.Bar, stats *DownloadStatistics) error {

	session, err := libtorrent.CreateSession()
	if err != nil {
		return err
	}

	defer session.CloseSession()

	torrent, err := session.AddTorrentData(torrentData, output)
	if err != nil {
		return err
	}

	bar.Total = 100

	for {

		status := torrent.GetTorrentStatus()
		if status == nil {
			break
		}

		bar.Set(int(status.Progress * 100))

		if int(status.Progress * 100) == 100 {
			break
		}

		if status.ErrorText != "" {
			return fmt.Errorf("ERROR: %s", status.ErrorText)
		}

		time.Sleep(1 * time.Second)
	}

	bar.Set(bar.Total)

	return nil
}

func getFileSizeAndModifyTime(filename string) (int64, time.Time, error) {

	fi, err := os.Stat(filename)
	if err != nil {
		return -1, time.Now(), err
	}

	return fi.Size(), fi.ModTime(), nil
}

func verifyTorrentLight(info *metainfo.Info, source string, bar *uiprogress.Bar) (bool, error) {

	bar.Set(0)
	var resumeData []ResumeData

	binData, err := ioutil.ReadFile(filepath.Join(source, "resumedata.bin"))
	if err != nil {
		return false, err
	}

	reader := bytes.NewReader(binData)

	dec := gob.NewDecoder(reader)
	err = dec.Decode(&resumeData)
	if err != nil {
		return false, err
	}

	bar.Total = len(resumeData)

	for _, resume := range resumeData {

		filename := filepath.Join(source, resume.Filename)
		size, time, err := getFileSizeAndModifyTime(filename)
		if err != nil {
			return false, err
		}

		if resume.Size != size || resume.ModTime != time {
			return false, nil
		}

		bar.Incr()
	}

	return true, nil
}

func VerifyTorrentFile(torrentFile string, source string, bar *uiprogress.Bar) error {

	torrentData, err := ioutil.ReadFile(torrentFile)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(torrentData)
	metaInfo, err := metainfo.Load(reader)
	if err != nil {
		return err
	}

	info, err := metaInfo.UnmarshalInfo()
	if err != nil {
		return err
	}

	test, err := verifyTorrentLight(&info, source, bar)
	if err != nil {

		return err
	}

	if !test {

		return fmt.Errorf("Checking failed")
	}

	return nil
}
