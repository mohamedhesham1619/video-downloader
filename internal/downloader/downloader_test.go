package downloader

import (
	"downloader/internal/config"
	"downloader/internal/models"
	"path/filepath"
	"testing"
)

func TestBuildOutputPath(t *testing.T) {
	cfg := config.New(false, models.FormatAny)
	dl := New(cfg)

	tests := []struct {
		name        string
		isAudioOnly bool
		index       int
		expected    string
	}{
		{
			name:        "single video download",
			isAudioOnly: false,
			index:       0,
			expected:    filepath.Join(cfg.DownloadPath, "%(title).50s-%(height)sp.%(ext)s"),
		},
		{
			name:        "multiple video download index 1",
			isAudioOnly: false,
			index:       1,
			expected:    filepath.Join(cfg.DownloadPath, "%(title).50s-1-%(height)sp.%(ext)s"),
		},
		{
			name:        "multiple video download index 2",
			isAudioOnly: false,
			index:       2,
			expected:    filepath.Join(cfg.DownloadPath, "%(title).50s-2-%(height)sp.%(ext)s"),
		},
		{
			name:        "single audio download",
			isAudioOnly: true,
			index:       0,
			expected:    filepath.Join(cfg.DownloadPath, "%(title).50s-audio.%(ext)s"),
		},
		{
			name:        "multiple audio download index 1",
			isAudioOnly: true,
			index:       1,
			expected:    filepath.Join(cfg.DownloadPath, "%(title).50s-1-audio.%(ext)s"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := dl.buildOutputPath(tt.isAudioOnly, tt.index)
			if actual != tt.expected {
				t.Errorf("got %q, expected %q", actual, tt.expected)
			}
		})
	}
}
