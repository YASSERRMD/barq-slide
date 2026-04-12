// Package images fetches and manages photographic assets for slide rendering.
package images

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	unsplashAPIBase  = "https://api.unsplash.com"
	unsplashSearchEP = "/search/photos"
	// defaultImageW is the width requested from Unsplash for slide rendering.
	// 1920px covers both the HTML render (1200px canvas) and PPTX (300dpi).
	defaultImageW = 1920
	httpTimeout   = 15 * time.Second
)

// UnsplashPhoto represents a resolved photo from Unsplash.
type UnsplashPhoto struct {
	// ID is the Unsplash photo ID.
	ID string

	// URL is the download URL at the requested width.
	URL string

	// Attribution is the photographer credit required by the Unsplash API Terms.
	Attribution Attribution

	// Width and Height are the original image dimensions.
	Width  int
	Height int

	// AltDescription is a human-readable description.
	AltDescription string

	// BlurHash is the blur placeholder string.
	BlurHash string
}

// Attribution holds photographer credit information.
type Attribution struct {
	PhotographerName string
	PhotographerURL  string
	UTMSource        string
}

// String returns a formatted attribution string for slide captions.
func (a Attribution) String() string {
	return fmt.Sprintf("Photo by %s on Unsplash", a.PhotographerName)
}

// unsplashSearchResponse mirrors the Unsplash search API JSON schema.
type unsplashSearchResponse struct {
	Total   int `json:"total"`
	Results []struct {
		ID   string `json:"id"`
		Urls struct {
			Raw     string `json:"raw"`
			Full    string `json:"full"`
			Regular string `json:"regular"`
		} `json:"urls"`
		User struct {
			Name  string `json:"name"`
			Links struct {
				HTML string `json:"html"`
			} `json:"links"`
		} `json:"user"`
		Width          int    `json:"width"`
		Height         int    `json:"height"`
		AltDescription string `json:"alt_description"`
		BlurHash       string `json:"blur_hash"`
	} `json:"results"`
}

// UnsplashAdapter fetches photos from the Unsplash API.
type UnsplashAdapter struct {
	accessKey  string
	httpClient *http.Client
	utmSource  string
}

// NewUnsplashAdapter creates a new UnsplashAdapter.
func NewUnsplashAdapter(accessKey string) *UnsplashAdapter {
	return &UnsplashAdapter{
		accessKey:  accessKey,
		httpClient: &http.Client{Timeout: httpTimeout},
		utmSource:  "barq-slides",
	}
}

// Search finds the best photo for a semantic query.
// Returns the top result or an error if no results are found.
func (a *UnsplashAdapter) Search(ctx context.Context, query string) (*UnsplashPhoto, error) {
	if a.accessKey == "" {
		return nil, fmt.Errorf("unsplash: access key not configured")
	}

	q := url.Values{}
	q.Set("query", query)
	q.Set("per_page", "3")
	q.Set("orientation", "landscape") // Slides are landscape.

	endpoint := fmt.Sprintf("%s%s?%s", unsplashAPIBase, unsplashSearchEP, q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("unsplash: building request: %w", err)
	}

	req.Header.Set("Authorization", "Client-ID "+a.accessKey)
	req.Header.Set("Accept-Version", "v1")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unsplash: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("unsplash: API error %d: %s", resp.StatusCode, string(body))
	}

	var result unsplashSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("unsplash: decoding response: %w", err)
	}

	if result.Total == 0 || len(result.Results) == 0 {
		return nil, fmt.Errorf("unsplash: no results for query %q", query)
	}

	photo := result.Results[0]

	// Build the CDN URL with the requested width.
	imageURL := buildUnsplashURL(photo.Urls.Raw, defaultImageW)

	return &UnsplashPhoto{
		ID:  photo.ID,
		URL: imageURL,
		Attribution: Attribution{
			PhotographerName: photo.User.Name,
			PhotographerURL:  appendUTM(photo.User.Links.HTML, a.utmSource),
			UTMSource:        a.utmSource,
		},
		Width:          photo.Width,
		Height:         photo.Height,
		AltDescription: photo.AltDescription,
		BlurHash:       photo.BlurHash,
	}, nil
}

// Download fetches the image bytes from a resolved UnsplashPhoto URL.
func (a *UnsplashAdapter) Download(ctx context.Context, photo *UnsplashPhoto) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, photo.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("unsplash download: building request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unsplash download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unsplash download: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unsplash download: reading body: %w", err)
	}

	return data, nil
}

// ── helpers ────────────────────────────────────────────────────────────────────

func buildUnsplashURL(rawURL string, width int) string {
	if !strings.Contains(rawURL, "?") {
		return fmt.Sprintf("%s?w=%d&auto=format&fit=crop&q=80", rawURL, width)
	}
	return fmt.Sprintf("%s&w=%d&auto=format&fit=crop&q=80", rawURL, width)
}

func appendUTM(u, source string) string {
	if u == "" {
		return u
	}
	sep := "?"
	if strings.Contains(u, "?") {
		sep = "&"
	}
	return u + sep + "utm_source=" + source + "&utm_medium=referral"
}
