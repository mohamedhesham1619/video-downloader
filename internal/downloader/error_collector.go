package downloader

import (
	"downloader/internal/utils"
	"strings"
	"sync"
)

type downloadError struct {
	URL     string
	Message string
}

// errorCollector safely collects errors from concurrent downloads
type errorCollector struct {
	mu     sync.Mutex
	errors []downloadError
}

func (ec *errorCollector) Add(url, message string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.errors = append(ec.errors, downloadError{URL: url, Message: message})
}

func (ec *errorCollector) RemoveByURL(url string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	filtered := ec.errors[:0]
	for _, e := range ec.errors {
		if e.URL != url {
			filtered = append(filtered, e)
		}
	}
	ec.errors = filtered
}

func (ec *errorCollector) GetAll() []downloadError {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	return ec.errors
}

// GetSignInErrors returns errors for YouTube URLs that failed due to a sign-in requirement.
func (ec *errorCollector) GetSignInErrors() []downloadError {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	var result []downloadError
	for _, e := range ec.errors {
		if utils.IsYouTubeURL(e.URL) && strings.Contains(strings.ToLower(e.Message), "sign in") {
			result = append(result, e)
		}
	}
	return result
}

func (ec *errorCollector) HasErrors() bool {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	return len(ec.errors) > 0
}
