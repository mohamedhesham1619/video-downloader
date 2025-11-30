package utils

import (
	"bufio"
	"downloader/internal/models"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/fatih/color"
)

// ReadLinesFromFile reads lines from a file and returns them as a slice of strings.
// It ignores empty lines and removes leading and trailing whitespace from each line.
func ReadLinesFromFile(fileName string) ([]string, error) {
	file, err := os.Open(fileName)

	if err != nil {
		return []string{}, fmt.Errorf("couldn't open the file: %v", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var urls []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line != "" {
			urls = append(urls, line)
		}
	}
	return urls, nil
}

// create a download request object from a line of text
// the line must follow these rules:
// - the first part is the url
// - for clip download, the line must contain a time range in the format HH:MM:SS-HH:MM:SS
// - for both clip and full video download, the quality can be specified using any number with "p" suffix (e.g., 1440p,1080p, 720p)
//
// Examples:
// - https://www.video.com/watch?v=dQw4w9WgXcQ    (download the full video in best quality)
// - https://www.video.com/watch?v=dQw4w9WgXcQ 1080p    (download the full video in 1080p quality)
// - https://www.video.com/watch?v=dQw4w9WgXcQ 00:00:00-00:01:00    (download a clip from 00:00:00 to 00:01:00)
// - https://www.video.com/watch?v=dQw4w9WgXcQ 1080p 00:00:00-00:01:00    (download a clip from 00:00:00 to 00:01:00 in 1080p quality)
func ParseDownloadRequest(line string) models.DownloadRequest {

	// split the line by spaces
	parts := strings.Fields(line)

	// the first part is the url
	req := models.DownloadRequest{
		Url: parts[0],
	}

	// if the line contains a time range or quality, add it to the request
	if len(parts) > 1 {
		for i := 1; i < len(parts); i++ {
			if strings.Contains(parts[i], "-") {
				req.IsClip = true
				req.ClipTimeRange = parts[i]
			} else if strings.HasSuffix(parts[i], "p") {
				req.Quality = strings.TrimSuffix(parts[i], "p")
			}
		}
	}

	return req
}

// CalculateClipDurationInSeconds the clip duration in seconds from a time range in format "hh:mm:ss-hh:mm:ss"
func CalculateClipDurationInSeconds(timeRange string) (int, error) {
	// Split the time range string into start and end times
	parts := strings.Split(timeRange, "-")

	// Parse the start and end times
	startTime, err := time.Parse("15:04:05", parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid start time format: %v", err)
	}

	endTime, err := time.Parse("15:04:05", parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid end time format: %v", err)
	}

	// Calculate the duration in seconds
	duration := int(endTime.Sub(startTime).Seconds())

	return duration, nil
}

// sanitize the filename to remove or replace characters that are problematic in filenames
func SanitizeFilename(filename string) string {

	replacements := map[rune]rune{
		'/':  '-',
		'\\': '-',
		':':  '-',
		'*':  '-',
		'?':  '-',
		'"':  '-',
		'<':  '-',
		'>':  '-',
		'|':  '-',
	}

	sanitized := []rune{}
	for _, r := range filename {
		if replaced, exists := replacements[r]; exists {
			sanitized = append(sanitized, replaced)
		} else if unicode.IsPrint(r) {
			sanitized = append(sanitized, r)
		}
	}

	return string(sanitized)
}

// Returns the absolute path to the binary to be executed based on the OS
func GetBinaryPath(binaryName string) string {
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	binPath := filepath.Join("bin", binaryName)
	abs, err := filepath.Abs(binPath)
	if err != nil {
		return binaryName
	}
	return abs
}

// FormatDuration formats a duration in seconds to a human-readable string (e.g., "2m 30s")
func FormatDuration(seconds int) string {
	m := seconds / 60
	s := seconds % 60
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// Formats a user-friendly duration text for clip downloads
func FormatClipDurationText(timeRange string) string {

	durationSecs, _ := CalculateClipDurationInSeconds(timeRange)
	startTime, endTime := strings.Split(timeRange, "-")[0], strings.Split(timeRange, "-")[1]

	return fmt.Sprintf("%s (from %s to %s)",
		color.YellowString(FormatDuration(durationSecs)),
		startTime,
		endTime)
}

// IsYouTubeURL returns true if the URL is a YouTube link.
func IsYouTubeURL(url string) bool {
	return strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be")
}
