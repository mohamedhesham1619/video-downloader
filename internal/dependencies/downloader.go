package dependencies

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// downloadFile downloads a file from url and saves it to dest
func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download failed: status %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// downloadYtDlp downloads the latest yt-dlp for the current OS
func downloadYtDlp() error {

	// download urls for different OSes
	downloadUrls := map[string]string{
		"windows": "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe",
		"linux":   "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux",
		"darwin":  "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos",
	}

	// get current OS type
	osType := runtime.GOOS

	// get download url for current OS
	downloadUrl, exist := downloadUrls[osType]
	if !exist {
		return fmt.Errorf("unsupported OS: %s", osType)
	}

	// determine filename for current OS
	var fileName string
	if osType == "windows" {
		fileName = "yt-dlp.exe"
	} else {
		fileName = "yt-dlp"
	}

	// find directory of the running executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	// create bin directory if it doesn't exist
	binDir := filepath.Join(execDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("cannot create bin directory: %w", err)
	}

	// download the file
	destPath := filepath.Join(binDir, fileName)

	if err := downloadFile(downloadUrl, destPath); err != nil {
		return fmt.Errorf("cannot download yt-dlp: %w", err)
	}

	return nil
}

// downloadAndExtractFfmpeg downloads the latest FFmpeg for the current OS,
// extracts the archive, places the ffmpeg binary into the bin folder,
// and removes the unneeded files.
func downloadAndExtractFfmpeg() error {

	// download urls for different OSes
	downloadUrls := map[string]string{
		"windows":      "https://github.com/BtbN/FFmpeg-Builds/releases/latest/download/ffmpeg-master-latest-win64-gpl.zip",
		"linux":        "https://github.com/BtbN/FFmpeg-Builds/releases/latest/download/ffmpeg-master-latest-linux64-gpl.tar.xz",
		"darwin-amd64": "https://ffmpeg.martin-riedl.de/download/macos/amd64/1764103068_N-121850-g2b221fdb4a/ffmpeg.zip",
		"darwin-arm64": "https://ffmpeg.martin-riedl.de/download/macos/arm64/1764095758_N-121850-g2b221fdb4a/ffmpeg.zip",
	}

	// get current OS type
	osType := runtime.GOOS

	if osType == "darwin" {
		osType = "darwin-" + runtime.GOARCH // "darwin-amd64" or "darwin-arm64"
	}

	// get download url for current OS
	downloadUrl, exist := downloadUrls[osType]
	if !exist {
		return fmt.Errorf("unsupported OS: %s", osType)
	}

	// find directory of the running executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	// create bin directory if it doesn't exist
	binDir := filepath.Join(execDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("cannot create bin directory: %w", err)
	}

	// determine archive filename
	archiveName := filepath.Base(downloadUrl)
	archivePath := filepath.Join(binDir, archiveName)

	// download the archive
	if err := downloadFile(downloadUrl, archivePath); err != nil {
		return fmt.Errorf("cannot download ffmpeg: %w", err)
	}

	// create temporary extraction directory
	extractDir := filepath.Join(binDir, "ffmpeg_extract_tmp")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("cannot create temporary extraction directory: %w", err)
	}

	// extract archive based on file type
	if strings.HasSuffix(archiveName, ".zip") {
		if err := unzipFile(archivePath, extractDir); err != nil {
			return fmt.Errorf("cannot unzip ffmpeg: %w", err)
		}
	} else if strings.HasSuffix(archiveName, ".tar.xz") {
		if err := untarXzFile(archivePath, extractDir); err != nil {
			return fmt.Errorf("cannot extract tar.xz ffmpeg: %w", err)
		}
	} else {
		return fmt.Errorf("unknown ffmpeg archive format: %s", archiveName)
	}

	// find ffmpeg binary inside extracted folder
	ffmpegPath := ""
	_ = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if osType == "windows" && info.Name() == "ffmpeg.exe" {
			ffmpegPath = path
			return filepath.SkipDir
		}
		if osType != "windows" && info.Name() == "ffmpeg" {
			ffmpegPath = path
			return filepath.SkipDir
		}
		return nil
	})

	if ffmpegPath == "" {
		return fmt.Errorf("ffmpeg binary not found after extraction")
	}

	// final destination path
	destName := "ffmpeg"
	if osType == "windows" {
		destName += ".exe"
	}
	destPath := filepath.Join(binDir, destName)

	// copy binary to bin directory
	if err := copyFile(ffmpegPath, destPath); err != nil {
		return fmt.Errorf("cannot copy ffmpeg binary: %w", err)
	}

	// make executable on Linux/macOS
	if osType != "windows" {
		if err := os.Chmod(destPath, 0755); err != nil {
			return fmt.Errorf("cannot set executable permissions: %w", err)
		}
	}

	// cleanup archive and temporary extract directory
	os.Remove(archivePath)
	os.RemoveAll(extractDir)

	return nil
}

// DownloadAndExtractDeno downloads the latest Deno for the current OS,
// extracts the archive, places the deno binary into the bin folder,
// and removes the unneeded files.
func DownloadAndExtractDeno() error {

	// get current OS type and architecture
	osType := runtime.GOOS
	arch := runtime.GOARCH

	// determine the correct download URL for the OS and architecture
	var downloadUrl string
	switch osType {
	case "windows":
		downloadUrl = "https://github.com/denoland/deno/releases/latest/download/deno-x86_64-pc-windows-msvc.zip"
	case "linux":
		downloadUrl = "https://github.com/denoland/deno/releases/latest/download/deno-x86_64-unknown-linux-gnu.zip"
	case "darwin":
		if arch == "arm64" {
			downloadUrl = "https://github.com/denoland/deno/releases/latest/download/deno-aarch64-apple-darwin.zip"
		} else {
			downloadUrl = "https://github.com/denoland/deno/releases/latest/download/deno-x86_64-apple-darwin.zip"
		}
	default:
		return fmt.Errorf("unsupported OS: %s", osType)
	}

	// find directory of the running executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	// create bin directory if it doesn't exist
	binDir := filepath.Join(execDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("cannot create bin directory: %w", err)
	}

	// determine archive filename
	archiveName := filepath.Base(downloadUrl)
	archivePath := filepath.Join(binDir, archiveName)

	// download the archive
	if err := downloadFile(downloadUrl, archivePath); err != nil {
		return fmt.Errorf("cannot download deno: %w", err)
	}

	// create temporary extraction directory
	extractDir := filepath.Join(binDir, "deno_extract_tmp")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("cannot create temporary extraction directory: %w", err)
	}

	// extract archive (zip only)
	if strings.HasSuffix(archiveName, ".zip") {
		if err := unzipFile(archivePath, extractDir); err != nil {
			return fmt.Errorf("cannot unzip deno: %w", err)
		}
	} else {
		return fmt.Errorf("unknown deno archive format: %s", archiveName)
	}

	// find deno binary inside extracted folder
	denoPath := ""
	_ = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if osType == "windows" && info.Name() == "deno.exe" {
			denoPath = path
			return filepath.SkipDir
		}
		if osType != "windows" && info.Name() == "deno" {
			denoPath = path
			return filepath.SkipDir
		}
		return nil
	})

	if denoPath == "" {
		return fmt.Errorf("deno binary not found after extraction")
	}

	// final destination path
	destName := "deno"
	if osType == "windows" {
		destName = "deno.exe"
	}
	destPath := filepath.Join(binDir, destName)

	// copy binary to bin directory
	if err := copyFile(denoPath, destPath); err != nil {
		return fmt.Errorf("cannot copy deno binary: %w", err)
	}

	// make executable on Linux/macOS
	if osType != "windows" {
		if err := os.Chmod(destPath, 0755); err != nil {
			return fmt.Errorf("cannot set executable permissions: %w", err)
		}
	}

	// cleanup archive and extract directory
	os.Remove(archivePath)
	os.RemoveAll(extractDir)

	return nil
}

// updateYtDlp removes the existing yt-dlp binary and downloads the latest version.
func updateYtDlp() error {
	// find directory of the running executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	// path to existing yt-dlp binary
	binDir := filepath.Join(execDir, "bin")
	var ytDlpPath string
	if runtime.GOOS == "windows" {
		ytDlpPath = filepath.Join(binDir, "yt-dlp.exe")
	} else {
		ytDlpPath = filepath.Join(binDir, "yt-dlp")
	}

	// remove old binary if it exists
	if _, err := os.Stat(ytDlpPath); err == nil {
		if err := os.Remove(ytDlpPath); err != nil {
			return fmt.Errorf("cannot remove old yt-dlp: %w", err)
		}
	}

	// download the latest version
	if err := downloadYtDlp(); err != nil {
		return fmt.Errorf("cannot download latest yt-dlp: %w", err)
	}

	return nil
}
