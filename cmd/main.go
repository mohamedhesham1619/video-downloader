package main

import (
	"bufio"
	"downloader/internal/config"
	"downloader/internal/dependencies"
	"downloader/internal/downloader"
	"downloader/internal/models"
	"downloader/internal/utils"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/AlecAivazis/survey/v2"
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

	// if there is any clip request, prompt the user to select the clip download method
	shouldReEncode := false

	if shouldPromptClipDownloadMethods {
		shouldReEncode, err = promptClipDownloadMethod()
		if err != nil {
			log.Fatal("Error prompting clip download method", err)
		}
	}

	// initialize config and downloader
	cfg := config.New(shouldReEncode)
	downloader := downloader.New(cfg)

	wg := sync.WaitGroup{}
	wg.Add(len(downloadRequests))

	// start downloading videos concurrently
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

// Prompt the user to select the download method for clips (should re-encode or not)
func promptClipDownloadMethod() (shouldReEncode bool, err error) {
	var selectedOption string
	prompt := &survey.Select{
		Message: "How would you like to download clips?",
		Options: []string{
			"⚡ Fast (recommended) - Clips may start a few seconds early or have frozen frames at the start",
			"🎯 Accurate - Switch to this if Fast didn't work properly (much slower)",
		},
	}

	err = survey.AskOne(prompt, &selectedOption)
	if err != nil {
		return false, err
	}

	shouldReEncode = strings.Contains(selectedOption, "Accurate")

	return shouldReEncode, nil
}
