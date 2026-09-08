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

	// If a URL appears multiple times, assign a 1-based sequential index to each request
	urlCounts := make(map[string]int)
	for _, req := range downloadRequests {
		urlCounts[req.Url]++
	}

	urlCurrentIndex := make(map[string]int)
	for i := range downloadRequests {
		if urlCounts[downloadRequests[i].Url] > 1 {
			urlCurrentIndex[downloadRequests[i].Url]++
			downloadRequests[i].Index = urlCurrentIndex[downloadRequests[i].Url]
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
			used := make([]bool, len(downloadRequests))
			retryRequests := make([]models.DownloadRequest, 0, len(signInErrors))
			for _, e := range signInErrors {
				for j, req := range downloadRequests {
					if !used[j] && req.Url == e.URL {
						used[j] = true
						retryRequests = append(retryRequests, req)
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

// buildProgressLabels returns the display lines for a download request.
func buildProgressLabels(req models.DownloadRequest) []string {
	var lines []string

	indexSuffix := ""
	if req.Index > 0 {
		indexSuffix = fmt.Sprintf(" #%d", req.Index)
	}

	if req.IsAudioOnly {
		if req.IsClip {
			durationText := utils.FormatClipDurationText(req.ClipTimeRange)
			lines = append(lines, fmt.Sprintf("Downloading audio clip%s %s", indexSuffix, color.CyanString("(best quality)")))
			lines = append(lines, fmt.Sprintf("Duration: %s", durationText))
			lines = append(lines, fmt.Sprintf("URL: %s", req.Url))
		} else {
			lines = append(lines, fmt.Sprintf("Downloading full audio%s %s", indexSuffix, color.CyanString("(best quality)")))
			lines = append(lines, fmt.Sprintf("URL: %s", req.Url))
		}
		return lines
	}

	quality := "(best quality)"
	if req.Quality != "" {
		quality = fmt.Sprintf("(%sp)", req.Quality)
	}
	if req.IsClip {
		durationText := utils.FormatClipDurationText(req.ClipTimeRange)
		lines = append(lines, fmt.Sprintf("Downloading clip%s %s", indexSuffix, color.CyanString(quality)))
		lines = append(lines, fmt.Sprintf("Duration: %s", durationText))
		lines = append(lines, fmt.Sprintf("URL: %s", req.Url))
	} else {
		lines = append(lines, fmt.Sprintf("Downloading full video%s %s", indexSuffix, color.CyanString(quality)))
		lines = append(lines, fmt.Sprintf("URL: %s", req.Url))
	}
	return lines
}

// runDownloads runs a set of download requests concurrently and waits for all to finish.
func runDownloads(dl *downloader.Downloader, requests []models.DownloadRequest) {
	progress := uiprogress.New()

	bars := make([]*uiprogress.Bar, len(requests))
	for i, req := range requests {
		// Add an empty line before each block
		ui.AddLabelBar(progress, "")
		
		// Add each line of the label as a text-only progress bar
		for _, line := range buildProgressLabels(req) {
			ui.AddLabelBar(progress, line)
		}
		
		// Add the actual progress bar
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
