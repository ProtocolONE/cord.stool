package libtorrent

// #cgo LDFLAGS: -static-libstdc++ -static-libgcc -static -L . -l lt_wrapper -LD:/lt/libtorrent-1.2.1/bin/gcc-8.1.0/release/link-static/threading-multi -ltorrent -L"C:/Program Files/mingw-w64/x86_64-8.1.0-win32-seh-rt_v6-rev0/mingw64/x86_64-w64-mingw32/lib" -lws2_32 -lwsock32 -liphlpapi -L"C:/Program Files/mingw-w64/x86_64-8.1.0-posix-seh-rt_v6-rev0/mingw64/lib/gcc/x86_64-w64-mingw32/8.1.0"  -lstdc++ -lgcc
// #include <stdlib.h>
// #include "lt_wrapper.h"
import "C"

import (
	"unsafe"
	"fmt"
)

type Session struct {
	session      unsafe.Pointer
}

type Torrent struct {
	torrent      C.int
}

type TorentStatus struct {
	State int
	Paused bool
	Progress float64
	ErrorText string
	NextAnnounce int
	AnnounceInterval int
	CurrentTracker string
	TotalDownload uint64
	TotalUpload uint64
	TotalPayloadDownload uint64
	TotalPayloadUpload uint64
	TotalFailedBytes uint64
	TotalRedundantBytes uint64
	DownloadRate float64
	UploadRate float64
	DownloadPayloadRate float64
	UploadPayloadRate float64
	NumSeeds int
	NumPeers int
	NumComplete int
	NumIncomplete int
	ListSeeds int
	ListPeers int
	ConnectCandidates int
	NumPieces int
	TotalDone uint64
	TotalWantedDone uint64
	TotalWanted uint64
	DistributedCopies float64
	BlockSize int
	NumUploads int
	NumConnections int
	UploadsLimit int
	ConnectionsLimit int
	UpBandwidthQueue int
	DownBandwidthQueue int
	AllTimeUpload uint64
	AllTimeSownload uint64
	ActiveTime int
	SeedingTime int
	SeedRank int
	LastScrape int
	HasIncoming int
	SeedMode int
}

type SessionAlert struct {
    Category int
    Message string
}

func CreateSession() (*Session, error) {
	ses := C.session_create2(6881, 6889, -1, 0)
	if ses == nil {
		return nil, fmt.Errorf("Session creating failed")
	}
	return &Session{session: ses}, nil
}

func (session *Session) CloseSession() {

	C.session_close(session.session)
}
/*
func (session *Session) SessionPopAlerts() []*SessionAlert {

	var alerts *C.struct_session_alert_t
	var count C.int

	r :=  C.session_pop_alerts(session.session, alerts, &count)
	if r < 0 {
		return nil
	}

	defer C.free(unsafe.Pointer(alerts))

	var result []*SessionAlert
	var alert *SessionAlert

	for i := 0; i < int(count); i++ {

		a := []C.struct_session_alert_t(unsafe.Pointer(alerts))
		alert.Category = int(a[i].category)
		alert.Message = C.GoString(&a[i].msg[0])
		result = append(result, alert)
	}

	return result
}
*/
func (session *Session) AddTorrentFile(filename string, output string) (*Torrent, error) {

	fn := C.CString(filename)
  	defer C.free(unsafe.Pointer(fn))

	od := C.CString(output)
  	defer C.free(unsafe.Pointer(od))

	torrent := C.session_add_torrent_file(session.session, fn, od)
	if torrent < 0 {
		return nil, fmt.Errorf("Add torrent file failed")
	}

	return &Torrent{torrent: torrent}, nil
}

func (session *Session) AddTorrentData(data []byte, output string) (*Torrent, error) {

	td := unsafe.Pointer(&data[0])
  	//defer C.free(unsafe.Pointer(td))

	od := C.CString(output)
  	defer C.free(unsafe.Pointer(od))

	torrent := C.session_add_torrent_data(session.session, (*C.char)(td), C.int(len(data)), od)
	if torrent < 0 {
		return nil, fmt.Errorf("Add torrent data failed")
	}

	return &Torrent{torrent: torrent}, nil
}

func (torrent *Torrent) GetTorrentStatus() *TorentStatus {

	var status C.struct_torrent_status
	r := C.torrent_get_status2(torrent.torrent, &status)
	if r < 0 {
		return nil
	}

	var result TorentStatus
	
	result.State = int(status.state)
	result.Paused = status.paused > 0
	result.Progress = float64(status.progress)
	result.ErrorText = C.GoString(&status.error[0])
	result.NextAnnounce = int(status.next_announce)
	result.AnnounceInterval = int(status.announce_interval)
	result.CurrentTracker = C.GoString(&status.current_tracker[0])
	result.TotalDownload = uint64(status.total_download)
	result.TotalUpload = uint64(status.total_upload)
	result.TotalPayloadDownload = uint64(status.total_payload_download)
	result.TotalPayloadUpload = uint64(status.total_payload_upload)
	result.TotalFailedBytes = uint64(status.total_failed_bytes)
	result.TotalRedundantBytes = uint64(status.total_redundant_bytes)
	result.DownloadRate = float64(status.download_rate)
	result.UploadRate = float64(status.upload_rate)
	result.DownloadPayloadRate = float64(status.download_payload_rate)
	result.UploadPayloadRate = float64(status.upload_payload_rate)
	result.NumSeeds = int(status.num_seeds)
	result.NumPeers = int(status.num_peers)
	result.NumComplete = int(status.num_complete)
	result.NumIncomplete = int(status.num_incomplete)
	result.ListSeeds = int(status.list_seeds)
	result.ListPeers = int(status.list_peers)
	result.ConnectCandidates = int(status.connect_candidates)
	result.NumPieces = int(status.num_pieces)
	result.TotalDone = uint64(status.total_done)
	result.TotalWantedDone = uint64(status.total_wanted_done)
	result.TotalWanted = uint64(status.total_wanted)
	result.DistributedCopies = float64(status.distributed_copies)
	result.BlockSize = int(status.block_size)
	result.NumUploads = int(status.num_uploads)
	result.NumConnections = int(status.num_connections)
	result.UploadsLimit = int(status.uploads_limit)
	result.ConnectionsLimit = int(status.connections_limit)
	result.UpBandwidthQueue = int(status.up_bandwidth_queue)
	result.DownBandwidthQueue = int(status.down_bandwidth_queue)
	result.AllTimeUpload = uint64(status.all_time_upload)
	result.AllTimeSownload = uint64(status.all_time_download)
	result.ActiveTime = int(status.active_time)
	result.SeedingTime = int(status.seeding_time)
	result.SeedRank = int(status.seed_rank)
	result.LastScrape = int(status.last_scrape)
	result.HasIncoming = int(status.has_incoming)
	result.SeedMode = int(status.seed_mode)
	
	return &result
}
