package update

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/anacrolix/envpprof"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/mmap_span"
	"github.com/bradfitz/iter"
	humanize "github.com/dustin/go-humanize"
	"github.com/edsrzf/mmap-go"
	"github.com/gosuri/uiprogress"
)

type ResumeData struct {
	Filename string
	Size     int64
	ModTime  time.Time
}

func exitSignalHandlers(client *torrent.Client) {

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	for {
		client.Close()
	}
}

func torrentBar(t *torrent.Torrent, bar *uiprogress.Bar, stats *DownloadStatistics) {

	startTime := time.Now()

	bar.Set(0)

	//completed := false

	bar.AppendFunc(func(*uiprogress.Bar) (ret string) {
		select {
		case <-t.GotInfo():
		default:
			return "getting info"
		}
		if t.Seeding() {
			return "seeding"
		} else if t.BytesCompleted() == t.Info().TotalLength() {
			//completed = true
			return "completed"
		} else {

			duration := uint64(time.Since(startTime)) / 1000000000
			if duration > 0 {
				stats.Update(uint64(t.BytesCompleted())/duration, 0, uint64(t.Info().TotalLength()))
			}
			return fmt.Sprintf("downloading (%s/%s)", humanize.Bytes(uint64(t.BytesCompleted())), humanize.Bytes(uint64(t.Info().TotalLength())))
		}
	})

	go func() {
		<-t.GotInfo()
		tl := int(t.Info().TotalLength())
		if tl == 0 {
			bar.Set(1)
			return
		}
		bar.Total = tl
		for {
			bc := t.BytesCompleted()
			bar.Set(int(bc))
			time.Sleep(time.Second)

			/*if completed {
				break
			}*/
		}
	}()

	//bar.Set(bar.Total)
}

func addTorrents(client *torrent.Client, torrentData []byte, bar *uiprogress.Bar, stats *DownloadStatistics) error {

	reader := bytes.NewReader(torrentData)
	mi, err := metainfo.Load(reader)
	if err != nil {
		return err
	}

	t, err := client.AddTorrent(mi)
	if err != nil {
		return err
	}

	torrentBar(t, bar, stats)

	stats.Start()

	go func() {
		<-t.GotInfo()
		t.DownloadAll()
	}()

	return nil
}

func StartDownloadFile(torrentFile string, output string, bar *uiprogress.Bar, stats *DownloadStatistics) error {

	torrentData, err := ioutil.ReadFile(torrentFile)
	if err != nil {
		return err
	}

	return StartDownload(torrentData, output, bar, stats)
}

func initSetting(clientConfig *torrent.ClientConfig) {

	clientConfig.Debug = false
	clientConfig.NoDHT = true
	clientConfig.DisableIPv6 = true
	clientConfig.NoDefaultPortForwarding = true

	clientConfig.HandshakesTimeout, _ = time.ParseDuration("5s")
	clientConfig.MinDialTimeout, _ = time.ParseDuration("10s")
}

func StartDownload(torrentData []byte, output string, bar *uiprogress.Bar, stats *DownloadStatistics) error {

	os.Remove(filepath.Join(output, ".torrent.bolt.db"))
	os.Remove(filepath.Join(output, ".torrent.bolt.db.lock"))
	defer os.Remove(filepath.Join(output, ".torrent.bolt.db"))
	defer os.Remove(filepath.Join(output, ".torrent.bolt.db.lock"))

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	defer w.Close()
	f := func() { os.Stdout = old }
	defer f()

	err := startDownload(torrentData, output, bar, stats)
	if err != nil {
		return err
	}

	saveFastResumeData(torrentData, output)

	return nil
}

func startDownload(torrentData []byte, output string, bar *uiprogress.Bar, stats *DownloadStatistics) error {

	defer envpprof.Stop()

	clientConfig := torrent.NewDefaultClientConfig()
	initSetting(clientConfig)
	clientConfig.DataDir = output

	client, err := torrent.NewClient(clientConfig)
	if err != nil {
		return err
	}

	defer client.Close()
	//go exitSignalHandlers(client)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		client.WriteStatus(w)
	})

	err = addTorrents(client, torrentData, bar, stats)
	if err != nil {
		return err
	}

	defer stats.Stop()

	if !client.WaitAll() {
		return fmt.Errorf("Download failed")
	}

	return nil
}

func saveFastResumeData(torrentData []byte, source string) error {

	reader := bytes.NewReader(torrentData)
	metaInfo, err := metainfo.Load(reader)
	if err != nil {
		return err
	}

	info, err := metaInfo.UnmarshalInfo()
	if err != nil {
		return err
	}

	var resumeData []ResumeData

	for _, file := range info.UpvertedFiles() {

		filename := filepath.Join(append([]string{source, info.Name}, file.Path...)...)
		size, modTime, err := getFileSizeAndModifyTime(filename)
		if err != nil {
			return err
		}

		filename = filepath.Join(append([]string{info.Name}, file.Path...)...)
		resumeData = append(resumeData, ResumeData{filename, size, modTime})
	}

	var binData bytes.Buffer

	enc := gob.NewEncoder(&binData)
	err = enc.Encode(resumeData)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(source, "resumedata.bin"), binData.Bytes(), 0777)
	if err != nil {
		return err
	}

	return nil
}

func mmapFile(name string) (mm mmap.MMap, err error) {
	f, err := os.Open(name)
	if err != nil {
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return
	}
	if fi.Size() == 0 {
		return
	}
	return mmap.MapRegion(f, -1, mmap.RDONLY, mmap.COPY, 0)
}

func verifyTorrent(info *metainfo.Info, source string, bar *uiprogress.Bar) error {

	bar.Set(0)

	span := new(mmap_span.MMapSpan)
	for _, file := range info.UpvertedFiles() {
		filename := filepath.Join(append([]string{source, info.Name}, file.Path...)...)
		mm, err := mmapFile(filename)
		if err != nil {
			return err
		}
		if int64(len(mm)) != file.Length {
			return fmt.Errorf("file %q has wrong length", filename)
		}
		span.Append(mm)
	}

	bar.Total = info.NumPieces()

	for i := range iter.N(info.NumPieces()) {
		p := info.Piece(i)
		hash := sha1.New()
		_, err := io.Copy(hash, io.NewSectionReader(span, p.Offset(), p.Length()))
		if err != nil {
			return err
		}
		good := bytes.Equal(hash.Sum(nil), p.Hash().Bytes())
		if !good {
			return fmt.Errorf("hash mismatch at piece %d", i)
		}
		bar.Incr()
	}
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

		return err //VerifyTorrent(torrentData, source, bar)
	}

	if !test {

		return fmt.Errorf("Checking failed")
	}

	return nil
}

func VerifyTorrent(torrentData []byte, source string, bar *uiprogress.Bar) error {

	reader := bytes.NewReader(torrentData)
	metaInfo, err := metainfo.Load(reader)
	if err != nil {
		return err
	}

	info, err := metaInfo.UnmarshalInfo()
	if err != nil {
		return err
	}

	err = verifyTorrent(&info, source, bar)
	if err != nil {
		return err
	}

	return nil
}
