package downloader

import "sync"

// errorCollector safely collects errors from concurrent downloads
type errorCollector struct {
	mu     sync.Mutex
	errors []string
}

func (ec *errorCollector) Add(err string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.errors = append(ec.errors, err)
}

func (ec *errorCollector) GetAll() []string {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	return ec.errors
}

func (ec *errorCollector) HasErrors() bool {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	return len(ec.errors) > 0
}
