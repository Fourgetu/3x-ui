package panel

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
)

type releaseCheckTransport func(*http.Request) (*http.Response, error)

func (f releaseCheckTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestLatestPanelVersionFallback(t *testing.T) {
	for _, tc := range []struct {
		name       string
		apiStatus  int
		apiBody    string
		apiErr     error
		pageStatus int
		location   string
		want       string
		wantError  bool
		requests   int
	}{
		{name: "API stable", apiStatus: 200, apiBody: `{"tag_name":"v3.11.0-fourgetu.3"}`, want: "v3.11.0-fourgetu.3", requests: 1},
		{name: "rate limited", apiStatus: 403, pageStatus: 302, location: "https://github.com/Fourgetu/3x-ui/releases/tag/v3.11.0-fourgetu.3", want: "v3.11.0-fourgetu.3", requests: 2},
		{name: "429 relative redirect", apiStatus: 429, pageStatus: 302, location: "/Fourgetu/3x-ui/releases/tag/v3.11.0-fourgetu.3", want: "v3.11.0-fourgetu.3", requests: 2},
		{name: "API unavailable", apiStatus: 503, pageStatus: 302, location: "/Fourgetu/3x-ui/releases/tag/v3.11.0-fourgetu.3", want: "v3.11.0-fourgetu.3", requests: 2},
		{name: "invalid JSON", apiStatus: 200, apiBody: "<html>error</html>", pageStatus: 302, location: "/Fourgetu/3x-ui/releases/tag/v3.11.0-fourgetu.3", want: "v3.11.0-fourgetu.3", requests: 2},
		{name: "missing API tag", apiStatus: 200, apiBody: "{}", pageStatus: 302, location: "/Fourgetu/3x-ui/releases/tag/v3.11.0-fourgetu.3", want: "v3.11.0-fourgetu.3", requests: 2},
		{name: "DNS failure", apiErr: &net.DNSError{Err: "not found", Name: "api.github.com"}, pageStatus: 302, location: "/Fourgetu/3x-ui/releases/tag/v3.11.0-fourgetu.3", want: "v3.11.0-fourgetu.3", requests: 2},
		{name: "other repository", apiStatus: 403, pageStatus: 302, location: "/MHSanaei/3x-ui/releases/tag/v9.0.0", wantError: true, requests: 2},
		{name: "other host", apiStatus: 403, pageStatus: 302, location: "https://example.com/Fourgetu/3x-ui/releases/tag/v9.0.0", wantError: true, requests: 2},
		{name: "dev tag", apiStatus: 403, pageStatus: 302, location: "/Fourgetu/3x-ui/releases/tag/dev-latest", wantError: true, requests: 2},
		{name: "pre-release", apiStatus: 403, pageStatus: 302, location: "/Fourgetu/3x-ui/releases/tag/v3.12.0-rc.1", wantError: true, requests: 2},
		{name: "both fail", apiStatus: 403, pageStatus: 503, wantError: true, requests: 2},
		{name: "no redirect", apiStatus: 403, pageStatus: 200, wantError: true, requests: 2},
		{name: "missing location", apiStatus: 403, pageStatus: 302, wantError: true, requests: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			client := &http.Client{Transport: releaseCheckTransport(func(req *http.Request) (*http.Response, error) {
				calls++
				status, body := tc.apiStatus, tc.apiBody
				headers := http.Header{}
				switch req.URL.String() {
				case "https://api.github.com/repos/Fourgetu/3x-ui/releases/latest":
					if tc.apiErr != nil {
						return nil, tc.apiErr
					}
				case "https://github.com/Fourgetu/3x-ui/releases/latest":
					status, body = tc.pageStatus, ""
					headers.Set("Location", tc.location)
				default:
					t.Fatalf("unexpected network request: %s", req.URL)
				}
				return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
			})}
			got, err := fetchLatestPanelVersionWithClient(client)
			if (err != nil) != tc.wantError || got != tc.want || calls != tc.requests {
				t.Fatalf("got version=%q err=%v calls=%d; want version=%q error=%v calls=%d", got, err, calls, tc.want, tc.wantError, tc.requests)
			}
		})
	}
}

func TestReleaseCheckDoesNotExposeProxyCredentials(t *testing.T) {
	client := &http.Client{Transport: releaseCheckTransport(func(_ *http.Request) (*http.Response, error) {
		return nil, errors.New("proxy socks5://user:secret-password@private-host:1080 unavailable")
	})}
	_, err := fetchLatestPanelVersionWithClient(client)
	if err == nil || strings.Contains(err.Error(), "secret-password") || strings.Contains(err.Error(), "private-host") {
		t.Fatalf("unexpected public error: %v", err)
	}
}
