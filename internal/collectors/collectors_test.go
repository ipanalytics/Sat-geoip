package collectors

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFetchSnapshotWritesDatedAppendOnlyFile(t *testing.T) {
	dir := t.TempDir()
	result, err := FetchSnapshot(context.Background(), SnapshotOptions{
		URL:         "https://example.test/feed.csv",
		Destination: dir,
		Client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"ETag": []string{`"abc"`}},
				Body:       io.NopCloser(strings.NewReader("14.1.64.0/24,PH,,Manila,\n")),
				Request:    r,
			}, nil
		})},
		Now: func() time.Time {
			return time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "snapshots", "2026-05-25", "feed.csv")
	if result.Path != want {
		t.Fatalf("path = %q, want %q", result.Path, want)
	}
	data, err := os.ReadFile(want)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Manila") {
		t.Fatalf("unexpected snapshot content: %q", data)
	}
}

func TestFetchSnapshotHonorsNotModified(t *testing.T) {
	result, err := FetchSnapshot(context.Background(), SnapshotOptions{
		URL:         "https://example.test/feed.csv",
		Destination: t.TempDir(),
		ETag:        `"abc"`,
		Client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Header.Get("If-None-Match") != `"abc"` {
				t.Fatalf("missing If-None-Match")
			}
			return &http.Response{
				StatusCode: http.StatusNotModified,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader("")),
				Request:    r,
			}, nil
		})},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.NotModified || result.Path != "" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestParseRIPEAnnouncedPrefixes(t *testing.T) {
	prefixes, err := ParseRIPEAnnouncedPrefixes(strings.NewReader(`{"data":{"prefixes":[{"prefix":"14.1.64.0/24"}]}}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(prefixes) != 1 || prefixes[0] != "14.1.64.0/24" {
		t.Fatalf("unexpected prefixes: %#v", prefixes)
	}
}
