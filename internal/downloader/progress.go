package downloader

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
)

func (d *Downloader) streamClipDownloadProgress(stderrPipe, stdoutPipe io.ReadCloser, clipDurationInSeconds int, progressChan chan int) {

	// Regex to match ffmpeg time output: time=00:00:05.84
	re := regexp.MustCompile(`time=(\d{2}):(\d{2}):(\d{2})`)

	// Regex to match errors
	errorRegex := regexp.MustCompile(`ERROR:\s*(.+)`)

	// Read stdout for errors in a separate goroutine
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			if errorMatch := errorRegex.FindStringSubmatch(line); errorMatch != nil {
				d.ErrorCollector.Add(errorMatch[1])
			}
		}
	}()

	// We need to read byte by byte because yt-dlp (and ffmpeg) use \r to update progress inline.
	reader := bufio.NewReader(stderrPipe)
	var line []byte

	for {
		// Read one byte at a time
		b, err := reader.ReadByte()
		if err != nil {
			break
		}

		// If we hit a delimiter, process the accumulated line
		if b == '\r' || b == '\n' {
			if len(line) > 0 {
				lineStr := string(line)

				// Check for errors in stderr too
				if errorMatch := errorRegex.FindStringSubmatch(lineStr); errorMatch != nil {
					d.ErrorCollector.Add(errorMatch[1])
					line = nil
					continue
				}

				// Parse progress
				match := re.FindStringSubmatch(lineStr)
				if len(match) == 4 {
					hours, _ := strconv.Atoi(match[1])
					minutes, _ := strconv.Atoi(match[2])
					seconds, _ := strconv.Atoi(match[3])

					processedTime := hours*3600 + minutes*60 + seconds
					percentage := (processedTime * 100) / clipDurationInSeconds

					if percentage >= 100 {
						percentage = 100
					}

					progressChan <- percentage
				}
				line = nil // Reset line buffer
			}
		} else {
			line = append(line, b)
		}
	}
}

func (d *Downloader) streamFullDownloadProgress(stderrPipe, stdoutPipe io.ReadCloser, progressChan chan int) {

	// Pattern 1: Fragment-based progress (frag N/M)
	// Example: [download]   6.5% of ~  20.20MiB at  889.24KiB/s ETA Unknown (frag 1/38)
	fragmentRegex := regexp.MustCompile(`\[download\].*?\(frag\s+(\d+)/(\d+)\)`)
	
	// Pattern 2: Simple percentage progress
	// Example: [download]  21.2% of    9.13MiB at    2.35MiB/s ETA 00:03
	percentRegex := regexp.MustCompile(`\[download\]\s+(\d+(?:\.\d+)?)%`)

	// Pattern: ERROR: Some error message
	errorRegex := regexp.MustCompile(`ERROR:\s*(.+)`)

	// Read stderr for errors in a separate goroutine
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			if errorMatch := errorRegex.FindStringSubmatch(line); errorMatch != nil {
				d.ErrorCollector.Add(errorMatch[1])
			}
		}
	}()

	lastPercentage := 0
	maxFragmentSeen := 0

	// yt-dlp writes progress to stdout when --newline is used
	scanner := bufio.NewScanner(stdoutPipe)
	for scanner.Scan() {
		line := scanner.Text()

		// Check for errors in stdout too
		if errorMatches := errorRegex.FindStringSubmatch(line); errorMatches != nil {
			d.ErrorCollector.Add(errorMatches[1])
			continue
		}

		// Try fragment-based progress first (for fragmented streams)
		if matches := fragmentRegex.FindStringSubmatch(line); matches != nil {
			currentFrag, _ := strconv.Atoi(matches[1])
			totalFrags, _ := strconv.Atoi(matches[2])
			
			if currentFrag > maxFragmentSeen {
				maxFragmentSeen = currentFrag
			}
			
			overallProgress := int(float64(maxFragmentSeen) / float64(totalFrags) * 100)
			
			if overallProgress > lastPercentage {
				lastPercentage = overallProgress
				progressChan <- overallProgress
			}
		} else if matches := percentRegex.FindStringSubmatch(line); matches != nil {
			// Simple percentage progress (for non-fragmented streams)
			percentage, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				currentPercentage := int(percentage)
				if currentPercentage > lastPercentage {
					lastPercentage = currentPercentage
					progressChan <- currentPercentage
				}
			}
		}
	}
}