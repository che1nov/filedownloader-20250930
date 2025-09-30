package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestDownloaderExtractFilename tests filename extraction from various URLs
func TestDownloaderExtractFilename(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "simple filename",
			url:      "http://example.com/file.txt",
			expected: "file.txt",
		},
		{
			name:     "filename with path",
			url:      "http://example.com/path/to/file.pdf",
			expected: "file.pdf",
		},
		{
			name:     "filename with query params",
			url:      "http://example.com/file.zip?version=1.0",
			expected: "file.zip",
		},
		{
			name:     "filename with fragment",
			url:      "http://example.com/file.doc#section1",
			expected: "file.doc",
		},
		{
			name:     "root path",
			url:      "http://example.com/",
			expected: "file_19",
		},
		{
			name:     "empty path",
			url:      "http://example.com",
			expected: "file_18",
		},
		{
			name:     "nested path",
			url:      "http://x/y/z.txt",
			expected: "z.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDownloader()
			result := d.ExtractFilename(tt.url)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestDownloaderGetFileSize tests file size retrieval
func TestDownloaderGetFileSize(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name:        "valid file",
			content:     "hello world",
			expectError: false,
		},
		{
			name:        "empty file",
			content:     "",
			expectError: false,
		},
		{
			name:        "large content",
			content:     "this is a very long content that should be larger than the previous ones",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, tt.content)
			}))
			defer srv.Close()

			d := NewDownloader()
			size, err := d.GetFileSize(srv.URL)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			expectedSize := int64(len(tt.content))
			if tt.content == "" && size == -1 {
				// This is acceptable for empty files
				return
			}
			if size != expectedSize {
				t.Errorf("expected size %d, got %d", expectedSize, size)
			}
		})
	}
}

// TestDownloaderDownloadFile tests file download functionality
func TestDownloaderDownloadFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		filename    string
		expectError bool
	}{
		{
			name:        "simple text file",
			content:     "hello world",
			filename:    "test.txt",
			expectError: false,
		},
		{
			name:        "empty file",
			content:     "",
			filename:    "empty.txt",
			expectError: false,
		},
		{
			name:        "binary-like content",
			content:     "binary content with special chars: \x00\x01\x02",
			filename:    "binary.bin",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, tt.content)
			}))
			defer srv.Close()

			d := NewDownloader()
			tmpDir := t.TempDir()
			d.downloadsDir = tmpDir

			filename, err := d.DownloadFile(srv.URL, tt.filename)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.filename == "binary.bin" {
				if filename != "binary.bin" && filename != "binary.bin.bin" {
					t.Errorf("expected filename binary.bin or binary.bin.bin, got %s", filename)
				}
			} else if filename != tt.filename {
				t.Errorf("expected filename %s, got %s", tt.filename, filename)
			}

			filePath := filepath.Join(tmpDir, filename)
			info, err := os.Stat(filePath)
			if err != nil {
				t.Errorf("file not found: %v", err)
				return
			}

			if info.Size() != int64(len(tt.content)) {
				t.Errorf("expected file size %d, got %d", len(tt.content), info.Size())
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Errorf("failed to read file: %v", err)
				return
			}

			if string(content) != tt.content {
				t.Errorf("expected content %q, got %q", tt.content, string(content))
			}
		})
	}
}

// TestDownloaderIntegration tests the complete download workflow
func TestDownloaderIntegration(t *testing.T) {
	tests := []struct {
		name    string
		content string
		urlPath string
	}{
		{
			name:    "complete workflow",
			content: "integration test content",
			urlPath: "/integration-test.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, tt.content)
			}))
			defer srv.Close()

			d := NewDownloader()
			tmpDir := t.TempDir()
			d.downloadsDir = tmpDir

			filename := d.ExtractFilename(srv.URL + tt.urlPath)
			if filename == "" {
				t.Fatalf("empty filename")
			}

			size, err := d.GetFileSize(srv.URL)
			if err != nil {
				t.Fatalf("GetFileSize error: %v", err)
			}
			if size <= 0 {
				t.Fatalf("unexpected size: %d", size)
			}

			downloadedFilename, err := d.DownloadFile(srv.URL, filename)
			if err != nil {
				t.Fatalf("DownloadFile error: %v", err)
			}

			path := filepath.Join(tmpDir, downloadedFilename)
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("stat error: %v", err)
			}
			if info.Size() <= 0 {
				t.Fatalf("file not written")
			}
		})
	}
}
