package urlutil

import (
	"fmt"
	"net/netip"
	"net/url"
	"strings"
)

// ResolvePublicHTTPSURL returns raw or defaultURL after validating it is a public HTTPS URL.
func ResolvePublicHTTPSURL(raw, defaultURL, label string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = defaultURL
	}

	parsed, err := parseAndValidatePublicHTTPSURL(value, label)
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}

// ValidatePublicHTTPSURL checks that raw is a public HTTPS URL without embedded credentials.
func ValidatePublicHTTPSURL(raw, label string) error {
	_, err := parseAndValidatePublicHTTPSURL(raw, label)
	return err
}

func parseAndValidatePublicHTTPSURL(raw, label string) (*url.URL, error) {
	value := strings.TrimSpace(raw)
	parsed, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("parse %s %q: %w", labelName(label), value, err)
	}
	if parsed.Scheme != "https" {
		return nil, fmt.Errorf("%s must use https scheme", labelName(label))
	}
	if parsed.Hostname() == "" {
		return nil, fmt.Errorf("%s host is empty", labelName(label))
	}
	if parsed.User != nil {
		return nil, fmt.Errorf("%s must not include userinfo", labelName(label))
	}
	if IsPrivateIPLiteral(parsed.Hostname()) {
		return nil, fmt.Errorf("%s must not use private ip literal", labelName(label))
	}
	return parsed, nil
}

// IsPrivateIPLiteral reports whether host is an IP literal from a non-public range.
func IsPrivateIPLiteral(host string) bool {
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	return addr.IsPrivate() ||
		addr.IsLoopback() ||
		addr.IsLinkLocalUnicast() ||
		addr.IsLinkLocalMulticast() ||
		addr.IsMulticast() ||
		addr.IsUnspecified()
}

func labelName(label string) string {
	if strings.TrimSpace(label) == "" {
		return "url"
	}
	return label
}
