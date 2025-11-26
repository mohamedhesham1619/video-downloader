package utils

import (
	"bufio"
	"downloader/internal/models"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
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

// parse clip timing info
// for ffmpeg to accurately extract the needed clip, it needs the start time and clip duration in seconds
func ParseClipDuration(timeRange string) (start string, duration string, err error) {
	// Split the range into start and end times
	parts := strings.Split(timeRange, "-")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid time range format. Expected HH:MM:SS-HH:MM:SS")
	}

	startTime := parts[0]
	endTime := parts[1]

	// Parse times to calculate duration
	layout := "15:04:05"

	t1, err := time.Parse(layout, startTime)
	if err != nil {
		return "", "", fmt.Errorf("invalid start time: %v", err)
	}

	t2, err := time.Parse(layout, endTime)
	if err != nil {
		return "", "", fmt.Errorf("invalid end time: %v", err)
	}

	// Calculate duration in seconds
	durationSeconds := int(t2.Sub(t1).Seconds())

	// Convert duration to string
	duration = strconv.Itoa(durationSeconds)

	return startTime, duration, nil
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

// Returns the absolute path to the command to be executed based on the OS
func GetCommand(commandName string) string {
	if runtime.GOOS == "windows" {
		commandName += ".exe"
	}
	
    binPath := filepath.Join("bin", commandName)
	abs, err := filepath.Abs(binPath)
	if err != nil {
		return commandName
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

// Formats a user-friendly message for clip downloads
func FormatClipDownloadMessage(timeRange string) string {
	_, durationStr, err := ParseClipDuration(timeRange)
	if err != nil {
		return ""
	}

	durationSecs, _ := strconv.Atoi(durationStr)
	startTime, endTime := strings.Split(timeRange, "-")[0], strings.Split(timeRange, "-")[1]

	return fmt.Sprintf("Downloading clip: %s duration (from %s to %s)",
		color.CyanString(FormatDuration(durationSecs)),
		color.YellowString(startTime),
		color.YellowString(endTime))
}
