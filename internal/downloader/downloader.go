package downloader

import (
	"downloader/internal/config"
	"downloader/internal/models"
	"downloader/internal/utils"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
)

type Downloader struct {
	config         *config.Config
	ErrorCollector *errorCollector
}

func New(cfg *config.Config) *Downloader {
	return &Downloader{
		config:         cfg,
		ErrorCollector: &errorCollector{},
	}
}

func (d *Downloader) Download(videoRequest models.DownloadRequest) <-chan int {

	var downloadCommand *exec.Cmd
	progressChan := make(chan int)

	// Build the download command based on the request type and setup progress tracking
	if videoRequest.IsClip {
		// Calculate clip duration in seconds
		// This is needed to calculate the progress percentage
		clipDurationInSeconds, err := utils.CalculateClipDurationInSeconds(videoRequest.ClipTimeRange)

		if err != nil {
			d.ErrorCollector.Add(fmt.Sprintf("failed to calculate clip duration: %v", err))
			close(progressChan)
			return progressChan
		}

		// Build the download command
		downloadCommand = d.buildClipDownloadCommand(videoRequest)

		// Get the command pipes
		stdoutPipe, stderrPipe, err := getCommandPipes(downloadCommand)

		if err != nil {
			d.ErrorCollector.Add(err.Error())
			close(progressChan)
			return progressChan
		}

		// Start the progress tracking
		go d.streamClipDownloadProgress(stderrPipe, stdoutPipe, clipDurationInSeconds, progressChan)
	} else {
		downloadCommand = d.buildFullDownloadCommand(videoRequest)

		// Get the command pipes
		stdoutPipe, stderrPipe, err := getCommandPipes(downloadCommand)

		if err != nil {
			d.ErrorCollector.Add(err.Error())
			close(progressChan)
			return progressChan
		}

		// Start the progress tracking
		go d.streamFullDownloadProgress(stderrPipe, stdoutPipe, progressChan)
	}

	// Start the download
	err := downloadCommand.Start()

	if err != nil {
		d.ErrorCollector.Add(fmt.Sprintf("failed to start download: %v", err))
		close(progressChan)
		return progressChan
	}

	// Clean up process resources and close the progress channel when the download is finished
	go func() {
		downloadCommand.Wait()
		close(progressChan)
	}()

	return progressChan
}

// prepare the command to download the whole video
func (d *Downloader) buildFullDownloadCommand(req models.DownloadRequest) *exec.Cmd {

	var downloadPath string
	var format string

	if req.IsAudioOnly {
		// yt-dlp output template for audio: "%(title).150s-audio.%(ext)s"
		downloadPath = filepath.Join(d.config.DownloadPath, "%(title).150s-audio.%(ext)s")
		format = "ba"
	} else {
		// yt-dlp output template: "%(title).150s-%(height)sp.%(ext)s"
		// - %(title)s: video title from metadata
		// - .150s: limits title to 150 characters to avoid filename length issues
		// - %(height)sp: adds resolution height (e.g., 1080p, 720p)
		// - %(ext)s: file extension based on selected format
		downloadPath = filepath.Join(d.config.DownloadPath, "%(title).150s-%(height)sp.%(ext)s")

		isYoutubeUrl := utils.IsYouTubeURL(req.Url)
		format = getYtdlpFormat(isYoutubeUrl, req.Quality, d.config.VideoFormat)
	}

	args := []string{
		"-f", format,
		"--user-agent", "random",
		"--no-playlist",
		"--audio-quality", "0",
		"--socket-timeout", "20",
		"--retries", "3",
		"--retry-sleep", "3",
		"--force-overwrites",
		"--concurrent-fragments", "3",
		"--buffer-size", "64K",
		"--newline",
		"--ffmpeg-location", utils.GetBinaryPath("ffmpeg"),
		"--js-runtimes", utils.GetBinaryPath("deno"),
		"-o", downloadPath,
	}

	if !req.IsAudioOnly && d.config.VideoFormat == models.FormatForceMP4 {
		args = append(args, "--remux-video", "mp4")
	}

	args = append(args, req.Url)

	return exec.Command(utils.GetBinaryPath("yt-dlp"), args...)
}

// prepare the command to download a clip of the video
func (d *Downloader) buildClipDownloadCommand(req models.DownloadRequest) *exec.Cmd {

	var downloadPath string
	var format string

	if req.IsAudioOnly {
		// yt-dlp output template for audio: "%(title).150s-audio.%(ext)s"
		downloadPath = filepath.Join(d.config.DownloadPath, "%(title).150s-audio.%(ext)s")
		format = "ba"
	} else {
		// Prepare the download path with the video title
		// yt-dlp output template: "%(title).150s-%(height)sp.%(ext)s"
		// - %(title)s: video title from metadata
		// - .150s: limits title to 150 characters to avoid filename length issues
		// - %(height)sp: adds resolution height (e.g., 1080p, 720p)
		// - %(ext)s: file extension based on selected format
		downloadPath = filepath.Join(d.config.DownloadPath, "%(title).150s-%(height)sp.%(ext)s")

		isYouTubeURL := utils.IsYouTubeURL(req.Url)
		format = getYtdlpFormat(isYouTubeURL, req.Quality, d.config.VideoFormat)
	}

	// Prepare the command arguments
	args := []string{
		"-f", format,
		"--download-sections", fmt.Sprintf("*%s", req.ClipTimeRange),
		"--user-agent", "random",
		"--no-playlist",
		"--audio-quality", "0",
		"--socket-timeout", "20",
		"--retries", "3",
		"--retry-sleep", "3",
		"--force-overwrites",
		"--concurrent-fragments", "3",
		"--buffer-size", "64K",
		"--newline",
		"--ffmpeg-location", utils.GetBinaryPath("ffmpeg"),
		"--js-runtimes", utils.GetBinaryPath("deno"),
		"-o", downloadPath,
	}

	// Audio clips don't need re-encoding or remuxing
	if !req.IsAudioOnly {
		// If the user choose to re-encode clips, add --postprocessor-args to force re-encoding with the selected encoder
		if d.config.ShouldReEncode {
			args = append(args, "--postprocessor-args", fmt.Sprintf("ffmpeg=-c:v %s", d.config.Encoder))

			if d.config.VideoFormat == models.FormatForceMP4 {
				args = append(args, "--merge-output-format", "mp4")
			}
		} else {
			// Only remux if not re-encoding
			if d.config.VideoFormat == models.FormatForceMP4 {
				args = append(args, "--remux-video", "mp4")
			}
		}
	}

	args = append(args, req.Url)

	return exec.Command(utils.GetBinaryPath("yt-dlp"), args...)
}

func getCommandPipes(cmd *exec.Cmd) (stdoutPipe, stderrPipe io.ReadCloser, err error) {
	stdoutPipe, err = cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get stdout pipe: %v", err)
	}
	stderrPipe, err = cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get stderr pipe: %v", err)
	}
	return stdoutPipe, stderrPipe, nil
}
