package update

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	//"log"

	"github.com/anacrolix/envpprof"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	humanize "github.com/dustin/go-humanize"
	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
)

func exitSignalHandlers(client *torrent.Client) {

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	for {
		fmt.Printf("close signal received: %+v", <-c)
		client.Close()
	}
}

func torrentBar(t *torrent.Torrent, bar *uiprogress.Bar) {

	bar.PrependFunc(func(*uiprogress.Bar) string {
		return strutil.Resize("Downloading ...", 35)
	})

	bar.AppendFunc(func(*uiprogress.Bar) (ret string) {
		select {
		case <-t.GotInfo():
		default:
			return "getting info"
		}
		if t.Seeding() {
			return "seeding"
		} else if t.BytesCompleted() == t.Info().TotalLength() {
			return "completed"
		} else {
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
		}
	}()

	bar.Set(bar.Total)
}

func addTorrents(client *torrent.Client, torrentData []byte, bar *uiprogress.Bar) error {

	reader := bytes.NewReader(torrentData)

	mi, err := metainfo.Load(reader)
	if err != nil {
		return err
	}

	t, err := client.AddTorrent(mi)
	if err != nil {
		return err
	}

	torrentBar(t, bar)

	go func() {
		<-t.GotInfo()
		t.DownloadAll()
	}()

	return nil
}

func startDownLoad(torrentData []byte, output string, bar *uiprogress.Bar) error {

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	defer w.Close()
	f := func() { os.Stdout = old }
	defer f()

	//log.SetOutput(os.NewFile(uintptr(0), "NUL"))

	defer envpprof.Stop()

	clientConfig := torrent.NewDefaultClientConfig()

	clientConfig.DataDir = output
	clientConfig.Debug = false
	clientConfig.NoDHT = true
	clientConfig.DisableIPv6 = true
	clientConfig.NoDefaultPortForwarding = true

	client, err := torrent.NewClient(clientConfig)
	if err != nil {
		return err
	}

	defer client.Close()
	go exitSignalHandlers(client)

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		client.WriteStatus(w)
	})

	err = addTorrents(client, torrentData, bar)
	if err != nil {
		return err
	}

	if !client.WaitAll() {
		return fmt.Errorf("Download failed")
	}

	return nil
}
