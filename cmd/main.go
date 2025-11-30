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

	// parse urls into download requests and check if there are clip requests
	downloadRequests := make([]models.DownloadRequest, len(urls))
	shouldPromptClipDownloadMethods := false

	for i, url := range urls {
		downloadRequests[i] = utils.ParseDownloadRequest(url)
		if downloadRequests[i].IsClip {
			shouldPromptClipDownloadMethods = true
		}
	}

	// Show setup header
	fmt.Println("Quick setup before we start...")
	fmt.Println()

	// prompt the user to select the preferred video format
	preferredFormat, err := ui.PromptVideoFormat()
	if err != nil {
		log.Fatal("Error prompting video format", err)
	}

	// if there is any clip request, prompt the user to select the clip download method
	shouldReEncode := false

	if shouldPromptClipDownloadMethods {
		fmt.Println()
		shouldReEncode, err = ui.PromptClipDownloadMethod()
		if err != nil {
			log.Fatal("Error prompting clip download method", err)
		}
	}

	// initialize config and downloader
	cfg := config.New(shouldReEncode, preferredFormat)
	downloader := downloader.New(cfg)

	// Add spacing between prompts and downloads
	fmt.Println()
	fmt.Println("Starting downloads...")
	fmt.Println("----------------------------------------")
	fmt.Println("Please keep the app open until you see “All downloads completed”. This ensures every download finishes correctly.")
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

	// Start the progress rendering system
	uiprogress.Start()

	// start downloading videos concurrently
	wg := sync.WaitGroup{}
	wg.Add(len(downloadRequests))

	for _, downloadRequest := range downloadRequests {
		go func() {

			quality := ""
			if downloadRequest.Quality != "" {
				quality = fmt.Sprintf("(%sp)", downloadRequest.Quality)
			} else {
				quality = "(best quality)"
			}

			// Prepare the progress label based on the download request type
			progressLabel := "\n"

			if downloadRequest.IsClip {
				durationText := utils.FormatClipDurationText(downloadRequest.ClipTimeRange)
				progressLabel += fmt.Sprintf("Downloading clip %s\nDuration: %s\nURL: %s", color.CyanString(quality), durationText, downloadRequest.Url)
			} else {
				progressLabel += fmt.Sprintf("Downloading full video %s\nURL: %s", color.CyanString(quality), downloadRequest.Url)
			}

			// Show the progress bar
			downloadProgressBar := ui.ShowDownloadProgress(progressLabel)

			// Start the download and get the progress channel
			progressChan := downloader.Download(downloadRequest)

			// Update the progress bar with the progress from the progress channel
			for progress := range progressChan {
				downloadProgressBar.Set(progress)
			}

			// Signal that the download process is complete
			wg.Done()
		}()
	}

	// ensure all goroutines complete
	wg.Wait()

	// Stop the progress rendering system
	uiprogress.Stop()

	// If there are errors, show them and wait for user input before exiting
	if downloader.ErrorCollector.HasErrors() {
		errors := downloader.ErrorCollector.GetAll()
		fmt.Println()
		fmt.Println("----------------------------------------")
		fmt.Println(color.RedString("Errors:"))
		fmt.Println()
		for _, err := range errors {
			fmt.Println(err)
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
