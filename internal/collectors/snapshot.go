package collectors

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type SnapshotResult struct {
	URL          string
	Path         string
	StatusCode   int
	ETag         string
	LastModified string
	NotModified  bool
}

type SnapshotOptions struct {
	URL          string
	Destination  string
	ETag         string
	LastModified string
	Client       *http.Client
	Now          func() time.Time
}

func FetchSnapshot(ctx context.Context, opts SnapshotOptions) (SnapshotResult, error) {
	if opts.URL == "" {
		return SnapshotResult{}, fmt.Errorf("url is required")
	}
	if opts.Destination == "" {
		return SnapshotResult{}, fmt.Errorf("destination is required")
	}
	client := opts.Client
	if client == nil {
		client = http.DefaultClient
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, opts.URL, nil)
	if err != nil {
		return SnapshotResult{}, err
	}
	if opts.ETag != "" {
		req.Header.Set("If-None-Match", opts.ETag)
	}
	if opts.LastModified != "" {
		req.Header.Set("If-Modified-Since", opts.LastModified)
	}

	resp, err := client.Do(req)
	if err != nil {
		return SnapshotResult{}, err
	}
	defer resp.Body.Close()

	result := SnapshotResult{
		URL:          opts.URL,
		StatusCode:   resp.StatusCode,
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
	}
	if resp.StatusCode == http.StatusNotModified {
		result.NotModified = true
		return result, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return result, fmt.Errorf("fetch %s: status %d", opts.URL, resp.StatusCode)
	}

	dateDir := now().UTC().Format("2006-01-02")
	if err := os.MkdirAll(filepath.Join(opts.Destination, "snapshots", dateDir), 0o755); err != nil {
		return result, err
	}
	name := filepath.Base(req.URL.Path)
	if name == "." || name == "/" || name == "" {
		name = "feed.csv"
	}
	path := filepath.Join(opts.Destination, "snapshots", dateDir, name)
	tmp := path + ".tmp"

	out, err := os.Create(tmp)
	if err != nil {
		return result, err
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		os.Remove(tmp)
		return result, err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return result, err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return result, err
	}
	result.Path = path
	return result, nil
}
