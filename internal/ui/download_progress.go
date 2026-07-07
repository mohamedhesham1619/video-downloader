package ui

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
)

// ShowDownloadProgress adds a single-line progress bar to the given progress instance.
func ShowDownloadProgress(progress *uiprogress.Progress) *uiprogress.Bar {
	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	bar := progress.AddBar(100)
	bar.Width = 50
	bar.Empty = ' '

	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return "Progress:"
	})

	bar.AppendFunc(func(b *uiprogress.Bar) string {
		percentage := strutil.PadLeft(fmt.Sprintf("%d%%", b.Current()), 4, ' ')
		if b.Current() >= 100 {
			return green(percentage) + " " + green("[DONE]")
		}
		return cyan(percentage)
	})

	return bar
}

// AddLabelBar creates a fake progress bar to safely display text above actual progress bars.
func AddLabelBar(progress *uiprogress.Progress, text string) {
	bar := progress.AddBar(1)
	bar.Width = 1 // Must be > 0 to prevent index out of range panic in uiprogress
	bar.AppendFunc(func(b *uiprogress.Bar) string {
		// \r moves to start to overwrite the dummy bar, \033[K clears the rest of the line
		return fmt.Sprintf("\r%s\033[K", text)
	})
}
