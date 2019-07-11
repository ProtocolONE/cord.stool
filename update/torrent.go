package update

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/ProtocolONE/rain/torrent"
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

	os.Remove(filepath.Join(output, "session.db"))
	os.Remove(filepath.Join(output, "session.lock"))
	defer os.Remove(filepath.Join(output, "session.db"))
	defer os.Remove(filepath.Join(output, "session.lock"))

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	defer w.Close()
	f := func() { os.Stdout = old }
	defer f()

	return startDownload(torrentData, output, bar, stats)
}

func startDownload(torrentData []byte, output string, bar *uiprogress.Bar, stats *DownloadStatistics) error {

	s, closeSession, err := newSession(output)
	if err != nil {
		return err
	}

	defer closeSession()

	id := filepath.Base(output)
	s.RemoveTorrent(id)

	r := bytes.NewReader(torrentData)
	tor, err := s.AddTorrent(r)
	if err != nil {
		return err
	}

	tor.Start()
	defer tor.Stop()

	stats.Start()
	defer stats.Stop()

	bar.Total = 100

	for {

		var p float64
		p = float64(tor.Stats().Bytes.Completed * 100)
		p = p / float64(tor.Stats().Bytes.Total)
		bar.Set(int(p))

		time.Sleep(time.Second)

		torStats := tor.Stats()
		if torStats.Error != nil {
			return torStats.Error
		} else if torStats.Status == torrent.Seeding {
			break
		}

		stats.Update(uint64(torStats.Speed.Download), uint64(torStats.Speed.Upload), uint64(torStats.Bytes.Total))
	}

	bar.Set(bar.Total)

	return nil
}

func newSession(output string) (*torrent.Session, func(), error) {

	cfg := torrent.DefaultConfig
	cfg.DataDir = output
	cfg.Database = filepath.Join(output, "session.db")

	s, err := torrent.NewSession(cfg)
	if err != nil {
		return nil, nil, err
	}

	return s, func() {
		err := s.Close()
		if err != nil {
			return
		}
	}, nil
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
