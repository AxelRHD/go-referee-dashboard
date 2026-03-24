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

// JS to redact sensitive financial data before screenshot
const redactJS = `
(function() {
    var colors = ['#5E81AC','#A3BE8C','#D08770','#88C0D0','#B48EAD','#EBCB8B','#BF616A','#81A1C1'];
    var ci = 0;
    function nextColor() { return colors[ci++ % colors.length]; }

    function redactEl(el) {
        var rect = el.getBoundingClientRect();
        if (rect.width < 5 || rect.height < 5) return;
        var overlay = document.createElement('div');
        overlay.style.cssText = 'position:absolute;left:'+rect.left+'px;top:'+rect.top+'px;width:'+rect.width+'px;height:'+rect.height+'px;background:'+nextColor()+';border-radius:3px;z-index:9999;';
        document.body.appendChild(overlay);
    }

    // Redact EUR amounts and financial numbers in table cells
    document.querySelectorAll('td').forEach(el => {
        if (/[\d.,]+\s*€/.test(el.textContent)) redactEl(el);
    });

    // Redact stat card values (fw-bold with numbers or EUR)
    document.querySelectorAll('.fw-bold').forEach(el => {
        var text = el.textContent.trim();
        if (/[\d.,]+\s*€/.test(text) || /^\d[\d.,]*$/.test(text) || /[\d,]+\s*ct/.test(text)) {
            redactEl(el);
        }
    });

    // Redact Vergütung chart containers (fee charts show EUR)
    document.querySelectorAll('[id^="chart-fee"]').forEach(el => {
        el.style.filter = 'blur(8px)';
    });
})();
`

func main() {
	baseURL := "http://localhost:8080"
	if len(os.Args) > 1 {
		baseURL = os.Args[1]
	}

	screenshots := []struct {
		name string
		url  string
	}{
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
			chromedp.Evaluate(redactJS, nil),
			chromedp.Sleep(500*time.Millisecond),
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

	// Dashboard: switch to 2025 season for a fuller screenshot
	fmt.Println("Capturing screenshot-dashboard ...")
	var dashBuf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL+"/dashboard/"),
		chromedp.Sleep(3*time.Second),
		chromedp.Evaluate(`Alpine.evaluate(document.querySelector('[x-data="dashboard"]'), 'season = "2025"')`, nil),
		chromedp.Sleep(3*time.Second),
		chromedp.Evaluate(redactJS, nil),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.FullScreenshot(&dashBuf, 90),
	)
	if err != nil {
		log.Printf("Error capturing dashboard: %v", err)
	} else {
		path := "docs/screenshot-dashboard.png"
		os.WriteFile(path, dashBuf, 0o644)
		fmt.Printf("  -> %s (%d KB)\n", path, len(dashBuf)/1024)
	}

	// Overview: click "Übersicht" button
	fmt.Println("Capturing screenshot-overview ...")
	var buf []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate(baseURL+"/dashboard/"),
		chromedp.Sleep(3*time.Second),
		chromedp.Evaluate(`document.querySelectorAll('.btn-group button')[1].click()`, nil),
		chromedp.Sleep(3*time.Second),
		chromedp.Evaluate(redactJS, nil),
		chromedp.Sleep(500*time.Millisecond),
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
