// Package render provides headless browser rendering of HTML slides to PNG
// snapshots using a chromedp tab pool for concurrency efficiency.
package render

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	defaultPoolSize    = 4
	renderTimeout      = 30 * time.Second
	viewportWidth      = 1200
	viewportHeight     = 675
	deviceScaleFactor  = 2.0 // Retina-quality screenshots.
)

// TabPool manages a pool of chromedp browser contexts for parallel rendering.
type TabPool struct {
	allocCtx context.Context
	cancel   context.CancelFunc
	tabs     chan context.Context
	once     sync.Once
}

// NewTabPool creates a chromedp browser with a shared allocator and
// a pool of tab contexts ready for rendering.
func NewTabPool(size int) (*TabPool, error) {
	if size <= 0 {
		size = defaultPoolSize
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("font-render-hinting", "none"),
		chromedp.WindowSize(viewportWidth, viewportHeight),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	pool := &TabPool{
		allocCtx: allocCtx,
		cancel:   allocCancel,
		tabs:     make(chan context.Context, size),
	}

	// Pre-warm the pool.
	for i := 0; i < size; i++ {
		tabCtx, _ := chromedp.NewContext(allocCtx)
		// Navigate to a blank page to initialise the tab.
		if err := chromedp.Run(tabCtx, chromedp.Navigate("about:blank")); err != nil {
			allocCancel()
			return nil, fmt.Errorf("render: pre-warming tab %d: %w", i, err)
		}
		pool.tabs <- tabCtx
	}

	return pool, nil
}

// RenderHTML renders an HTML string to a PNG screenshot.
// Acquires a tab from the pool, renders, and returns the tab.
func (p *TabPool) RenderHTML(ctx context.Context, html string) ([]byte, error) {
	// Acquire a tab with a timeout.
	var tabCtx context.Context
	select {
	case tabCtx = <-p.tabs:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(renderTimeout):
		return nil, fmt.Errorf("render: timeout acquiring tab from pool")
	}

	// Return the tab to the pool when done.
	defer func() { p.tabs <- tabCtx }()

	renderCtx, cancel := context.WithTimeout(tabCtx, renderTimeout)
	defer cancel()

	var pngBytes []byte

	tasks := chromedp.Tasks{
		// Set data URI for the HTML content.
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.SetContent(html).Do(ctx)
		}),
		// Set viewport to slide dimensions.
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.EmulateViewport(
				viewportWidth, viewportHeight,
				chromedp.EmulateScale(deviceScaleFactor),
			).Do(ctx)
		}),
		// Wait for fonts and images to load.
		chromedp.Sleep(200 * time.Millisecond),
		// Capture screenshot of the .slide-canvas element.
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Screenshot(".slide-canvas", &pngBytes,
				chromedp.NodeVisible,
				chromedp.ByQuery,
			).Do(ctx)
		}),
	}

	if err := chromedp.Run(renderCtx, tasks...); err != nil {
		return nil, fmt.Errorf("render: chromedp screenshot: %w", err)
	}

	return pngBytes, nil
}

// Close shuts down the browser and all tabs.
func (p *TabPool) Close() {
	p.once.Do(func() {
		// Drain the pool.
		done := make(chan struct{})
		go func() {
			for i := 0; i < cap(p.tabs); i++ {
				<-p.tabs
			}
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}

		p.cancel()
	})
}

// IsChromiumAvailable checks if chromium/chrome is available for headless rendering.
func IsChromiumAvailable() bool {
	ctx, cancel := chromedp.NewExecAllocator(context.Background(),
		chromedp.DefaultExecAllocatorOptions[:]...,
	)
	defer cancel()

	tabCtx, tabCancel := chromedp.NewContext(ctx)
	defer tabCancel()

	err := chromedp.Run(tabCtx, chromedp.Navigate("about:blank"))
	return err == nil
}
