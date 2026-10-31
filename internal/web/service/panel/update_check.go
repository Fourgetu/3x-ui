package panel

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
)

var stablePanelTag = regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+(-fourgetu\.[0-9]+)?$`)

// Only published stable release metadata may choose an update. Never fall back
// to main's version file: it can name a build that has not passed release checks.
func fetchLatestPanelVersionWithClient(client *http.Client) (string, error) {
	release, apiErr := fetchPanelReleaseWithClient(client, "")
	if apiErr == nil {
		if stablePanelTag.MatchString(release.TagName) {
			return release.TagName, nil
		}
		apiErr = fmt.Errorf("GitHub API returned an invalid stable release tag")
	}
	tag, err := fetchLatestPanelRedirect(client)
	if err != nil {
		return "", fmt.Errorf("%v; %w", apiErr, err)
	}
	return tag, nil
}

// The release-page redirect uses github.com rather than the rate-limited API.
// Reuse the panel's transport/proxy and TLS verification, and stop at the
// redirect so an unexpected target is never contacted.
func fetchLatestPanelRedirect(client *http.Client) (string, error) {
	redirectClient := *client
	redirectClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		"https://github.com/"+panelRepository+"/releases/latest", nil)
	if err != nil {
		return "", fmt.Errorf("invalid release page URL")
	}
	resp, err := redirectClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("GitHub release page: %s", releaseNetworkError(err))
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
	default:
		return "", fmt.Errorf("GitHub release page HTTP %d (expected release redirect)", resp.StatusCode)
	}
	location, err := resp.Location()
	if err != nil {
		return "", fmt.Errorf("GitHub release page returned an invalid redirect")
	}
	prefix := "/" + panelRepository + "/releases/tag/"
	if location.Scheme != "https" || location.Host != "github.com" || location.User != nil ||
		location.RawQuery != "" || location.Fragment != "" || !strings.HasPrefix(location.Path, prefix) {
		return "", fmt.Errorf("GitHub release page returned an unexpected redirect")
	}
	tag := strings.TrimPrefix(location.Path, prefix)
	if !stablePanelTag.MatchString(tag) {
		return "", fmt.Errorf("GitHub release page returned an invalid stable release tag")
	}
	return tag, nil
}

// Network errors can embed proxy credentials in URLs. Return a useful category
// to the browser without exposing raw transport errors or response bodies.
func releaseNetworkError(err error) string {
	var dnsErr *net.DNSError
	var tlsErr *tls.CertificateVerificationError
	var netErr net.Error
	switch {
	case errors.As(err, &dnsErr):
		return "DNS lookup failed"
	case errors.As(err, &tlsErr):
		return "TLS certificate verification failed"
	case errors.As(err, &netErr) && netErr.Timeout():
		return "request timed out"
	default:
		return "network request failed (check server connectivity and panel outbound proxy)"
	}
}
