package downloader

import (
	"downloader/internal/config"
	"downloader/internal/models"
	"downloader/internal/utils"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
)

type Downloader struct {
	Config *config.Config
}

func New(cfg *config.Config) *Downloader {
	return &Downloader{
		Config: cfg,
	}
}

func (d *Downloader) Download(video models.DownloadRequest) error {
	var command *exec.Cmd
	var err error

	if video.IsClip {
		command = d.buildClipDownloadCommand(video)
	} else {
		command = d.buildFullDownloadCommand(video)
	}

	// Run the command
	if video.IsClip {
		fmt.Println(utils.FormatClipDownloadMessage(video.ClipTimeRange))
		fmt.Printf("From URL: %s\n\n", video.Url)
	} else {
		fmt.Printf("Downloading video: %s\n\n", video.Url)
	}
	output, err := command.CombinedOutput()

	if err != nil {
		return fmt.Errorf("%v\nOutput: %s", err, string(output))
	}

	if video.IsClip {
		fmt.Printf("%s Clip from %s\n\n", color.GreenString("Download completed:"), video.Url)
	} else {
		fmt.Printf("%s %s\n\n", color.GreenString("Download completed:"), video.Url)
	}
	return nil
}

// prepare the command to download the whole video
func (d *Downloader) buildFullDownloadCommand(req models.DownloadRequest) *exec.Cmd {

	// yt-dlp output template: "%(title).244s.%(ext)s"
	// - %(title)s: video title from metadata
	// - .244s: limits title to 244 characters to avoid path length issues on Windows
	// - %(ext)s: file extension based on selected format
	downloadPath := filepath.Join(d.Config.DownloadPath, "%(title).244s.%(ext)s")

	cmd := exec.Command(
		utils.GetCommand("yt-dlp"),
		"-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best", // This will get the best quality
		"-o", downloadPath,
		req.Url)

	return cmd
}

// prepare the command to download a clip of the video
func (d *Downloader) buildClipDownloadCommand(req models.DownloadRequest) *exec.Cmd {

	// Prepare the download path with the video title
	// yt-dlp output template: "%(title).244s.%(ext)s"
	// - %(title)s: video title from metadata
	// - .244s: limits title to 244 characters to avoid path length issues on Windows
	// - %(ext)s: file extension based on selected format
	downloadPath := filepath.Join(d.Config.DownloadPath, "%(title).244s.%(ext)s")

	// Prepare the command arguments
	args := []string{
		"-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
		"--merge-output-format", "mp4",
		"--download-sections", fmt.Sprintf("*%s", req.ClipTimeRange),
		"-o", downloadPath,
		req.Url,
	}

	// If the user choose to re-encode clips, add --postprocessor-args to force re-encoding with the selected encoder
	if d.Config.ShouldReEncode {
		args = append(args, "--postprocessor-args", fmt.Sprintf("ffmpeg=-c:v %s", d.Config.Encoder))
	}

	cmd := exec.Command(utils.GetCommand("yt-dlp"), args...)
	return cmd
}
