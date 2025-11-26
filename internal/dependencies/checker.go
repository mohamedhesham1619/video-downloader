package dependencies

import (
	"downloader/internal/utils"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// programExists checks if a program exists in the bin directory
func programExists(programName string) bool {
	entries, err := os.ReadDir("bin")
	if err != nil {
		return false
	}

	searchName := strings.ToLower(programName)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.Contains(strings.ToLower(entry.Name()), searchName) {
			return true
		}
	}
	return false
}

// getLatestYtDlpVersion gets the latest yt-dlp version tag from GitHub releases
func getLatestYtDlpVersion() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

// getCurrentYtDlpVersion gets the existing yt-dlp version tag
func getCurrentYtDlpVersion() (string, error) {
	cmd := exec.Command(utils.GetCommand("yt-dlp"), "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// isYtDlpUpdateAvailable checks if a newer version of yt-dlp is available
func isYtDlpUpdateAvailable() (isUpdateAvailable bool, current string, latest string, err error) {
	current, err = getCurrentYtDlpVersion()
	if err != nil {
		return false, "", "", err
	}

	latest, err = getLatestYtDlpVersion()
	if err != nil {
		return false, "", "", err
	}
	isUpdateAvailable = current != latest

	return isUpdateAvailable, current, latest, nil
}