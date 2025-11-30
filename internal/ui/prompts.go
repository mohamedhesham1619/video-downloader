package ui

import (
	"downloader/internal/models"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

// Prompt the user to select the preferred video format
func PromptVideoFormat() (models.VideoFormat, error) {
	var selectedOption string
	prompt := &survey.Select{
		Message: "Choose your preferred video format:",
		Options: []string{
			"Any format",
			"Prefer MP4 when available",
			"Force MP4 (convert if necessary)",
		},
	}

	err := survey.AskOne(prompt, &selectedOption)
	if err != nil {
		return models.FormatAny, err
	}

	switch selectedOption {
	case "Prefer MP4 when available":
		return models.FormatPreferMP4, nil
	case "Force MP4 (convert if necessary)":
		return models.FormatForceMP4, nil
	default:
		return models.FormatAny, nil
	}
}

// Prompt the user to select the download method for clips (should re-encode or not)
func PromptClipDownloadMethod() (shouldReEncode bool, err error) {
	var selectedOption string
	prompt := &survey.Select{
		Message: "How would you like to download clips?",
		Options: []string{
			"Fast (recommended) - Clips may start a few seconds early or have frozen frames at the start",
			"Accurate - Switch to this if Fast didn't work properly (much slower)",
		},
	}

	err = survey.AskOne(prompt, &selectedOption)
	if err != nil {
		return false, err
	}

	shouldReEncode = strings.Contains(selectedOption, "Accurate")

	return shouldReEncode, nil
}
