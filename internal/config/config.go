package config

import (
	"downloader/internal/models"
	"downloader/internal/utils"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jaypipes/ghw"
)

type Config struct {

	// the path to the download directory (the default is the directory where the program is executed)
	DownloadPath string

	// the video format to download
	VideoFormat models.VideoFormat

	// if true, the downloader will re-encode clips using the encoder specified in the config
	ShouldReEncode bool

	// the encoder to use for re-encoding if ShouldUseEncoder is true
	Encoder string
}

func New(shouldReEncode bool, videoFormat models.VideoFormat) *Config {

	downloadPathFlag := flag.String("path", "", "path to the download directory (the default is the current directory)")

	flag.Parse()

	// if the user provides a path flag, the downloaded videos will be saved in that directory. Otherwise, they will be saved in the "Downloads" folder in the current folder.
	downloadPath := *downloadPathFlag

	if downloadPath == "" {
		err := os.MkdirAll("Downloads", os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating Downloads folder: %v Will use the current folder instead.\n\n", err)
			downloadPath = ""
		} else {
			downloadPath = "Downloads"
		}

	}

	// If shouldReEncode is true, select the encoder to use based on the GPU.
	// If the GPU is not detected or the GPU encoder is not working, the CPU encoder will be used.
	encoder := ""

	if shouldReEncode {
		encoder = selectEncoder()
	}

	// create the config
	cfg := &Config{
		DownloadPath:   downloadPath,
		VideoFormat:    videoFormat,
		Encoder:        encoder,
		ShouldReEncode: shouldReEncode,
	}

	return cfg
}

const (
	// GPU vendors
	NvidiaGPU = "nvidia"
	AMDGPU    = "amd"
	IntelGPU  = "intel"

	// Default CPU encoder
	CPUEncoder = "libx264"
)

// GPUEncoders maps GPU names to their corresponding encoder names
var GPUEncoders = map[string]string{
	NvidiaGPU: "h264_nvenc",
	AMDGPU:    "h264_amf",
	IntelGPU:  "h264_qsv",
}

// Detect the GPU vendor name
func detectGpu() (string, error) {

	gpuInfo, err := ghw.GPU()

	if err != nil {
		return "", fmt.Errorf("error getting GPU info: %v", err)
	}

	gpu := ""
	gpuVendorName := strings.ToLower(gpuInfo.GraphicsCards[0].DeviceInfo.Vendor.Name)

	// Check the GPU vendor name and set the GPU variable accordingly
	switch {

	case strings.Contains(gpuVendorName, "nvidia"):
		gpu = NvidiaGPU
	case strings.Contains(gpuVendorName, "amd") || strings.Contains(gpuVendorName, "advanced micro devices"):
		gpu = AMDGPU
	case strings.Contains(gpuVendorName, "intel"):
		gpu = IntelGPU
	}

	return gpu, nil

}

// Test if the GPU encoder is working
// If the command runs successfully and doesn't return any error, the encoder is working
func isGpuEncoderWorking(encoder string) bool {
	testCmd := exec.Command(
		utils.GetBinaryPath("ffmpeg"),
		"-hide_banner",
		"-loglevel", "error",
		"-f", "lavfi",
		"-i", "testsrc=duration=1",
		"-c:v", encoder,
		"-frames:v", "10",
		"-f", "null",
		"-",
	)
	return testCmd.Run() == nil
}

// The encoder will be selected based on the GPU. If the GPU is not detected or the GPU encoder is not working, the CPU encoder will be used.
func selectEncoder() string {
	gpu, err := detectGpu()
	if err != nil || gpu == "" || !isGpuEncoderWorking(GPUEncoders[gpu]) {
		return CPUEncoder
	}

	return GPUEncoders[gpu]
}
