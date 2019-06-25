package update

import (
	"time"
)

type DownloadStatistics struct {
	ID                      string
	StartTime               time.Time
	DownloadTimeAtStart     time.Time
	DownloadTime            time.Time
	ApplicationRestartCount int
	MaxDownloadSpeed        uint64
	MaxUploadSpeed          uint64
	TotalSize               uint64
	Started                 bool
}

func NewDownloadStatistics(id string) *DownloadStatistics {
	return &DownloadStatistics{ID: id}
}

func (stats *DownloadStatistics) Start() {

	if stats.Started {
		return
	}

	stats.ApplicationRestartCount++
	stats.DownloadTimeAtStart = stats.DownloadTime
	stats.StartTime = time.Now()
	stats.Started = true
}

func (stats *DownloadStatistics) Stop() {

	if !stats.Started {
		return
	}

	stats.Started = false
	duration := time.Until(stats.StartTime)
	stats.DownloadTime = stats.DownloadTimeAtStart.Add(duration)
}

func (stats *DownloadStatistics) Update(downloadSpeed uint64, uploadSpeed uint64, totalSize uint64) {

	if !stats.Started {
		return
	}

	duration := time.Until(stats.StartTime)
	stats.DownloadTime = stats.DownloadTimeAtStart.Add(duration)

	if downloadSpeed > stats.MaxDownloadSpeed {
		stats.MaxDownloadSpeed = downloadSpeed
	}

	if uploadSpeed > stats.MaxUploadSpeed {
		stats.MaxUploadSpeed = uploadSpeed
	}

	if totalSize > stats.TotalSize {
		stats.TotalSize = totalSize
	}
}
