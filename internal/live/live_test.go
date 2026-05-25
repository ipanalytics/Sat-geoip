package live

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestEvidenceBuildsFromFeedsAndBGP(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body := ""
		switch {
		case strings.Contains(r.URL.String(), "geoip.starlinkisp.net/feed.csv"):
			body = "14.1.64.0/24,PH,,Manila,\n"
		case strings.Contains(r.URL.String(), "geoip.starlinkisp.net/pops.csv"):
			body = "14.1.64.0/24,mnlaphl1,mnl\n"
		case strings.Contains(r.URL.String(), "Viasat/geofeed"):
			body = "# prefix/ip,country_code,region_code,city,postal\n184.63.131.8/29,LK,,Riyadh,\n"
		case strings.Contains(r.URL.String(), "resource=AS45700"):
			body = `{"data":{"prefixes":[{"prefix":"14.1.64.0/24"}]}}`
		case strings.Contains(r.URL.String(), "resource=AS7155"):
			body = `{"data":{"prefixes":[{"prefix":"184.63.131.8/29"}]}}`
		case strings.Contains(r.URL.String(), "resource=AS60725"):
			body = `{"data":{"prefixes":[{"prefix":"203.0.113.0/24"}]}}`
		case strings.Contains(r.URL.String(), "resource=AS800"):
			body = `{"data":{"prefixes":[{"prefix":"23.160.32.0/24"}]}}`
		case strings.Contains(r.URL.String(), "resource=AS22351"):
			body = `{"data":{"prefixes":[{"prefix":"209.159.160.0/19"}]}}`
		case strings.Contains(r.URL.String(), "resource=AS29286"):
			body = `{"data":{"prefixes":[{"prefix":"176.227.128.0/20"}]}}`
		default:
			body = `{"data":{"prefixes":[]}}`
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}, nil
	})}

	evidence, err := Evidence(context.Background(), Options{Client: client})
	if err != nil {
		t.Fatal(err)
	}
	if len(evidence) != 6 {
		t.Fatalf("expected 6 evidence records, got %d", len(evidence))
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
