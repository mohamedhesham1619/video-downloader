package ui

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
)

func ShowDownloadProgress(message string) *uiprogress.Bar {
	// Define colors
	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	// Create the progress bar
	bar := uiprogress.AddBar(100)
	bar.Width = 50
	bar.Empty = ' '

	// Display the message (first line)
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("%s\nProgress:", message)
	})

	// Display the percentage (after progress bar)
	bar.AppendFunc(func(b *uiprogress.Bar) string {
		percentage := strutil.PadLeft(fmt.Sprintf("%d%%", b.Current()), 4, ' ')
		if b.Current() >= 100 {
			return green(percentage) + " " + green("[DONE]")
		}
		return cyan(percentage)
	})

	// Return the bar so caller can update it
	return bar
}
