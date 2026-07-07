package main

import (
	"downloader/internal/config"
	"downloader/internal/dependencies"
	"downloader/internal/downloader"
	"downloader/internal/models"
	"downloader/internal/ui"
	"downloader/internal/utils"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/gosuri/uiprogress"
)

func main() {

	// Ensure all needed dependencies are ready
	err := dependencies.EnsureReady()
	if err != nil {
		log.Fatal(err)
	}

	// check if the urls.txt file exists
	if _, err := os.Stat("urls.txt"); os.IsNotExist(err) {
		log.Fatal("urls.txt file not found in the current directory")
	}

	// read urls from file
	urls, err := utils.ReadLinesFromFile("urls.txt")

	if err != nil {
		log.Fatal("Error reading urls from urls.txt file \n", err)
	}

	// parse urls into download requests and check if there are video clip requests
	downloadRequests := make([]models.DownloadRequest, len(urls))
	hasVideoRequests := false
	hasVideoClipRequests := false

	for i, url := range urls {
		downloadRequests[i] = utils.ParseDownloadRequest(url)
		if !downloadRequests[i].IsAudioOnly {
			hasVideoRequests = true
			if downloadRequests[i].IsClip {
				hasVideoClipRequests = true
			}
		}
	}

	// Only show setup prompts if there are video requests
	preferredFormat := models.FormatAny
	shouldReEncode := false

	if hasVideoRequests {
		// Show setup header
		fmt.Println("Quick setup before we start...")
		fmt.Println()

		// prompt the user to select the preferred video format
		var err error
		preferredFormat, err = ui.PromptVideoFormat()
		if err != nil {
			log.Fatal("Error prompting video format", err)
		}

		// if there is any video clip request, prompt the user to select the clip download method
		if hasVideoClipRequests {
			fmt.Println()
			shouldReEncode, err = ui.PromptClipDownloadMethod()
			if err != nil {
				log.Fatal("Error prompting clip download method", err)
			}
		}
	}

	// initialize config and downloader
	cfg := config.New(shouldReEncode, preferredFormat)
	dl := downloader.New(cfg)

	// Add spacing between prompts and downloads
	fmt.Println()
	fmt.Println("Starting downloads...")
	fmt.Println("----------------------------------------")
	fmt.Println("Please keep the app open until you see \u201cAll downloads completed\u201d. This ensures every download finishes correctly.")
	fmt.Println()

	// Print the encoder that will be used for clips
	if shouldReEncode {
		if cfg.Encoder == config.CPUEncoder {
			color.Cyan("Could not use GPU encoder. Falling back to CPU encoder: %s\n", cfg.Encoder)
		} else {
			color.Cyan("Using GPU encoder: %s\n", cfg.Encoder)
		}
		fmt.Println()
	}

	// --- First pass: download everything ---
	runDownloads(dl, downloadRequests)

	// --- Cookie retry pass (YouTube sign-in failures only) ---
	signInErrors := dl.ErrorCollector.GetSignInErrors()
	if len(signInErrors) > 0 {
		fmt.Println()
		fmt.Println("----------------------------------------")

		retry, err := ui.PromptCookieRetry(len(signInErrors))
		if err != nil {
			log.Fatal("Error prompting cookie retry", err)
		}

		if retry {
			browser, err := ui.PromptBrowser()
			if err != nil {
				log.Fatal("Error prompting browser selection", err)
			}

			// Remove the sign-in errors so they don't show up in the final report
			for _, e := range signInErrors {
				dl.ErrorCollector.RemoveByURL(e.URL)
			}

			// Build retry requests from the failed URLs
			retryRequests := make([]models.DownloadRequest, len(signInErrors))
			for i, e := range signInErrors {
				for _, req := range downloadRequests {
					if req.Url == e.URL {
						retryRequests[i] = req
						break
					}
				}
			}

			// Create a new downloader with cookies configured, reusing the same error collector
			cookieCfg := cfg.WithCookiesBrowser(browser)
			cookieDl := downloader.NewWithErrorCollector(cookieCfg, dl.ErrorCollector)

			fmt.Println()
			fmt.Println("Retrying with browser cookies...")
			fmt.Println("----------------------------------------")
			fmt.Println()

			runDownloads(cookieDl, retryRequests)
		}
	}

	// --- Final result ---
	if dl.ErrorCollector.HasErrors() {
		errors := dl.ErrorCollector.GetAll()
		fmt.Println()
		fmt.Println("----------------------------------------")
		fmt.Println(color.RedString("Errors:"))
		fmt.Println()
		for _, err := range errors {
			fmt.Printf("URL: %s\n%s\n", err.URL, err.Message)
			fmt.Println("-------------------------")
		}
		fmt.Println()
		fmt.Println("All downloads completed.")
		var input string
		fmt.Scanln(&input)
	} else {
		fmt.Println()
		fmt.Println("All downloads completed successfully.")
		var input string
		fmt.Scanln(&input)
	}
}

// buildProgressLabel returns a display label for a download request.
func buildProgressLabel(req models.DownloadRequest) string {
	if req.IsAudioOnly {
		if req.IsClip {
			durationText := utils.FormatClipDurationText(req.ClipTimeRange)
			return fmt.Sprintf("Downloading audio clip %s\nDuration: %s\nURL: %s", color.CyanString("(best quality)"), durationText, req.Url)
		}
		return fmt.Sprintf("Downloading full audio %s\nURL: %s", color.CyanString("(best quality)"), req.Url)
	}

	quality := "(best quality)"
	if req.Quality != "" {
		quality = fmt.Sprintf("(%sp)", req.Quality)
	}
	if req.IsClip {
		durationText := utils.FormatClipDurationText(req.ClipTimeRange)
		return fmt.Sprintf("Downloading clip %s\nDuration: %s\nURL: %s", color.CyanString(quality), durationText, req.Url)
	}
	return fmt.Sprintf("Downloading full video %s\nURL: %s", color.CyanString(quality), req.Url)
}

// runDownloads runs a set of download requests concurrently and waits for all to finish.
func runDownloads(dl *downloader.Downloader, requests []models.DownloadRequest) {
	progress := uiprogress.New()

	// Print all download labels as plain output BEFORE progress starts.
	// Since bars are single-line, uiprogress re-renders only overwrite the
	// progress lines themselves — labels above them stay intact.
	for _, req := range requests {
		fmt.Println(buildProgressLabel(req))
	}

	// Register all bars before Start() so the render loop sees them immediately.
	bars := make([]*uiprogress.Bar, len(requests))
	for i := range requests {
		bars[i] = ui.ShowDownloadProgress(progress)
	}

	progress.Start()

	wg := sync.WaitGroup{}
	wg.Add(len(requests))

	for i, req := range requests {
		go func(bar *uiprogress.Bar, req models.DownloadRequest) {
			progressChan := dl.Download(req)
			for pct := range progressChan {
				bar.Set(pct)
			}
			wg.Done()
		}(bars[i], req)
	}

	wg.Wait()
	progress.Stop()
}
