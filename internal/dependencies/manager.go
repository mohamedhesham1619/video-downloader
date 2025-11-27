package dependencies

import (
	"downloader/internal/ui"
	"fmt"
	"sync"
)

// EnsureReady checks for required external dependencies (yt-dlp, ffmpeg, deno),
// download the missing dependencies,
// and ensure yt-dlp is up-to-date
func EnsureReady() error {
	// Print header
	fmt.Println("Verifying dependencies...")
	fmt.Println()

	programs := []string{"yt-dlp", "ffmpeg", "deno"}
	progressLines := make(map[string]*ui.ProgressLine)
	ytdlpExists := false

	var wg sync.WaitGroup

	// Phase 1: Check all programs in parallel
	for _, program := range programs {
		wg.Add(1)
		progressLines[program] = ui.ShowLoading("Checking for " + program)

		go func() {
			defer wg.Done()

			if programExists(program) {
				progressLines[program].Complete(program + " found")
				if program == "yt-dlp" {
					ytdlpExists = true
				}
			} else {
				progressLines[program].Fail(program + " not found")

				// Download missing program
				downloadProgressLine := ui.ShowLoading("Downloading " + program)
				var err error

				switch program {
				case "yt-dlp":
					err = downloadYtDlp()
				case "ffmpeg":
					err = downloadAndExtractFfmpeg()
				case "deno":
					err = DownloadAndExtractDeno()
				}

				if err != nil {
					downloadProgressLine.Fail("Failed to download " + program)
				} else {
					downloadProgressLine.Complete("Downloaded " + program)
				}
			}
		}()
	}

	wg.Wait()

	// Stop multi-printer to add spacing
	ui.StopMultiPrinter()

	// Add blank line between phases
	fmt.Println()

	// Phase 2: Check for yt-dlp updates (only if it already existed)
	if ytdlpExists {
		updateLine := ui.ShowLoading("Checking for yt-dlp updates")

		hasUpdate, currentVersion, latestVersion, err := isYtDlpUpdateAvailable()
		if err != nil {
			updateLine.Fail("Failed to check for yt-dlp updates: " + err.Error())
			ui.StopMultiPrinter()
			return err
		}

		if hasUpdate {
			updateLine.Fail("Update available")

			// Download the update
			downloadLine := ui.ShowLoading("Updating yt-dlp from " + currentVersion + " to " + latestVersion)

			if err := updateYtDlp(); err != nil {
				downloadLine.Fail("Failed to update yt-dlp: " + err.Error())
				ui.StopMultiPrinter()
				return err
			}

			downloadLine.Complete("Updated yt-dlp to " + latestVersion)
		} else {
			updateLine.Complete("yt-dlp is up to date (" + currentVersion + ")")
		}
	}

	// Stop the multi-printer at the end
	ui.StopMultiPrinter()

	// Print footer
	fmt.Println()
	fmt.Println("Dependencies ready!")
	fmt.Println("----------------------------------------")
	fmt.Println()

	return nil
}
