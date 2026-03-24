//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	baseURL := "http://localhost:8080"
	if len(os.Args) > 1 {
		baseURL = os.Args[1]
	}

	screenshots := []struct {
		name string
		url  string
	}{
		{"screenshot-dashboard", "/dashboard/"},
		{"screenshot-games", "/games"},
		{"screenshot-form-validation", "/games/new"},
		{"screenshot-data", "/data"},
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(1920, 1080),
		chromedp.Flag("force-dark-mode", true),
		chromedp.ExecPath("/usr/bin/chromium-browser"),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	os.MkdirAll("docs", 0o755)

	for _, s := range screenshots {
		var buf []byte
		url := baseURL + s.url
		fmt.Printf("Capturing %s ...\n", s.name)

		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.Sleep(3*time.Second),
			chromedp.FullScreenshot(&buf, 90),
		)
		if err != nil {
			log.Printf("Error capturing %s: %v", s.name, err)
			continue
		}

		path := fmt.Sprintf("docs/%s.png", s.name)
		if err := os.WriteFile(path, buf, 0o644); err != nil {
			log.Printf("Error writing %s: %v", path, err)
			continue
		}
		fmt.Printf("  -> %s (%d KB)\n", path, len(buf)/1024)
	}

	// Overview: need to click the "Übersicht" button first
	fmt.Println("Capturing screenshot-overview ...")
	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL+"/dashboard/"),
		chromedp.Sleep(3*time.Second),
		chromedp.Click(`button[class*="btn"]`, chromedp.ByQuery, chromedp.NodeVisible),
		chromedp.Sleep(500*time.Millisecond),
		// Click the "Übersicht" button (second in the btn-group)
		chromedp.Evaluate(`document.querySelectorAll('.btn-group button')[1].click()`, nil),
		chromedp.Sleep(3*time.Second),
		chromedp.FullScreenshot(&buf, 90),
	)
	if err != nil {
		log.Printf("Error capturing overview: %v", err)
	} else {
		path := "docs/screenshot-overview.png"
		os.WriteFile(path, buf, 0o644)
		fmt.Printf("  -> %s (%d KB)\n", path, len(buf)/1024)
	}

	fmt.Println("Done!")
}
