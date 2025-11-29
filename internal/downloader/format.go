package downloader

import (
	"downloader/internal/models"
	"fmt"
)

func getYtdlpFormat(isYouTubeUrl bool, quality string, videoFormat models.VideoFormat) string {

	// Build format string based on platform and format preference
	switch videoFormat {
	case models.FormatAny:
		return buildFormatAny(isYouTubeUrl, quality)
	case models.FormatPreferMP4:
		return buildFormatPreferMP4(isYouTubeUrl, quality)
	case models.FormatForceMP4:
		return buildFormatForceMP4(isYouTubeUrl, quality)
	default:
		return buildFormatAny(isYouTubeUrl, quality)
	}
}

// Youtube often seperate the audio and video streams, so we need to prefer seperate streams to get the required video quality.
//
// For other sites, we can prefer the merged stream.
func buildFormatAny(isYouTube bool, quality string) string {
	if isYouTube {
		if quality != "" {
			return fmt.Sprintf("bv*[height<=%[1]s]+ba/best[height<=%[1]s]/bv*+ba/best", quality)
		}
		return "bv*+ba/best"
	}

	if quality != "" {
		return fmt.Sprintf("best[height<=%[1]s]/bv*[height<=%[1]s]+ba/best/bv*+ba", quality)
	}
	return "best/bv*+ba"
}

func buildFormatPreferMP4(isYouTube bool, quality string) string {
	if isYouTube {
		if quality != "" {
			return fmt.Sprintf("bv*[height<=%[1]s][ext=mp4]+ba[ext=m4a]/bv*[height<=%[1]s]+ba/best[height<=%[1]s]/bv*[ext=mp4]+ba[ext=m4a]/bv*+ba/best", quality)
		}
		return "bv*[ext=mp4]+ba[ext=m4a]/bv*+ba/best"
	}

	if quality != "" {
		return fmt.Sprintf("best[height<=%[1]s][ext=mp4]/best[height<=%[1]s]/best[ext=mp4]/best", quality)
	}
	return "best[ext=mp4]/best"
}

func buildFormatForceMP4(isYouTube bool, quality string) string {
	// Force MP4 uses same format as "Any" but adds --remux-video mp4 flag
	return buildFormatAny(isYouTube, quality)
}
