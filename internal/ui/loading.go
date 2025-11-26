package ui

import (
	"io"
	"sync"

	"github.com/pterm/pterm"
)

var (
	multiPrinter *pterm.MultiPrinter
	once         sync.Once
)

// StopMultiPrinter stops the multi-printer
func StopMultiPrinter() {
	if multiPrinter != nil {
		multiPrinter.Stop()
	}
}

// ProgressLine represents a single progress line with spinner
type ProgressLine struct {
	spinner *pterm.SpinnerPrinter
	writer  io.Writer
}

// ShowLoading shows a loading spinner with custom message and returns a handle to complete it
func ShowLoading(message string) *ProgressLine {
	once.Do(func() {
		multi := pterm.DefaultMultiPrinter
		multi.Start()
		multiPrinter = &multi
	})

	customSpinner := pterm.DefaultSpinner

	customSpinner.Sequence = []string{"|", "/", "-", "\\"}
	customSpinner.Style = pterm.NewStyle(pterm.FgCyan)
	customSpinner.RemoveWhenDone = true

	writer := multiPrinter.NewWriter()
	spinner, _ := customSpinner.WithWriter(writer).Start(message)

	return &ProgressLine{
		spinner: spinner,
		writer:  writer,
	}
}

// Complete changes the line to show success with custom message
func (pl *ProgressLine) Complete(message string) {
	pl.spinner.Stop()
	pterm.Fprintln(pl.writer, pterm.Green(message+" [DONE]"))
}

// Fail changes the line to show failure with custom message
func (pl *ProgressLine) Fail(message string) {
	pl.spinner.Stop()
	pterm.Fprintln(pl.writer, pterm.Red(message+" X"))
}
