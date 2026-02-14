// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package interactive

import (
	"fmt"
	"io"
	"time"

	"github.com/briandowns/spinner"
	"github.com/schollz/progressbar/v3"
)

// Spinner wraps spinner functionality.
type Spinner struct {
	spinner *spinner.Spinner
	writer  io.Writer
}

// NewSpinner creates a new spinner.
func NewSpinner(message string, writer io.Writer) *Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(writer))
	s.Suffix = " " + message
	s.FinalMSG = ""
	return &Spinner{
		spinner: s,
		writer:  writer,
	}
}

// Start starts the spinner.
func (s *Spinner) Start() {
	s.spinner.Start()
}

// Stop stops the spinner.
func (s *Spinner) Stop() {
	s.spinner.Stop()
}

// UpdateMessage updates the spinner message.
func (s *Spinner) UpdateMessage(message string) {
	s.spinner.Suffix = " " + message
}

// SuccessMessage stops the spinner with a success message.
func (s *Spinner) SuccessMessage(message string) {
	s.spinner.FinalMSG = fmt.Sprintf("✅ %s\n", message)
	s.spinner.Stop()
}

// ErrorMessage stops the spinner with an error message.
func (s *Spinner) ErrorMessage(message string) {
	s.spinner.FinalMSG = fmt.Sprintf("❌ %s\n", message)
	s.spinner.Stop()
}

// ProgressBar wraps progress bar functionality.
type ProgressBar struct {
	bar *progressbar.ProgressBar
}

// NewProgressBar creates a new progress bar.
func NewProgressBar(maxValue int64, description string, writer io.Writer) *ProgressBar {
	bar := progressbar.NewOptions64(
		maxValue,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(writer),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(writer, "\n")
		}),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	return &ProgressBar{bar: bar}
}

// Add increments the progress bar.
func (pb *ProgressBar) Add(n int) error {
	return pb.bar.Add(n)
}

// Set sets the progress bar to a specific value.
func (pb *ProgressBar) Set(n int) error {
	return pb.bar.Set(n)
}

// Finish completes the progress bar.
func (pb *ProgressBar) Finish() error {
	return pb.bar.Finish()
}

// Clear clears the progress bar.
func (pb *ProgressBar) Clear() error {
	return pb.bar.Clear()
}
