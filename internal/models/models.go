package models

type VideoFormat int

const (
	FormatAny       VideoFormat = iota // Any format
	FormatPreferMP4                    // Prefer MP4 when available
	FormatForceMP4                     // Force MP4 (convert if necessary)
)

type DownloadRequest struct {
	Url           string
	Quality       string
	IsClip        bool
	ClipTimeRange string // should be in the format HH:MM:SS-HH:MM:SS
	IsAudioOnly   bool
	Index         int // 1-based index if multiple downloads exist for the same URL; 0 if single
}
