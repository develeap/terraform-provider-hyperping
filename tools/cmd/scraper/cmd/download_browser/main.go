// download_browser uses Rod's built-in downloader to fetch a compatible Chromium binary.
package main

import (
	"fmt"
	"os"

	"github.com/go-rod/rod/lib/launcher"
)

func main() {
	fmt.Println("Downloading Chromium via Rod browser manager...")
	b := launcher.NewBrowser()
	path, err := b.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Browser available at: %s\n", path)
}
