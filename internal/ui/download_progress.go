package ui

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
)

// ShowDownloadProgress adds a single-line progress bar to the given progress instance.
// Labels should be printed to stdout before calling this, so uiprogress re-renders
// don't overwrite them.
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
