package models

type DownloadRequest struct {
	Url           string
	Quality       string 
	IsClip        bool
	ClipTimeRange string // should be in the format HH:MM:SS-HH:MM:SS
}