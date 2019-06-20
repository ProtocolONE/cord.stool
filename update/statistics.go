package update

import (
	"time"
)

type DownLoadStatistics struct {
	ID                      string
	StartTime               time.Time
	DownloadTimeAtStart     time.Time
	DownloadTime            time.Time
	ApplicationRestartCount int
	MaxDownloadSpeed        int
	MaxUploadSpeed          int
	TotalSize               uint64
	Started                 bool
}

func NewDownLoadStatistics(id string) *DownLoadStatistics {
	return &DownLoadStatistics{ID: id}
}

func (stats *DownLoadStatistics) Start() {

	if stats.Started {
		return
	}

	stats.ApplicationRestartCount++
	stats.DownloadTimeAtStart = stats.DownloadTime
	stats.StartTime = time.Now()
	stats.Started = true
}

func (stats *DownLoadStatistics) Stop() {

	if !stats.Started {
		return
	}

	stats.Started = false
	duration := time.Until(stats.StartTime)
	stats.DownloadTime = stats.DownloadTimeAtStart.Add(duration)
}

func (stats *DownLoadStatistics) Update(downloadSpeed int, uploadSpeed int, totalSize uint64) {

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
