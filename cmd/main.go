package main

import (
	"bufio"
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

	// prompt the user to select the preferred video format
	preferredFormat, err := ui.PromptVideoFormat()
	if err != nil {
		log.Fatal("Error prompting video format", err)
	}

	// if there is any clip request, prompt the user to select the clip download method
	shouldReEncode := false

	if shouldPromptClipDownloadMethods {
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

	// start downloading videos concurrently
	wg := sync.WaitGroup{}
	wg.Add(len(downloadRequests))

	for _, downloadRequest := range downloadRequests {
		go func() {
			err := downloader.Download(downloadRequest)

			if err != nil {
				fmt.Printf("%s failed to download video(%s): %v\n", color.RedString("Error:"), downloadRequest.Url, err)
			}
			wg.Done()
		}()
	}

	// ensure all goroutines complete
	wg.Wait()

	// Wait for user input before exiting
	fmt.Print("\nAll downloads completed. Press Enter to exit...\n")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
