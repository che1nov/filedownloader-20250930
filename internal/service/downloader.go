package service

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Downloader struct {
	downloadsDir string
	timeout      time.Duration
	maxFileSize  int64
	userAgent    string
}

// NewDownloader creates a new downloader instance
func NewDownloader() *Downloader {
	return &Downloader{
		downloadsDir: "downloads",
		timeout:      60 * time.Second,
		maxFileSize:  100 * 1024 * 1024, // 100MB
		userAgent:    "FileDownloader/1.0",
	}
}

// DownloadFile downloads a file from URL and saves it to local directory
func (d *Downloader) DownloadFile(url, filename string) (string, error) {
	client := &http.Client{
		Timeout: d.timeout,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request for %s: %w", url, err)
	}

	req.Header.Set("User-Agent", d.userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status code %d for %s", resp.StatusCode, url)
	}

	if d.maxFileSize > 0 && resp.ContentLength > d.maxFileSize {
		return "", fmt.Errorf("file size %d exceeds limit %d", resp.ContentLength, d.maxFileSize)
	}

	finalName := filename
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if n := parseFilenameFromContentDisposition(cd); n != "" {
			finalName = n
		}
	}
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		hasValidExtension := false
		if strings.Contains(finalName, ".") {
			ext := strings.ToLower(filepath.Ext(finalName))
			hasValidExtension = ext == ".html" || ext == ".htm" || ext == ".txt" || ext == ".pdf" || ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".css" || ext == ".js"
		}

		if !hasValidExtension {
			if exts, _ := mime.ExtensionsByType(ct); len(exts) > 0 {
				finalName = finalName + exts[0]
			} else if strings.Contains(strings.ToLower(ct), "text/html") {
				finalName = finalName + ".html"
			}
		}
	}

	if err := os.MkdirAll(d.downloadsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create downloads dir: %w", err)
	}
	filePath := filepath.Join(d.downloadsDir, finalName)
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return finalName, nil
}

// ExtractFilename extracts filename from URL
func (d *Downloader) ExtractFilename(u string) string {
	parsed, err := neturl.Parse(u)
	if err != nil {
		return fmt.Sprintf("file_%d", len(u))
	}
	p := parsed.Path
	if p == "" || p == "/" {
		return fmt.Sprintf("file_%d", len(u))
	}
	segs := strings.Split(p, "/")
	name := segs[len(segs)-1]
	if name == "" {
		return fmt.Sprintf("file_%d", len(u))
	}
	return name
}

// GetFileSize returns file size by URL using HEAD request
func (d *Downloader) GetFileSize(url string) (int64, error) {
	client := &http.Client{
		Timeout: d.timeout,
	}

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create HEAD request for %s: %w", url, err)
	}

	req.Header.Set("User-Agent", d.userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bad status code %d", resp.StatusCode)
	}

	return resp.ContentLength, nil
}

// parseFilenameFromContentDisposition extracts filename from Content-Disposition header
func parseFilenameFromContentDisposition(cd string) string {
	cd = strings.TrimSpace(cd)
	parts := strings.Split(cd, ";")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(strings.ToLower(p), "filename*=") {
			v := strings.TrimPrefix(p, "filename*=")
			idx := strings.Index(v, "''")
			if idx >= 0 && idx+2 < len(v) {
				return v[idx+2:]
			}
		}
		if strings.HasPrefix(strings.ToLower(p), "filename=") {
			v := strings.TrimPrefix(p, "filename=")
			v = strings.Trim(v, "\"")
			if v != "" {
				return v
			}
		}
	}
	return ""
}
